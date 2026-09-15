import { useEffect, useState } from 'react'
import { Button, Modal, Form, Input, Select, DatePicker, message, Card, Table } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import FilterBar from '@/components/common/FilterBar'
import { useCaseStore } from '@/stores/caseStore'
import { useUserStore } from '@/stores/userStore'
import { useClientStore } from '@/stores/clientStore'
import { createCase } from '@/api/case'
import { CaseStatusOptions, CaseTypeOptions } from '@/constants/case'
import StatusBadge from '@/components/common/StatusBadge'
import type { CaseItem } from '@/types'

export default function Cases() {
  const store = useCaseStore()
  const userStore = useUserStore()
  const clientStore = useClientStore()
  const navigate = useNavigate()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [filters, setFilters] = useState<Record<string, unknown>>({})
  const [open, setOpen] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    userStore.fetchLawyers()
    clientStore.fetchList({ page: 1, page_size: 200 })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    store.fetchList({ page, page_size: pageSize, ...filters })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, pageSize, filters])

  async function onCreate() {
    const values = await form.validateFields()
    await createCase({
      title: values.title,
      case_type: values.case_type,
      client_id: values.client_id,
      lead_lawyer_id: values.lead_lawyer_id,
      accept_date: values.accept_date ? values.accept_date.format('YYYY-MM-DD') : undefined,
      summary: values.summary,
    })
    message.success('案件创建成功')
    setOpen(false)
    form.resetFields()
    setPage(1)
  }

  return (
    <Card>
      <FilterBar
        typeOptions={CaseTypeOptions}
        statusOptions={CaseStatusOptions}
        lawyerOptions={userStore.lawyers.map((l) => ({ label: l.real_name || l.username, value: l.id }))}
        onSearch={(v) => { setFilters(v); setPage(1) }}
      />
      <Button type="primary" icon={<PlusOutlined />} style={{ marginBottom: 16 }} onClick={() => setOpen(true)}>
        创建案件
      </Button>
      <Table<CaseItem>
        rowKey="id"
        dataSource={store.list}
        loading={false}
        pagination={{ current: page, pageSize, total: store.total, showTotal: (t) => `共 ${t} 条`, onChange: (p, ps) => { setPage(p); setPageSize(ps) } }}
        onRow={(row) => ({ onClick: () => navigate(`/cases/${row.id}`), style: { cursor: 'pointer' } })}
        columns={[
          { title: '案号', dataIndex: 'case_no' },
          { title: '标题', dataIndex: 'title' },
          { title: '类型', dataIndex: 'case_type', render: (v) => CaseTypeOptions.find((o) => o.value === v)?.label || v },
          { title: '状态', dataIndex: 'status', render: (v) => <StatusBadge status={v} /> },
          { title: '主办律师', dataIndex: 'lead_lawyer_id', render: (v) => userStore.lawyers.find((l) => l.id === v)?.real_name || v },
        ]}
      />
      <Modal title="创建案件" open={open} onOk={onCreate} onCancel={() => setOpen(false)}>
        <Form form={form} layout="vertical">
          <Form.Item name="title" label="案件标题" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="case_type" label="案件类型" rules={[{ required: true }]}>
            <Select options={CaseTypeOptions} />
          </Form.Item>
          <Form.Item name="client_id" label="客户" rules={[{ required: true }]}>
            <Select options={clientStore.list.map((c) => ({ label: c.name, value: c.id }))} showSearch optionFilterProp="label" />
          </Form.Item>
          <Form.Item name="lead_lawyer_id" label="主办律师" rules={[{ required: true }]}>
            <Select options={userStore.lawyers.map((l) => ({ label: l.real_name || l.username, value: l.id }))} />
          </Form.Item>
          <Form.Item name="accept_date" label="受理日期">
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="summary" label="摘要">
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  )
}
