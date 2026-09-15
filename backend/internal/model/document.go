package model

import "time"

// Document 文档实体。
type Document struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	Title      string    `gorm:"size:200;not null" json:"title"`
	FileType   string    `gorm:"size:30;not null;default:other" json:"file_type"`
	FileURL    string    `gorm:"size:500;not null;default:''" json:"file_url"`
	UploadTime time.Time `json:"upload_time"`
	CaseID     uint64    `gorm:"not null;index" json:"case_id"`
	UploaderID uint64    `gorm:"not null" json:"uploader_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// TableName 指定表名。
func (Document) TableName() string { return "documents" }
