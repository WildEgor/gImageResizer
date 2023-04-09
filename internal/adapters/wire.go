package adapters

import (
	"github.com/google/wire"
)

var AdaptersSet = wire.NewSet(
	NewS3Adapter,
	wire.Bind(new(IS3Adapter), new(*S3Adapter)),
)
