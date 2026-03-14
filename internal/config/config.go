package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"dario.cat/mergo"
	"github.com/caarlos0/env"
	"github.com/pkg/errors"
)

type Address struct {
	Host string
	Port int
}

func (a Address) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}

func (a *Address) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("need address in a form host:port")
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	a.Host = hp[0]
	a.Port = port

	return nil
}

func (a *Address) UnmarshalText(text []byte) error {
	return a.Set(string(text))
}

type Config struct {
	LaunchAddress   Address `env:"SERVER_ADDRESS" json:"server_address"`
	ShortenAddress  string  `env:"BASE_URL" json:"base_url"`
	LogLevel        string  `env:"LOG_LEVEL" json:"log_level"`
	FileStoragePath string  `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DatabaseDSN     string  `env:"DATABASE_DSN" json:"database_dsn"`
	MigrationsPath  string  `env:"MIGRATIONS_PATH" json:"migrations_path"`
	JWTSecret       string  `env:"JWT_SECRET" json:"jwt_secret"`
	AuditFile       string  `env:"AUDIT_FILE" json:"audit_file"`
	AuditURL        string  `env:"AUDIT_URL" json:"audit_url"`
	PprofAddress    string  `env:"PPROF_ADDRESS" json:"pprof_address"`
	EnableHTTPS     bool    `env:"ENABLE_HTTPS" json:"enable_https"`
	TrustedSubnet   string  `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
	GRPCAddress     string  `env:"GRPC_ADDRESS" json:"grpc_address"`
}

func New() (*Config, error) {
	cfg := &Config{}
	var configPath string
	flag.StringVar(&configPath, "c", "", "path to config file")
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Var(&cfg.LaunchAddress, "a", "launch address")
	flag.StringVar(&cfg.ShortenAddress, "b", "", "shorten address")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "file storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database dsn")
	flag.StringVar(&cfg.MigrationsPath, "m", "", "migrations path")
	flag.StringVar(&cfg.JWTSecret, "j", "", "JWT secret")
	flag.StringVar(&cfg.AuditFile, "audit-file", "", "path to audit log file")
	flag.StringVar(&cfg.AuditURL, "audit-url", "", "remote audit server URL")
	flag.StringVar(&cfg.PprofAddress, "pprof", "", "pprof address")
	flag.BoolVar(&cfg.EnableHTTPS, "s", false, "enable https")
	flag.StringVar(&cfg.TrustedSubnet, "t", "", "trusted subnet CIDR")
	flag.StringVar(&cfg.GRPCAddress, "g", "", "grpc server address")
	flag.Parse()

	if envPath := os.Getenv("CONFIG"); envPath != "" {
		configPath = envPath
	}

	if configPath != "" {
		jsonCfg, err := loadJSON(configPath)
		if err != nil {
			return nil, errors.Wrap(err, "load json")
		}

		if err := mergo.Merge(cfg, jsonCfg); err != nil {
			return nil, errors.Wrap(err, "merge json configs")
		}
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	setDefaults(cfg)

	return cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.LaunchAddress.Host == "" {
		cfg.LaunchAddress = Address{Host: "localhost", Port: 8080}
	}
	if cfg.ShortenAddress == "" {
		prefix := "http"
		if cfg.EnableHTTPS {
			prefix = "https"
		}
		cfg.ShortenAddress = fmt.Sprintf("%s://%s", prefix, cfg.LaunchAddress.String())
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = "./storage.json"
	}
	if cfg.MigrationsPath == "" {
		cfg.MigrationsPath = "./migrations"
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "defaultsecret"
	}
}

func loadJSON(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "open file")
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, errors.Wrap(err, "decode json")
	}

	return &cfg, nil
}
