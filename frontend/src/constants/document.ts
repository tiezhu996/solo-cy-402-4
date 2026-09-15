// 文档类型枚举（与后端 backend/internal/constants/billing.go 对应）
export const DocumentTypeText: Record<string, string> = {
  complaint: '起诉状',
  defense: '答辩状',
  evidence: '证据',
  judgment: '判决书',
  contract: '合同',
  other: '其他',
}

export const DocumentTypeOptions = Object.entries(DocumentTypeText).map(([value, label]) => ({ label, value }))
