package configs

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:""`
}

func (c *Config) readCommandLineArgs() {
	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "server and port to listen on")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "base url of the resulting shorthand")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "file storage path")
	flag.Parse()
}

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
