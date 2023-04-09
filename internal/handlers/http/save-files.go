package handlers

import (
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/WildEgor/gImageResizer/internal/adapters"
	"github.com/WildEgor/gImageResizer/internal/configs"
	dtos "github.com/WildEgor/gImageResizer/internal/dtos"
	"github.com/gofiber/fiber/v2"
	uuid "github.com/google/uuid"
)

type PutObjResult struct {
	FileName string
	Status   bool
	Error    *error
}

type SaveFilesHandler struct {
	appConfig *configs.AppConfig
	s3Adapter adapters.IS3Adapter
}

func NewSaveFilesHandler(
	appConfig *configs.AppConfig,
	s3Adapter adapters.IS3Adapter,
) *SaveFilesHandler {
	return &SaveFilesHandler{
		appConfig: appConfig,
		s3Adapter: s3Adapter,
	}
}

// SaveFiles godoc
//
//		@Summary		Upload any valid files
//		@Description	Upload files
//		@Tags			upload
//		@Accept			multipart/form-data
//		@Produce		json
//	 @Param files formData file true "Files"
//	 @Param request formData object true "Request Body"
//		@Router			/api/v1/upload [post]
func (h *SaveFilesHandler) Handle(ctx *fiber.Ctx) error {
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(dtos.ErrResponse("ERR_MULTIPART"))
	}

	files := form.File["files"]
	if len(files) == 0 {
		ctx.Status(fiber.StatusBadRequest).JSON(dtos.ErrResponse("ERR_EMPTY_FILES"))
	}

	wg := sync.WaitGroup{}

	successPaths := make([]dtos.UploadFilesResponse, len(files))
	pathNumber := 0
	for _, formFile := range files {
		wg.Add(1)

		binaryFile, err := h.readFile(formFile)
		if err != nil {
			ctx.Status(fiber.StatusInternalServerError).JSON(dtos.ErrResponse("ERR_READ_FILE"))
			break
		}

		key := uuid.New().String() + "-" + formFile.Filename
		contentType := http.DetectContentType(binaryFile)

		go func(filename string) {
			defer wg.Done()
			_, uerr := h.s3Adapter.SessionUpload(ctx.Context(), &adapters.S3Obj{
				Key:           key,
				Bytes:         binaryFile,
				ContentType:   contentType,
				ContentLength: int64(len(binaryFile)),
			})
			if uerr == nil {
				successPaths[pathNumber] = dtos.UploadFilesResponse{
					Name:       filename,
					Url:        h.appConfig.BaseURL + "/" + key,
					UploadedAt: time.Now(),
				}
			}

			pathNumber++
		}(formFile.Filename)
	}

	wg.Wait()

	ctx.Status(fiber.StatusOK).JSON(dtos.SuccessResponse(successPaths))

	return nil
}

func (h *SaveFilesHandler) readFile(file *multipart.FileHeader) ([]byte, error) {
	openedFile, _ := file.Open()

	binaryFile, err := ioutil.ReadAll(openedFile)
	if err != nil {
		return nil, err
	}

	defer func(openedFile multipart.File) {
		err := openedFile.Close()
		if err != nil {
			log.Fatalf("Failed closing file %v", file.Filename)
		}
	}(openedFile)

	return binaryFile, nil
}
