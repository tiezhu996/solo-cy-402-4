package model

import "time"

// Billing 费用账单实体。
type Billing struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	BillNo      string    `gorm:"size:50;uniqueIndex;not null" json:"bill_no"`
	BillingType string    `gorm:"size:30;not null;default:attorney_fee" json:"billing_type"`
	Amount      float64   `gorm:"type:numeric(12,2);not null;default:0" json:"amount"`
	Status      string    `gorm:"size:30;not null;default:pending;index" json:"status"`
	CaseID      uint64    `gorm:"not null;index" json:"case_id"`
	ClientID    uint64    `gorm:"not null;index" json:"client_id"`
	InvoiceInfo string    `gorm:"size:255;not null;default:''" json:"invoice_info"`
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 指定表名。
func (Billing) TableName() string { return "billings" }
