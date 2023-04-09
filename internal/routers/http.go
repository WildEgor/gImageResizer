package routers

import (
	"fmt"

	handlers "github.com/WildEgor/gImageResizer/internal/handlers/http"
	"github.com/gofiber/fiber/v2"
	swagger "github.com/gofiber/swagger"
)

type HTTPRouter struct {
	saveFilesHandler    *handlers.SaveFilesHandler
	downloadFileHandler *handlers.DownloadFileHandler
}

func NewHTTPRouter(
	saveFilesHandler *handlers.SaveFilesHandler,
	downloadFileHandler *handlers.DownloadFileHandler,
) *HTTPRouter {
	return &HTTPRouter{
		saveFilesHandler:    saveFilesHandler,
		downloadFileHandler: downloadFileHandler,
	}
}

func (r *HTTPRouter) SetupRoutes(app *fiber.App) error {

	v1 := app.Group("/api/v1")

	upload := v1.Group("/upload")

	upload.Post("/", r.saveFilesHandler.Handle)
	upload.Get("/:key", r.downloadFileHandler.Handle)

	return nil
}

func (r *HTTPRouter) SwaggerRoute(a *fiber.App, host string) {
	// Create routes group.
	route := a.Group("/swagger")

	// Routes for GET method:
	route.Get("*", swagger.New(swagger.Config{ // custom
		URL:         fmt.Sprintf("http://%v/docs/swagger.json", host),
		DeepLinking: false,
		// Expand ("list") or Collapse ("none") tag groups by default
		DocExpansion: "none",
		// Prefill OAuth ClientId on Authorize popup
		// OAuth: &swagger.OAuthConfig{
		// 	AppName:  "OAuth Provider",
		// 	ClientId: "21bb4edc-05a7-4afc-86f1-2e151e4ba6e2",
		// },
		// Ability to change OAuth2 redirect uri location
		// OAuth2RedirectUrl: "http://localhost:8080/swagger/oauth2-redirect.html",
	}))
}
