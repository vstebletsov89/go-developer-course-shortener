package configs

import (
	"github.com/caarlos0/env/v6"
	"log"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL       string `env:"BASE_URL" envDefault:"http://localhost:8080"`
}

var EnvConfig Config

func InitConfiguration() {
	err := env.Parse(&EnvConfig)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v\n\n", EnvConfig)
}
