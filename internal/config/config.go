package config

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

const (
	defautHost  = "localhost"
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
	ShortenAddress Address
}

func New() *Config {
	cfg := &Config{
		LaunchAddress:  Address{Host: defautHost, Port: defaultPort},
		ShortenAddress: Address{Host: defautHost, Port: defaultPort},
	}
	flag.Var(&cfg.LaunchAddress, "a", "launch address")
	flag.Var(&cfg.ShortenAddress, "b", "shorten address")

	flag.Parse()

	return cfg
}

func (c *Config) GetLaunchAddress() string {
	return fmt.Sprintf("%s:%d", c.LaunchAddress.Host, c.LaunchAddress.Port)
}

func (c *Config) GetShortenAddress() string {
	return fmt.Sprintf("%s:%d", c.ShortenAddress.Host, c.ShortenAddress.Port)
}
