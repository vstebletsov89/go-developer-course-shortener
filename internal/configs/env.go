// Package configs provides primitives for settings of service.
package configs

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
	"reflect"
	"sync"

	"github.com/caarlos0/env/v6"
)

// Config contains global settings of service.
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080" json:"server_address"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080" json:"base_url"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"" json:"file_storage_path"`
	DatabaseDsn     string `env:"DATABASE_DSN" envDefault:"" json:"database_dsn"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS" envDefault:"false" json:"enable_https"`
	Config          string `env:"CONFIG" envDefault:""`
	TrustedSubnet   string `env:"TRUSTED_SUBNET" envDefault:"" json:"trusted_subnet"`
}

var once sync.Once

func (c *Config) readCommandLineArgs() {
	once.Do(func() {
		flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "server and port to listen on")
		flag.StringVar(&c.BaseURL, "b", c.BaseURL, "base url of the resulting shorthand")
		flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "file storage path")
		flag.StringVar(&c.DatabaseDsn, "d", c.DatabaseDsn, "database dsn")
		flag.BoolVar(&c.EnableHTTPS, "s", c.EnableHTTPS, "enable https mode")
		flag.StringVar(&c.Config, "c", c.Config, "json config path")
		flag.StringVar(&c.TrustedSubnet, "t", c.TrustedSubnet, "enable trusted subnet mode")
		flag.Parse()
	})
}

// ReadConfig merges settings from environment and command line arguments.
func ReadConfig() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)

	if err != nil {
		return nil, err
	}
	cfg.readCommandLineArgs()

	// read json config
	if cfg.Config != "" {
		data, err := os.ReadFile(cfg.Config)
		if err != nil {
			return nil, err
		}

		var fileConfig Config
		err = json.Unmarshal(data, &fileConfig)
		if err != nil {
			return nil, err
		}

		if reflect.ValueOf(fileConfig).IsZero() {
			return nil, errors.New("empty config file")
		}

		// config file has low priority
		// overwrite only default or empty options
		if cfg.ServerAddress == "localhost:8080" && fileConfig.ServerAddress != "" {
			cfg.ServerAddress = fileConfig.ServerAddress
		}
		if cfg.BaseURL == "http://localhost:8080" && fileConfig.BaseURL != "" {
			cfg.BaseURL = fileConfig.BaseURL
		}
		if cfg.FileStoragePath == "" && fileConfig.FileStoragePath != "" {
			cfg.FileStoragePath = fileConfig.FileStoragePath
		}
		if cfg.DatabaseDsn == "" && fileConfig.DatabaseDsn != "" {
			cfg.DatabaseDsn = fileConfig.DatabaseDsn
		}
		if cfg.TrustedSubnet == "" && fileConfig.TrustedSubnet != "" {
			cfg.TrustedSubnet = fileConfig.TrustedSubnet
		}
		if !cfg.EnableHTTPS && fileConfig.EnableHTTPS {
			cfg.EnableHTTPS = fileConfig.EnableHTTPS
		}
	}

	log.Printf("%+v\n\n", cfg)
	return &cfg, nil
}
