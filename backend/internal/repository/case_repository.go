package repository

import (
	"errors"
	"fmt"
	"time"

	"cylawcase/internal/model"

	"gorm.io/gorm"
)

// CaseRepository 案件仓储。
type CaseRepository struct {
	db *gorm.DB
}

// NewCaseRepository 构造案件仓储。
func NewCaseRepository(db *gorm.DB) *CaseRepository {
	return &CaseRepository{db: db}
}

// Create 创建案件。
func (r *CaseRepository) Create(c *model.Case) error {
	if err := r.db.Create(c).Error; err != nil {
		return fmt.Errorf("create case: %w", err)
	}
	return nil
}

// FindByID 按 ID 查询案件。
func (r *CaseRepository) FindByID(id uint64) (*model.Case, error) {
	var c model.Case
	if err := r.db.First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find case by id: %w", err)
	}
	return &c, nil
}

// List 分页查询案件，支持类型/状态/律师/时间范围筛选。
func (r *CaseRepository) List(page, pageSize int, caseType, status string, lawyerID uint64, startDate, endDate *time.Time) ([]model.Case, int64, error) {
	var list []model.Case
	var total int64
	q := r.db.Model(&model.Case{})
	if caseType != "" {
		q = q.Where("case_type = ?", caseType)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if lawyerID > 0 {
		q = q.Where("lead_lawyer_id = ?", lawyerID)
	}
	if startDate != nil {
		q = q.Where("accept_date >= ?", startDate)
	}
	if endDate != nil {
		q = q.Where("accept_date <= ?", endDate)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count cases: %w", err)
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("list cases: %w", err)
	}
	return list, total, nil
}

// ListByClient 查询某客户的案件。
func (r *CaseRepository) ListByClient(clientID uint64) ([]model.Case, error) {
	var list []model.Case
	if err := r.db.Where("client_id = ?", clientID).Order("id DESC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("list cases by client: %w", err)
	}
	return list, nil
}

// ListByLawyer 查询某律师主办的案件。
func (r *CaseRepository) ListByLawyer(lawyerID uint64) ([]model.Case, error) {
	var list []model.Case
	if err := r.db.Where("lead_lawyer_id = ?", lawyerID).Order("id DESC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("list cases by lawyer: %w", err)
	}
	return list, nil
}

// Update 更新案件。
func (r *CaseRepository) Update(c *model.Case) error {
	if err := r.db.Save(c).Error; err != nil {
		return fmt.Errorf("update case: %w", err)
	}
	return nil
}

// Count 统计案件总数。
func (r *CaseRepository) Count() (int64, error) {
	var n int64
	if err := r.db.Model(&model.Case{}).Count(&n).Error; err != nil {
		return 0, fmt.Errorf("count cases: %w", err)
	}
	return n, nil
}
