package adapters

import (
	"github.com/WildEgor/gImageResizer/internal/configs"
)

type IS3Adapter interface {
}

type S3Adapter struct {
	config *configs.S3Config
}

func NewSMSAdapter(
	config *configs.S3Config,
) *S3Adapter {
	return &S3Adapter{
		config: config,
	}
}
