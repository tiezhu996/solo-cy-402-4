// 费用类型/账单状态枚举（与后端 backend/internal/constants/billing.go 保持一致）
export const BillingType = {
  ATTORNEY_FEE: 'attorney_fee',
  COURT_FEE: 'court_fee',
  TRAVEL_FEE: 'travel_fee',
  OTHER: 'other',
} as const

export const BillingTypeText: Record<string, string> = {
  [BillingType.ATTORNEY_FEE]: '律师费',
  [BillingType.COURT_FEE]: '诉讼费',
  [BillingType.TRAVEL_FEE]: '差旅费',
  [BillingType.OTHER]: '其他',
}

export const BillingStatus = {
  PENDING: 'pending',
  PAID: 'paid',
  INVOICED: 'invoiced',
  VOID: 'void',
} as const

export const BillingStatusText: Record<string, string> = {
  [BillingStatus.PENDING]: '待支付',
  [BillingStatus.PAID]: '已支付',
  [BillingStatus.INVOICED]: '已开票',
  [BillingStatus.VOID]: '已作废',
}

export const BillingTypeOptions = Object.entries(BillingTypeText).map(([value, label]) => ({ label, value }))
export const BillingStatusOptions = Object.entries(BillingStatusText).map(([value, label]) => ({ label, value }))
