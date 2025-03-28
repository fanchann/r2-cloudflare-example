package dto

type FileSuccessUploadResponse struct {
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size_in_kb"`
}

type FileListsResponse struct {
	Files []R2ListFile `json:"files"`
}

type R2ListFile struct {
	FileName string `json:"file_name"`
	Size     int64  `json:"file_size_in_kb"`
}

type MakeFilePublicResponse struct {
	URL      string `json:"url"`
	Duration string `json:"duration"`
}

type WebAPIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
