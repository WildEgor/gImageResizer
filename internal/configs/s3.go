package configs

import (
	"github.com/caarlos0/env/v7"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type S3Config struct {
	Region    string `env:"S3_REGION"`
	Bucket    string `env:"S3_DEFAUL_BUCKET"`
	Endpoint  string `env:"S3_ENDPOINT"`
	AccessKey string `env:"S3_AKEY"`
	SecretKey string `env:"S3_SKEY"`
	UseSSL    bool   `env:"S3_USE_SSL"`
}

func NewS3Config() *S3Config {
	cfg := S3Config{}

	if err := godotenv.Load(".env", ".env.local"); err == nil {
		if err := env.Parse(&cfg); err != nil {
			log.Printf("%+v\n", err)
		}

		if cfg.Region == "" {
			cfg.Region = "eu-east-1"
		}

		if cfg.Bucket == "" {
			cfg.Bucket = "test"
		}
	}

	return &cfg
}
