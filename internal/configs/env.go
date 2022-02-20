package configs

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"FILE_STORAGE_PATH_NOT_DEFINED"`
}

func flagExists(name string) bool {
	if flag.Lookup(name) == nil {
		return false
	}
	return true
}

func (c *Config) readCommandLineArgs() {
	if !flagExists("a") {
		flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "server and port to listen on")
	}
	if !flagExists("b") {
		flag.StringVar(&c.BaseURL, "b", c.BaseURL, "base url of the resulting shorthand")
	}
	if !flagExists("f") {
		flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "file storage path")
	}
	flag.Parse()
}

func ReadConfig() *Config {
	var cfg Config
	err := env.Parse(&cfg)

	if err != nil {
		log.Fatal(err)
	}
	cfg.readCommandLineArgs()
	log.Printf("%+v\n\n", cfg)
	return &cfg
}
