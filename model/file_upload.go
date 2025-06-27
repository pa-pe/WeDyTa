package model

type UploadCheckRequest struct {
	Field string `json:"field"`
	Model string `json:"model"`
	ID    string `json:"id"`
}

type UploadCheckResponse struct {
	Allowed bool   `json:"allowed"`
	Message string `json:"message"`
}
