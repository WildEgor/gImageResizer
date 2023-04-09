package handlers

import (
	log "github.com/sirupsen/logrus"

	"github.com/WildEgor/gImageResizer/internal/adapters"
	"github.com/WildEgor/gImageResizer/internal/configs"
	"github.com/WildEgor/gImageResizer/internal/dtos"
	"github.com/gofiber/fiber/v2"
)

type DownloadFileHandler struct {
	imgProxyConfig *configs.ImgProxyConfig
	appConfig      *configs.AppConfig
	s3Adapter      adapters.IS3Adapter
}

func NewDownloadFileHandler(
	imgProxyConfig *configs.ImgProxyConfig,
	appConfig *configs.AppConfig,
	s3Adapter adapters.IS3Adapter,
) *DownloadFileHandler {
	return &DownloadFileHandler{
		imgProxyConfig: imgProxyConfig,
		appConfig:      appConfig,
		s3Adapter:      s3Adapter,
	}
}

var Sizes map[string]string = map[string]string{
	"_small":  "_small",
	"_medium": "_medium",
	"_blurry": "_blurry",
	"default": "_medium",
}

// TODO: generate sign for private bucket by key and create link
// to image proxy, check extension if image -> proxy, else direct redirect
// Example: GET https://yourdomain.com/api/v1/upload/{uuid.path = key}?size=200

// DownloadFiles godoc
//
//	@Summary		Get file
//	@Description	Get file
//	@Tags			upload
//	@Produce		json
//	@Router			/api/v1/upload [get]
func (h *DownloadFileHandler) Handle(ctx *fiber.Ctx) error {
	log.WithContext(ctx.Context())

	// link, err := h.s3Adapter.GetPresign(ctx.Context(), &adapters.S3Obj{
	// 	Key: key,
	// })
	// if err != nil {
	// 	log.Error(err)
	// 	ctx.Status(fiber.StatusInternalServerError).JSON(dtos.ErrResponse("ERR_PRESIGN"))
	// }

	key := ctx.Params("key", "")
	if key == "" {
		ctx.Status(fiber.StatusBadRequest).JSON(dtos.ErrResponse("ERR_EMPTY_KEY"))
	}

	query, err := h.parseQuery(ctx)
	if err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(dtos.ErrResponse("ERR_QUERY"))
	}

	URL := h.buildURL(key, query)

	ctx.Redirect(h.imgProxyConfig.BaseURL + "/" + URL)

	return nil
}

func (h *DownloadFileHandler) buildURL(key string, params *dtos.DownloadFileQuery) string {
	region := "@bucket" // TODO: use region here

	size := Sizes[params.Size]

	if size == "" {
		size = Sizes["default"]
	}

	return region + "/" + size + "/" + key
}

func (h *DownloadFileHandler) parseQuery(ctx *fiber.Ctx) (*dtos.DownloadFileQuery, error) {
	var queryParams dtos.DownloadFileQuery
	if err := ctx.QueryParser(&queryParams); err != nil {
		return nil, err
	}

	return &queryParams, nil
}
