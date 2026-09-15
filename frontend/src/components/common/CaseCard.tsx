import { Card } from 'antd'
import { useNavigate } from 'react-router-dom'
import StatusBadge from './StatusBadge'
import { CaseTypeText } from '@/constants/case'
import type { CaseItem } from '@/types'

export default function CaseCard({ item }: { item: CaseItem }) {
  const navigate = useNavigate()
  return (
    <Card
      hoverable
      title={`${item.case_no} ${item.title}`}
      extra={<StatusBadge status={item.status} />}
      onClick={() => navigate(`/cases/${item.id}`)}
    >
      <p>类型：{CaseTypeText[item.case_type] || item.case_type}</p>
      <p style={{ color: '#999' }}>{item.summary || '暂无摘要'}</p>
    </Card>
  )
}
