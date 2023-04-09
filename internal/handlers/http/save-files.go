package handlers

import (
	"io/ioutil"
	"mime/multipart"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/WildEgor/gImageResizer/internal/adapters"
	"github.com/gofiber/fiber/v2"
	uuid "github.com/google/uuid"
)

type PutObjResult struct {
	FileName string
	Status   bool
	Error    *error
}

type SaveFilesHandler struct {
	s3Adapter adapters.IS3Adapter
}

func NewSaveFilesHandler(
	s3Adapter adapters.IS3Adapter,
) *SaveFilesHandler {
	return &SaveFilesHandler{
		s3Adapter: s3Adapter,
	}
}

// SaveFiles godoc
//
//	@Summary		Upload any valid files
//	@Description	Upload files
//	@Tags			upload
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			files	file
//	@Router			/api/v1/upload [post]
func (h *SaveFilesHandler) Handle(ctx *fiber.Ctx) error {
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"isOk": false,
			"data": fiber.Map{
				"message": "ERR_MULTIPART",
			},
		})
	}

	files := form.File["files"]
	if len(files) == 0 {
		ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"isOk": false,
			"data": fiber.Map{
				"message": "ERR_EMPTY_FILES",
			},
		})
	}

	successPaths := make([]string, len(files))
	pathNumber := 0
	for _, formFile := range files {

		binaryFile, err := h.readFile(formFile)
		if err != nil {
			ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"isOk": false,
				"data": fiber.Map{
					"message": "ERR_READ_FILE",
				},
			})
			break
		}

		key := uuid.New().String() + "-" + formFile.Filename
		contentType := http.DetectContentType(binaryFile)

		result, uerr := h.s3Adapter.SessionUpload(ctx.Context(), &adapters.S3Obj{
			Key:           key,
			Bytes:         binaryFile,
			ContentType:   contentType,
			ContentLength: int64(len(binaryFile)),
		})
		if uerr != nil {
			log.Error(uerr)
			continue
		}

		successPaths[pathNumber] = *result
		pathNumber++
	}

	ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"isOk": true,
		"data": fiber.Map{
			"message": "SUCCESS",
			"paths":   successPaths,
		},
	})

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
