package server

type ErrorResponse struct {
	Error string `json:"error"`
}

type UploadResponse struct {
	FID string `json:"fid"`
}
