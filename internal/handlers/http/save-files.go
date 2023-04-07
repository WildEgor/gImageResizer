package handlers

import (
	"context"
	"io/ioutil"
	"mime/multipart"
	"sync"

	"github.com/WildEgor/gImageResizer/internal/adapters"
	"github.com/gofiber/fiber/v2"
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

	var wg sync.WaitGroup
	wg.Add(len(files))
	readCh := make(chan []*PutObjResult, len(files))

	for _, file := range files[1:] {
		go func(ctx context.Context, file *multipart.FileHeader, resp chan []*PutObjResult) {
			defer func() {
				wg.Done()
			}()

			s := []*PutObjResult{}
			result := &PutObjResult{}

			buffer, err := file.Open()
			if err != nil {
				result.Error = &err
				result.Status = false
			}
			defer buffer.Close()

			buf, err := ioutil.ReadAll(buffer)
			if err != nil {
				result.Error = &err
				result.Status = false
			}

			err = h.s3Adapter.PutObj(ctx, &adapters.S3Obj{
				Key:         file.Filename,
				Size:        file.Size,
				ContentType: file.Header["Content-Type"][0],
				Bytes:       buf,
			})
			if err != nil {
				result.Error = &err
				result.Status = false
			}

			s = append(s, result)

			readCh <- s
		}(ctx.Context(), file, readCh)
	}
	wg.Wait()

	putResults := <-readCh
	putResult := &PutObjResult{}

	for _, r := range putResults {
		if r.Status != true {
			putResult.Error = r.Error
			putResult.Status = r.Status
			putResult.FileName = r.FileName
		}
	}

	if putResult.Status != true {
		ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"isOk": false,
			"data": fiber.Map{
				"message": "ERR_UPLOAD_FILE",
			},
		})
	}

	ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"isOk": true,
		"data": putResult,
	})

	return nil
}
