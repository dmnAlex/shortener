package config

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/caarlos0/env"
)

const (
	defaultHost = "localhost"
	defaultPort = 8080
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
	LaunchAddress   Address `env:"SERVER_ADDRESS"`
	ShortenAddress  string  `env:"BASE_URL"`
	LogLevel        string  `env:"LOG_LEVEL"`
	FileStoragePath string  `env:"FILE_STORAGE_PATH"`
}

func New() (*Config, error) {
	cfg := &Config{
		LaunchAddress: Address{Host: defaultHost, Port: defaultPort},
	}
	flag.Var(&cfg.LaunchAddress, "a", "launch address")
	flag.StringVar(&cfg.ShortenAddress, "b", fmt.Sprintf("http://%s:%d", defaultHost, defaultPort), "shorten address")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.FileStoragePath, "f", "./storage.json", "file storage path")

	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
