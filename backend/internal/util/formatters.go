package util

import (
	"fmt"
	"time"

	"cylawcase/internal/constants"
)

// formatters.go 同时提供日期格式化、状态文本、案件类型文本、账单类型文本，多个 handler/service 直接引用。

// FormatDateTime 格式化日期时间。
func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

// FormatDate 格式化日期。
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// CaseStatusText 案件状态文本。
func CaseStatusText(status string) string {
	switch status {
	case constants.CaseStatusFiled:
		return "已立案"
	case constants.CaseStatusInvestigating:
		return "调查取证"
	case constants.CaseStatusHearing:
		return "庭审中"
	case constants.CaseStatusClosed:
		return "已结案"
	case constants.CaseStatusArchived:
		return "已归档"
	default:
		return status
	}
}

// CaseTypeText 案件类型文本。
func CaseTypeText(t string) string {
	switch t {
	case constants.CaseTypeCivil:
		return "民事"
	case constants.CaseTypeCriminal:
		return "刑事"
	case constants.CaseTypeAdmin:
		return "行政"
	case constants.CaseTypeCommercial:
		return "商事"
	case constants.CaseTypeLabor:
		return "劳动"
	default:
		return t
	}
}

// BillingTypeText 费用类型文本。
func BillingTypeText(t string) string {
	switch t {
	case constants.BillingTypeAttorneyFee:
		return "律师费"
	case constants.BillingTypeCourtFee:
		return "诉讼费"
	case constants.BillingTypeTravelFee:
		return "差旅费"
	case constants.BillingTypeOther:
		return "其他"
	default:
		return t
	}
}

// BillingStatusText 账单状态文本。
func BillingStatusText(s string) string {
	switch s {
	case constants.BillingStatusPending:
		return "待支付"
	case constants.BillingStatusPaid:
		return "已支付"
	case constants.BillingStatusInvoiced:
		return "已开票"
	case constants.BillingStatusVoid:
		return "已作废"
	default:
		return s
	}
}

// DocumentTypeText 文档类型文本。
func DocumentTypeText(t string) string {
	switch t {
	case constants.DocTypeComplaint:
		return "起诉状"
	case constants.DocTypeDefense:
		return "答辩状"
	case constants.DocTypeEvidence:
		return "证据"
	case constants.DocTypeJudgment:
		return "判决书"
	case constants.DocTypeContract:
		return "合同"
	default:
		return "其他"
	}
}

// FormatMoney 格式化金额。
func FormatMoney(v float64) string {
	return fmt.Sprintf("¥%.2f", v)
}
