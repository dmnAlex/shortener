package config

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
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

type Config struct {
	LaunchAddress  Address
	ShortenAddress string
}

func New() *Config {
	cfg := &Config{
		LaunchAddress: Address{Host: defaultHost, Port: defaultPort},
	}
	flag.Var(&cfg.LaunchAddress, "a", "launch address")
	flag.StringVar(&cfg.ShortenAddress, "b", fmt.Sprintf("http://%s:%d", defaultHost, defaultPort), "shorten address")

	flag.Parse()

	return cfg
}
