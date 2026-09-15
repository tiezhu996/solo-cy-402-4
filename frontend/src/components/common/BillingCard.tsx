import { Card, Descriptions } from 'antd'
import StatusBadge from './StatusBadge'
import { BillingTypeText } from '@/constants/billing'
import { formatAmount } from '@/utils/amountFormatter'
import type { Billing } from '@/types'

export default function BillingCard({ item }: { item: Billing }) {
  return (
    <Card size="small" title={`${item.bill_no} ${BillingTypeText[item.billing_type] || item.billing_type}`} extra={<StatusBadge status={item.status} kind="billing" />}>
      <Descriptions column={1} size="small">
        <Descriptions.Item label="金额">{formatAmount(item.amount)}</Descriptions.Item>
        <Descriptions.Item label="案件ID">{item.case_id}</Descriptions.Item>
        <Descriptions.Item label="客户ID">{item.client_id}</Descriptions.Item>
        <Descriptions.Item label="发票信息">{item.invoice_info || '-'}</Descriptions.Item>
      </Descriptions>
    </Card>
  )
}
