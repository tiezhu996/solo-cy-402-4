package dto

import "time"

// CaseCreateRequest 创建案件请求。
type CaseCreateRequest struct {
	ClientID     uint64   `json:"client_id" binding:"required"`
	LeadLawyerID uint64   `json:"lead_lawyer_id" binding:"required"`
	Title        string   `json:"title" binding:"required,max=200"`
	CaseType     string   `json:"case_type" binding:"required,oneof=civil criminal administrative commercial labor"`
	Summary      string   `json:"summary"`
	AcceptDate   *string  `json:"accept_date"`
	CoLawyerIDs  []uint64 `json:"co_lawyer_ids"`
}

// CaseUpdateRequest 更新案件请求。
type CaseUpdateRequest struct {
	Title       string   `json:"title" binding:"max=200"`
	Summary     string   `json:"summary"`
	CoLawyerIDs []uint64 `json:"co_lawyer_ids"`
}

// CaseStatusRequest 状态流转请求。
type CaseStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=filed investigating hearing closed archived"`
}

// CaseAssignRequest 分配律师请求。
type CaseAssignRequest struct {
	LeadLawyerID uint64   `json:"lead_lawyer_id" binding:"required"`
	CoLawyerIDs  []uint64 `json:"co_lawyer_ids"`
}

// ParseAcceptDate 解析接受日期字符串。
func ParseAcceptDate(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
