import { Table } from 'antd'
import { useNavigate } from 'react-router-dom'
import StatusBadge from './StatusBadge'
import { CaseTypeText } from '@/constants/case'
import type { CaseItem } from '@/types'

export default function CaseTable({ cases, loading = false }: { cases: CaseItem[]; loading?: boolean }) {
  const navigate = useNavigate()
  return (
    <Table<CaseItem>
      rowKey="id"
      loading={loading}
      dataSource={cases}
      pagination={false}
      columns={[
        { title: '案号', dataIndex: 'case_no' },
        { title: '标题', dataIndex: 'title' },
        { title: '类型', dataIndex: 'case_type', render: (v) => CaseTypeText[v] || v },
        { title: '状态', dataIndex: 'status', render: (v) => <StatusBadge status={v} /> },
        {
          title: '操作',
          render: (_, row) => (
            <a onClick={() => navigate(`/cases/${row.id}`)}>查看</a>
          ),
        },
      ]}
    />
  )
}
