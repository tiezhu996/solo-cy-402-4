package dto

// BillingCreateRequest 创建账单请求。
type BillingCreateRequest struct {
	CaseID      uint64  `json:"case_id" binding:"required"`
	ClientID    uint64  `json:"client_id" binding:"required"`
	BillingType string  `json:"billing_type" binding:"required,oneof=attorney_fee court_fee travel_fee other"`
	Amount      float64 `json:"amount" binding:"required"`
	InvoiceInfo string  `json:"invoice_info" binding:"max=255"`
}

// BillingInvoiceRequest 开票请求。
type BillingInvoiceRequest struct {
	InvoiceInfo string `json:"invoice_info" binding:"max=255"`
}
