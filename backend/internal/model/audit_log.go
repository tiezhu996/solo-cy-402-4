package model

import "time"

// AuditLog 审计日志实体。
type AuditLog struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	OperatorID   uint64    `gorm:"not null;default:0" json:"operator_id"`
	OperatorName string    `gorm:"size:50;not null;default:''" json:"operator_name"`
	Action       string    `gorm:"size:50;not null" json:"action"`
	EntityType   string    `gorm:"size:50;not null" json:"entity_type"`
	EntityID     string    `gorm:"size:50;not null;default:''" json:"entity_id"`
	Detail       string    `gorm:"type:text" json:"detail"`
	IP           string    `gorm:"size:50;not null;default:''" json:"ip"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName 指定表名。
func (AuditLog) TableName() string { return "audit_logs" }
