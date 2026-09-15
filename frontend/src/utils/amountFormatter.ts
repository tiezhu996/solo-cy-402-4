export function formatAmount(value?: number | string | null): string {
  if (value === null || value === undefined || value === '') return '¥0.00'
  const n = Number(value)
  if (Number.isNaN(n)) return String(value)
  return `¥${n.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

export function formatAmountPlain(value?: number | string | null): string {
  if (value === null || value === undefined || value === '') return '0.00'
  const n = Number(value)
  if (Number.isNaN(n)) return String(value)
  return n.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
