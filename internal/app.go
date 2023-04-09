package app

import (
	"fmt"

	"github.com/WildEgor/gImageResizer/internal/adapters"
	"github.com/WildEgor/gImageResizer/internal/configs"
	handlers_http "github.com/WildEgor/gImageResizer/internal/handlers/http"
	"github.com/WildEgor/gImageResizer/internal/routers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/wire"
	log "github.com/sirupsen/logrus"
)

var AppSet = wire.NewSet(
	NewApp,
	adapters.AdaptersSet,
	configs.ConfigsSet,
	routers.RoutersSet,
)

func NewApp(
	appConfig *configs.AppConfig,
	httpRouter *routers.HTTPRouter,
) *fiber.App {
	app := fiber.New(fiber.Config{
		EnablePrintRoutes: true,
		ErrorHandler:      handlers_http.ErrorHandler,
	})

	app.Use(cors.New(cors.Config{
		AllowHeaders:     "Origin, Content-Type, Accept, Content-Length, Accept-Language, Accept-Encoding, Connection, Access-Control-Allow-Origin",
		AllowOrigins:     "*",
		AllowCredentials: true,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))
	app.Use(recover.New())

	if !appConfig.IsProduction() {
		httpRouter.SwaggerRoute(app, "localhost:8888/docs")
		log.SetLevel(log.DebugLevel)
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}

	httpRouter.SetupRoutes(app)

	log.Info(fmt.Sprintf("Application is running on %v port...", appConfig.Port))
	log.Info(fmt.Sprintf("Swagger served at %v", "http://localhost:8888/swagger"))

	return app
}
