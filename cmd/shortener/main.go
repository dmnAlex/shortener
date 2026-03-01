package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dmnAlex/shortener/internal/audit"
	"github.com/dmnAlex/shortener/internal/config"
	"github.com/dmnAlex/shortener/internal/handler"
	"github.com/dmnAlex/shortener/internal/logger"
	"github.com/dmnAlex/shortener/internal/repository"
	"github.com/dmnAlex/shortener/internal/service"
	"github.com/dmnAlex/shortener/internal/storage/pg"
	"github.com/dmnAlex/shortener/internal/tls"
	"go.uber.org/zap"

	"net/http"
	_ "net/http/pprof"
)

// go build -ldflags "-X main.buildVersion=v1.0.1 -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')' -X 'main.buildCommit=$(git rev-parse HEAD)'" ./cmd/shortener
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	logInfo()

	globalCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	if err = logger.Init(cfg.LogLevel); err != nil {
		log.Fatalf("init logger error: %v", err)
	}

	var repo repository.URLRepository
	if cfg.DatabaseDSN != "" {
		logger.Log.Info("using pg database")
		var db *pg.DB
		db, err = pg.New(globalCtx, cfg.DatabaseDSN, cfg.MigrationsPath)
		if err != nil {
			log.Fatalf("db error: %v", err)
		}

		repo = repository.NewPostgresRepo(db)

	} else {
		logger.Log.Info("using file database")
		repo, err = repository.NewFileRepo(cfg.FileStoragePath)
		if err != nil {
			log.Fatalf("file repo error: %v", err)
		}
	}
	defer func() {
		if err := repo.Close(); err != nil {
			logger.Log.Error("repo close error", zap.Error(err))
		}
	}()

	auditMgr := audit.NewAuditManager()
	defer auditMgr.Close()

	if cfg.AuditFile != "" {
		fileAuditor, err := audit.NewFileAuditor(cfg.AuditFile)
		if err != nil {
			log.Fatalf("init file auditor: %v", err)
		}

		auditMgr.Subscribe(fileAuditor)
	}

	if cfg.AuditURL != "" {
		auditMgr.Subscribe(audit.NewRemoteAuditor(cfg.AuditURL))
	}

	service := service.NewURLService(repo)
	handler := handler.NewShortenerHandler(service, cfg, auditMgr)
	router := newRouter(handler, cfg)

	srv := &http.Server{
		Addr:    cfg.LaunchAddress.String(),
		Handler: router,
	}

	go func() {
		var err error
		if cfg.EnableHTTPS {
			certFile, keyFile, genErr := tls.GenerateSelfSignedCert()
			if genErr != nil {
				log.Fatalf("generate cert: %v", genErr)
			}
			defer os.Remove(certFile)
			defer os.Remove(keyFile)
			err = srv.ListenAndServeTLS(certFile, keyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen:  %v", err)
		}
	}()

	var pprofSrv *http.Server
	if cfg.PprofAddress != "" {
		pprofSrv = &http.Server{
			Addr:    cfg.PprofAddress,
			Handler: nil,
		}
		go func() {
			if err := pprofSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Log.Error("pprof server error", zap.Error(err))
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	logger.Log.Info("Shutdown server...")

	ctx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Server shutdown:", zap.Error(err))
	}

	if pprofSrv != nil {
		if err := pprofSrv.Shutdown(ctx); err != nil {
			logger.Log.Error("Pprof shutdown:", zap.Error(err))
		}
	}
}

func logInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	log.Println("Build version: ", buildVersion)
	log.Println("Build date: ", buildDate)
	log.Println("Build commit: ", buildCommit)
}
