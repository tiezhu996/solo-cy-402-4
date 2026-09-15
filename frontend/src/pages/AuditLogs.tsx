import { useEffect, useState } from 'react'
import { Card, Table } from 'antd'
import { listAuditLogs } from '@/api/auditLog'
import { formatDateTime } from '@/utils/dateFormat'
import type { AuditLog } from '@/types'

export default function AuditLogs() {
  const [list, setList] = useState<AuditLog[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    setLoading(true)
    listAuditLogs({ page, page_size: pageSize }).then((res: any) => {
      setList(res.data.list)
      setTotal(res.data.total)
    }).finally(() => setLoading(false))
  }, [page, pageSize])

  return (
    <Card title="审计日志">
      <Table<AuditLog>
        rowKey="id"
        loading={loading}
        dataSource={list}
        pagination={{ current: page, pageSize, total, onChange: (p, ps) => { setPage(p); setPageSize(ps) } }}
        columns={[
          { title: 'ID', dataIndex: 'id' },
          { title: '操作人', dataIndex: 'operator_name' },
          { title: '动作', dataIndex: 'action' },
          { title: '实体', dataIndex: 'entity_type' },
          { title: '实体ID', dataIndex: 'entity_id' },
          { title: 'IP', dataIndex: 'ip' },
          { title: '时间', dataIndex: 'created_at', render: (v) => formatDateTime(v) },
        ]}
      />
    </Card>
  )
}
