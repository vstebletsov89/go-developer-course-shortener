// Package configs provides primitives for settings of service.
package configs

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

// Config contains global settings of service.
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:""`
	DatabaseDsn     string `env:"DATABASE_DSN" envDefault:""`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS" envDefault:"false"`
}

func (c *Config) readCommandLineArgs() {
	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "server and port to listen on")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "base url of the resulting shorthand")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "file storage path")
	flag.StringVar(&c.DatabaseDsn, "d", c.DatabaseDsn, "database dsn")
	flag.BoolVar(&c.EnableHTTPS, "s", c.EnableHTTPS, "enable https mode")
	flag.Parse()
}

// ReadConfig merges settings from environment and command line arguments.
func ReadConfig() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)

	if err != nil {
		return nil, err
	}
	cfg.readCommandLineArgs()
	log.Printf("%+v\n\n", cfg)
	return &cfg, nil
}
