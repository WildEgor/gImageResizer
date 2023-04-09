package configs

import (
	"github.com/caarlos0/env/v7"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type ImgProxyConfig struct {
	BaseURL string `env:"IMG_PROXY_BASE_URL"`
}

func NewImgProxyConfig() *ImgProxyConfig {
	cfg := ImgProxyConfig{}

	if err := godotenv.Load(".env", ".env.local"); err == nil {
		if err := env.Parse(&cfg); err != nil {
			log.Printf("%+v\n", err)
		}
	}

	return &cfg
}
