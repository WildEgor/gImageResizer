package dtos

type DownloadFileQuery struct {
	Size string `query:"size" default:"_small"`
}

type DownloadFileResponse struct {
	Url string `json:"url"`
}
