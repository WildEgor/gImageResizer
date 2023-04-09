package routers

import (
	handlers "github.com/WildEgor/gImageResizer/internal/handlers"
	"github.com/google/wire"
)

var RoutersSet = wire.NewSet(
	handlers.HandlersSet,
	NewHTTPRouter,
)
