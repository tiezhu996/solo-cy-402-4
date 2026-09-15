package constants

// CaseStatus 案件状态枚举。
const (
	CaseStatusFiled         = "filed"
	CaseStatusInvestigating = "investigating"
	CaseStatusHearing       = "hearing"
	CaseStatusClosed        = "closed"
	CaseStatusArchived      = "archived"
)

// CaseStatusValues 全部案件状态值。
var CaseStatusValues = []string{CaseStatusFiled, CaseStatusInvestigating, CaseStatusHearing, CaseStatusClosed, CaseStatusArchived}

// CaseType 案件类型枚举。
const (
	CaseTypeCivil      = "civil"
	CaseTypeCriminal   = "criminal"
	CaseTypeAdmin      = "administrative"
	CaseTypeCommercial = "commercial"
	CaseTypeLabor      = "labor"
)

// CaseTypeValues 全部案件类型值。
var CaseTypeValues = []string{CaseTypeCivil, CaseTypeCriminal, CaseTypeAdmin, CaseTypeCommercial, CaseTypeLabor}

// IsValidCaseStatus 校验案件状态。
func IsValidCaseStatus(s string) bool {
	for _, v := range CaseStatusValues {
		if v == s {
			return true
		}
	}
	return false
}

// IsValidCaseType 校验案件类型。
func IsValidCaseType(s string) bool {
	for _, v := range CaseTypeValues {
		if v == s {
			return true
		}
	}
	return false
}

// UserRole 用户角色枚举。
const (
	RoleAdmin     = "admin"
	RoleLawyer    = "lawyer"
	RoleAssistant = "assistant"
)

// UserRoleValues 全部角色值。
var UserRoleValues = []string{RoleAdmin, RoleLawyer, RoleAssistant}
