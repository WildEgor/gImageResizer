package dtos

import "github.com/gofiber/fiber/v2"

type GenericResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func SuccessResponse(data interface{}) GenericResponse {
	resp := GenericResponse{}
	resp.Data = data
	resp.Status = true
	return resp
}

func ErrResponse(msg string) GenericResponse {
	resp := GenericResponse{}
	resp.Data = fiber.Map{}
	resp.Message = msg
	return resp
}
