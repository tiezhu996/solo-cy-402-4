package constants

// BillingType 费用类型枚举。
const (
	BillingTypeAttorneyFee = "attorney_fee"
	BillingTypeCourtFee    = "court_fee"
	BillingTypeTravelFee   = "travel_fee"
	BillingTypeOther       = "other"
)

// BillingTypeValues 全部费用类型值。
var BillingTypeValues = []string{BillingTypeAttorneyFee, BillingTypeCourtFee, BillingTypeTravelFee, BillingTypeOther}

// BillingStatus 账单状态枚举。
const (
	BillingStatusPending  = "pending"
	BillingStatusPaid     = "paid"
	BillingStatusInvoiced = "invoiced"
	BillingStatusVoid     = "void"
)

// BillingStatusValues 全部账单状态值。
var BillingStatusValues = []string{BillingStatusPending, BillingStatusPaid, BillingStatusInvoiced, BillingStatusVoid}

// IsValidBillingType 校验费用类型。
func IsValidBillingType(s string) bool {
	for _, v := range BillingTypeValues {
		if v == s {
			return true
		}
	}
	return false
}

// IsValidBillingStatus 校验账单状态。
func IsValidBillingStatus(s string) bool {
	for _, v := range BillingStatusValues {
		if v == s {
			return true
		}
	}
	return false
}

// DocumentFileType 文档类型枚举。
const (
	DocTypeComplaint = "complaint"
	DocTypeDefense   = "defense"
	DocTypeEvidence  = "evidence"
	DocTypeJudgment  = "judgment"
	DocTypeContract  = "contract"
	DocTypeOther     = "other"
)

// DocumentFileTypeValues 全部文档类型值。
var DocumentFileTypeValues = []string{DocTypeComplaint, DocTypeDefense, DocTypeEvidence, DocTypeJudgment, DocTypeContract, DocTypeOther}
