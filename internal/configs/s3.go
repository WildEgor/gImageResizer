package configs

import (
	"github.com/caarlos0/env/v7"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type S3Config struct {
}

func NewS3Config() *S3Config {
	cfg := S3Config{}

	if err := godotenv.Load(".env", ".env.local"); err == nil {
		if err := env.Parse(&cfg); err != nil {
			log.Printf("%+v\n", err)
		}
	}

	return &cfg
}
