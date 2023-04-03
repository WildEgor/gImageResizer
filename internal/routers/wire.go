package routers

import (
	"github.com/google/wire"
)

var RoutersSet = wire.NewSet(
	NewHTTPRouter,
)
