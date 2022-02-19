package configs

import (
	"github.com/caarlos0/env/v6"
	"log"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"FILE_STORAGE_PATH_NOT_DEFINED"`
}

func ReadConfig() *Config {
	var cfg Config
	err := env.Parse(&cfg)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v\n\n", cfg)
	return &cfg
}
