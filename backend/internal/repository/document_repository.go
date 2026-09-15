package repository

import (
	"errors"
	"fmt"

	"cylawcase/internal/model"

	"gorm.io/gorm"
)

// DocumentRepository 文档仓储。
type DocumentRepository struct {
	db *gorm.DB
}

// NewDocumentRepository 构造文档仓储。
func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// Create 创建文档。
func (r *DocumentRepository) Create(d *model.Document) error {
	if err := r.db.Create(d).Error; err != nil {
		return fmt.Errorf("create document: %w", err)
	}
	return nil
}

// FindByID 按 ID 查询文档。
func (r *DocumentRepository) FindByID(id uint64) (*model.Document, error) {
	var d model.Document
	if err := r.db.First(&d, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find document by id: %w", err)
	}
	return &d, nil
}

// ListByCase 查询某案件文档。
func (r *DocumentRepository) ListByCase(caseID uint64) ([]model.Document, error) {
	var list []model.Document
	if err := r.db.Where("case_id = ?", caseID).Order("upload_time DESC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("list documents by case: %w", err)
	}
	return list, nil
}

// List 分页查询文档，支持类型/关键词筛选。
func (r *DocumentRepository) List(page, pageSize int, fileType, keyword string) ([]model.Document, int64, error) {
	var list []model.Document
	var total int64
	q := r.db.Model(&model.Document{})
	if fileType != "" {
		q = q.Where("file_type = ?", fileType)
	}
	if keyword != "" {
		q = q.Where("title LIKE ?", "%"+keyword+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count documents: %w", err)
	}
	if err := q.Order("upload_time DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("list documents: %w", err)
	}
	return list, total, nil
}

// Delete 删除文档。
func (r *DocumentRepository) Delete(id uint64) error {
	if err := r.db.Delete(&model.Document{}, id).Error; err != nil {
		return fmt.Errorf("delete document: %w", err)
	}
	return nil
}
