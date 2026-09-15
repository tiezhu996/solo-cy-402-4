import { Tag } from 'antd'
import { CaseStatusText } from '@/constants/case'
import { BillingStatusText } from '@/constants/billing'

export type StatusKind = 'case' | 'billing'

const caseColor: Record<string, string> = {
  filed: 'blue',
  investigating: 'orange',
  hearing: 'purple',
  closed: 'green',
  archived: 'default',
}

const billingColor: Record<string, string> = {
  pending: 'orange',
  paid: 'green',
  invoiced: 'blue',
  void: 'default',
}

export default function StatusBadge({ status, kind = 'case' }: { status: string; kind?: StatusKind }) {
  const text = kind === 'case' ? CaseStatusText[status] || status : BillingStatusText[status] || status
  const color = kind === 'case' ? caseColor[status] : billingColor[status]
  return <Tag color={color}>{text}</Tag>
}
