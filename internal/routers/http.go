package routers

import (
	"github.com/gofiber/fiber/v2"
)

type HTTPRouter struct {
}

func NewHTTPRouter() *HTTPRouter {
	return &HTTPRouter{}
}

func (r *HTTPRouter) SetupRoutes(app *fiber.App) error {
	return nil
}
