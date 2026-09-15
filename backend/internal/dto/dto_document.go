package dto

// DocumentCreateRequest 上传文档请求。
type DocumentCreateRequest struct {
	CaseID   uint64 `json:"case_id" binding:"required"`
	Title    string `json:"title" binding:"required,max=200"`
	FileType string `json:"file_type" binding:"required,oneof=complaint defense evidence judgment contract other"`
	FileURL  string `json:"file_url" binding:"required,max=500"`
}
