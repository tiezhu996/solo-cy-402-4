package model

import "time"

// Case 案件实体。
type Case struct {
	ID           uint64       `gorm:"primaryKey" json:"id"`
	CaseNo       string       `gorm:"size:50;uniqueIndex;not null" json:"case_no"`
	Title        string       `gorm:"size:200;not null" json:"title"`
	CaseType     string       `gorm:"size:30;not null;default:civil" json:"case_type"`
	Status       string       `gorm:"size:30;not null;default:filed;index" json:"status"`
	AcceptDate   *time.Time   `json:"accept_date"`
	CloseDate    *time.Time   `json:"close_date"`
	Summary      string       `gorm:"type:text" json:"summary"`
	ClientID     uint64       `gorm:"not null;index" json:"client_id"`
	LeadLawyerID uint64       `gorm:"not null;index" json:"lead_lawyer_id"`
	CoLawyerIDs  CoLawyerJSON `gorm:"type:jsonb" json:"co_lawyer_ids"`
	CreatedAt    time.Time    `json:"created_at"`
}

// TableName 指定表名。
func (Case) TableName() string { return "cases" }

// CoLawyerJSON 律师 ID 列表 JSON 类型。
type CoLawyerJSON []byte
