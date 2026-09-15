import { useEffect, useState } from 'react'
import { Button, Card, Form, Input, InputNumber, message, Modal, Select, Space, Table, Tag } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import AmountSummary from '@/components/common/AmountSummary'
import StatusBadge from '@/components/common/StatusBadge'
import { useBillingStore } from '@/stores/billingStore'
import { createBilling, markPaid, markInvoiced, voidBilling } from '@/api/billing'
import { BillingStatusOptions, BillingTypeOptions, BillingTypeText } from '@/constants/billing'
import { formatAmount } from '@/utils/amountFormatter'
import type { Billing } from '@/types'

export default function Billing() {
  const store = useBillingStore()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [filters, setFilters] = useState<Record<string, unknown>>({})
  const [open, setOpen] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    store.fetchList({ page, page_size: pageSize, ...filters })
    store.fetchSummary()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, pageSize, filters])

  async function onCreate() {
    const values = await form.validateFields()
    await createBilling(values)
    message.success('账单创建成功')
    setOpen(false)
    form.resetFields()
    setPage(1)
  }

  return (
    <Card>
      <AmountSummary receivable={store.summary.receivable} received={store.summary.received} pending={store.summary.pending} />
      <Space style={{ marginBottom: 16 }}>
        <Select placeholder="账单状态" allowClear style={{ width: 140 }} options={BillingStatusOptions} onChange={(v) => { setFilters({ status: v }); setPage(1) }} />
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setOpen(true)}>创建账单</Button>
      </Space>
      <Table<Billing>
        rowKey="id"
        dataSource={store.list}
        pagination={{ current: page, pageSize, total: store.total, onChange: (p, ps) => { setPage(p); setPageSize(ps) } }}
        columns={[
          { title: '单号', dataIndex: 'bill_no' },
          { title: '类型', dataIndex: 'billing_type', render: (v) => BillingTypeText[v] || v },
          { title: '金额', dataIndex: 'amount', render: (v) => formatAmount(v) },
          { title: '状态', dataIndex: 'status', render: (v) => <StatusBadge status={v} kind="billing" /> },
          { title: '案件ID', dataIndex: 'case_id' },
          { title: '客户ID', dataIndex: 'client_id' },
          {
            title: '操作',
            render: (_, row) => (
              <Space>
                {row.status === 'pending' && <Button size="small" type="primary" onClick={async () => { await markPaid(row.id); message.success('已标记支付'); store.fetchList({ page, page_size: pageSize, ...filters }); store.fetchSummary() }}>标记支付</Button>}
                {row.status === 'paid' && <Button size="small" onClick={async () => { await markInvoiced(row.id); message.success('已开票'); store.fetchList({ page, page_size: pageSize, ...filters }); store.fetchSummary() }}>开票</Button>}
                {row.status !== 'void' && <Button size="small" danger onClick={async () => { await voidBilling(row.id); message.success('已作废'); store.fetchList({ page, page_size: pageSize, ...filters }); store.fetchSummary() }}>作废</Button>}
              </Space>
            ),
          },
        ]}
      />
      <Modal title="创建账单" open={open} onOk={onCreate} onCancel={() => setOpen(false)}>
        <Form form={form} layout="vertical">
          <Form.Item name="case_id" label="案件ID" rules={[{ required: true }]}><InputNumber style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="client_id" label="客户ID" rules={[{ required: true }]}><InputNumber style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="billing_type" label="费用类型" rules={[{ required: true }]}><Select options={BillingTypeOptions} /></Form.Item>
          <Form.Item name="amount" label="金额" rules={[{ required: true }]}><InputNumber style={{ width: '100%' }} min={0} precision={2} /></Form.Item>
          <Form.Item name="invoice_info" label="发票信息"><Input /></Form.Item>
        </Form>
      </Modal>
    </Card>
  )
}
