package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dmnAlex/shortener/internal/audit"
	"github.com/dmnAlex/shortener/internal/config"
	grpcpb "github.com/dmnAlex/shortener/internal/grpc"
	"github.com/dmnAlex/shortener/internal/handler"
	"github.com/dmnAlex/shortener/internal/logger"
	"github.com/dmnAlex/shortener/internal/middleware"
	"github.com/dmnAlex/shortener/internal/repository"
	"github.com/dmnAlex/shortener/internal/router"
	"github.com/dmnAlex/shortener/internal/service"
	"github.com/dmnAlex/shortener/internal/storage/pg"
	"github.com/dmnAlex/shortener/internal/tlsutil"
	pb "github.com/dmnAlex/shortener/proto"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func run() error {
	globalCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.New()
	if err != nil {
		return errors.Wrap(err, "new config")
	}

	if err = logger.Init(cfg.LogLevel); err != nil {
		return errors.Wrap(err, "init logger")
	}

	var repo repository.URLRepository
	if cfg.DatabaseDSN != "" {
		logger.Log.Info("using pg database")
		var db *pg.DB
		db, err = pg.New(globalCtx, cfg.DatabaseDSN, cfg.MigrationsPath)
		if err != nil {
			return errors.Wrap(err, "new pg")
		}

		repo = repository.NewPostgresRepo(db)

	} else {
		logger.Log.Info("using file database")
		repo, err = repository.NewFileRepo(cfg.FileStoragePath)
		if err != nil {
			return errors.Wrap(err, "new file repo")
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
			return errors.Wrap(err, "new file auditor")
		}

		auditMgr.Subscribe(fileAuditor)
	}

	if cfg.AuditURL != "" {
		auditMgr.Subscribe(audit.NewRemoteAuditor(cfg.AuditURL))
	}

	service := service.NewURLService(repo)
	handler := handler.NewShortenerHandler(service, cfg, auditMgr)
	router := router.New(handler, cfg)

	srv := &http.Server{
		Addr:    cfg.LaunchAddress.String(),
		Handler: router,
	}

	var grpcSrv *grpc.Server
	if cfg.GRPCAddress != "" {
		grpcSrv = grpc.NewServer(grpc.UnaryInterceptor(middleware.GRPCAuth(cfg)))
		grpcHandler := grpcpb.NewShortenerServer(service, cfg, auditMgr)
		pb.RegisterShortenerServiceServer(grpcSrv, grpcHandler)

		listen, err := net.Listen("tcp", cfg.GRPCAddress)
		if err != nil {
			return errors.Wrap(err, "grpc listen")
		}
		go func() {
			if err := grpcSrv.Serve(listen); err != nil && err != grpc.ErrServerStopped {
				logger.Log.Error("grpc server error", zap.Error(err))
			}
		}()
	}

	var certFile, keyFile string
	if cfg.EnableHTTPS {
		var genErr error
		certFile, keyFile, genErr = tlsutil.GenerateSelfSignedCert()
		if genErr != nil {
			return errors.Wrap(err, "generate cert")
		}
		defer os.Remove(certFile)
		defer os.Remove(keyFile)
	}

	serverErrors := make(chan error, 1)

	go func() {
		var err error
		if cfg.EnableHTTPS {
			err = srv.ListenAndServeTLS(certFile, keyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			serverErrors <- errors.Wrap(err, "main server listen")
		}
	}()

	var pprofSrv *http.Server
	if cfg.PprofAddress != "" {
		pprofSrv = &http.Server{Addr: cfg.PprofAddress}
		go func() {
			if err := pprofSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Log.Error("pprof server error", zap.Error(err))
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case err := <-serverErrors:
		return err
	case sig := <-quit:
		logger.Log.Info("Shutdown signal received", zap.String("signal", sig.String()))
	}

	ctx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Error("Server shutdown:", zap.Error(err))
	}

	if grpcSrv != nil {
		grpcSrv.GracefulStop()
	}

	if pprofSrv != nil {
		if err := pprofSrv.Shutdown(ctx); err != nil {
			logger.Log.Error("Pprof shutdown:", zap.Error(err))
		}
	}

	return nil
}
