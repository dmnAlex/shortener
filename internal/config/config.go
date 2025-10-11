package config

import (
	"fmt"
	"os"
)

type Config struct {
	Host string
	Port string
}

func New() *Config {
	return &Config{
		Host: getEnv("HOST", "localhost"),
		Port: getEnv("PORT", "8080"),
	}
}

func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
