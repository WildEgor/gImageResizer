package handlers

import (
	http_handlers "github.com/WildEgor/gImageResizer/internal/handlers/http"
	"github.com/google/wire"
)

var HandlersSet = wire.NewSet(
	http_handlers.NewSaveFilesHandler,
	http_handlers.NewDownloadFileHandler,
)
