package repository

import (
	"errors"
	"fmt"
	"time"

	"cylawcase/internal/model"

	"gorm.io/gorm"
)

// BillingRepository 账单仓储。
type BillingRepository struct {
	db *gorm.DB
}

// NewBillingRepository 构造账单仓储。
func NewBillingRepository(db *gorm.DB) *BillingRepository {
	return &BillingRepository{db: db}
}

// Create 创建账单。
func (r *BillingRepository) Create(b *model.Billing) error {
	if err := r.db.Create(b).Error; err != nil {
		return fmt.Errorf("create billing: %w", err)
	}
	return nil
}

// FindByID 按 ID 查询账单。
func (r *BillingRepository) FindByID(id uint64) (*model.Billing, error) {
	var b model.Billing
	if err := r.db.First(&b, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find billing by id: %w", err)
	}
	return &b, nil
}

// List 分页查询账单，支持案件/客户/状态筛选。
func (r *BillingRepository) List(page, pageSize int, caseID, clientID uint64, status string) ([]model.Billing, int64, error) {
	var list []model.Billing
	var total int64
	q := r.db.Model(&model.Billing{})
	if caseID > 0 {
		q = q.Where("case_id = ?", caseID)
	}
	if clientID > 0 {
		q = q.Where("client_id = ?", clientID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count billings: %w", err)
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("list billings: %w", err)
	}
	return list, total, nil
}

// ListByCase 查询某案件账单。
func (r *BillingRepository) ListByCase(caseID uint64) ([]model.Billing, error) {
	var list []model.Billing
	if err := r.db.Where("case_id = ?", caseID).Order("id DESC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("list billings by case: %w", err)
	}
	return list, nil
}

// Update 更新账单。
func (r *BillingRepository) Update(b *model.Billing) error {
	if err := r.db.Save(b).Error; err != nil {
		return fmt.Errorf("update billing: %w", err)
	}
	return nil
}

// Summary 汇总：本月应收、已收、待收。
func (r *BillingRepository) Summary(month time.Time) (map[string]float64, error) {
	start := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)
	var receivables, received, pending float64
	if err := r.db.Model(&model.Billing{}).
		Select("COALESCE(SUM(CASE WHEN status <> 'void' THEN amount ELSE 0 END), 0)").
		Where("created_at >= ? AND created_at < ?", start, end).
		Scan(&receivables).Error; err != nil {
		return nil, fmt.Errorf("summary receivables: %w", err)
	}
	if err := r.db.Model(&model.Billing{}).
		Select("COALESCE(SUM(CASE WHEN status = 'paid' OR status = 'invoiced' THEN amount ELSE 0 END), 0)").
		Where("created_at >= ? AND created_at < ?", start, end).
		Scan(&received).Error; err != nil {
		return nil, fmt.Errorf("summary received: %w", err)
	}
	if err := r.db.Model(&model.Billing{}).
		Select("COALESCE(SUM(CASE WHEN status = 'pending' THEN amount ELSE 0 END), 0)").
		Where("created_at >= ? AND created_at < ?", start, end).
		Scan(&pending).Error; err != nil {
		return nil, fmt.Errorf("summary pending: %w", err)
	}
	return map[string]float64{"receivable": receivables, "received": received, "pending": pending}, nil
}
