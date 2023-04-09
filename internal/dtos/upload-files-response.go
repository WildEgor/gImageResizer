package dtos

import "time"

type UploadFilesResponse struct {
	Name       string    `json:"name"`
	Url        string    `json:"url"`
	UploadedAt time.Time `json:"uploadedAt"`
}
