package repository

import (
	"errors"
	"fmt"

	"cylawcase/internal/model"

	"gorm.io/gorm"
)

// ClientRepository 客户仓储。
type ClientRepository struct {
	db *gorm.DB
}

// NewClientRepository 构造客户仓储。
func NewClientRepository(db *gorm.DB) *ClientRepository {
	return &ClientRepository{db: db}
}

// Create 创建客户。
func (r *ClientRepository) Create(c *model.Client) error {
	if err := r.db.Create(c).Error; err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	return nil
}

// FindByID 按 ID 查询客户。
func (r *ClientRepository) FindByID(id uint64) (*model.Client, error) {
	var c model.Client
	if err := r.db.First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find client by id: %w", err)
	}
	return &c, nil
}

// List 分页查询客户，支持关键词检索。
func (r *ClientRepository) List(page, pageSize int, keyword string) ([]model.Client, int64, error) {
	var list []model.Client
	var total int64
	q := r.db.Model(&model.Client{})
	if keyword != "" {
		q = q.Where("name LIKE ? OR contact LIKE ? OR id_number LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count clients: %w", err)
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("list clients: %w", err)
	}
	return list, total, nil
}

// Update 更新客户。
func (r *ClientRepository) Update(c *model.Client) error {
	if err := r.db.Save(c).Error; err != nil {
		return fmt.Errorf("update client: %w", err)
	}
	return nil
}

// Delete 删除客户。
func (r *ClientRepository) Delete(id uint64) error {
	if err := r.db.Delete(&model.Client{}, id).Error; err != nil {
		return fmt.Errorf("delete client: %w", err)
	}
	return nil
}
