import { useEffect, useState } from 'react'
import { Button, Card, Drawer, Form, Input, message, Modal, Space, Table } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { createClient, deleteClient, getClient } from '@/api/client'
import { useClientStore } from '@/stores/clientStore'
import ClientCard from '@/components/common/ClientCard'
import CaseTable from '@/components/common/CaseTable'
import type { Client, CaseItem } from '@/types'

export default function Clients() {
  const store = useClientStore()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [keyword, setKeyword] = useState('')
  const [open, setOpen] = useState(false)
  const [detail, setDetail] = useState<Client | null>(null)
  const [cases, setCases] = useState<CaseItem[]>([])
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    store.fetchList({ page, page_size: pageSize, keyword })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, pageSize, keyword])

  async function onCreate() {
    const values = await form.validateFields()
    await createClient(values)
    message.success('客户创建成功')
    setOpen(false)
    form.resetFields()
    setPage(1)
  }

  async function showDetail(id: number) {
    const res: any = await getClient(id)
    setDetail(res.data.client)
    setCases(res.data.cases)
    setDrawerOpen(true)
  }

  async function onDelete(id: number) {
    await deleteClient(id)
    message.success('已删除')
    store.fetchList({ page, page_size: pageSize, keyword })
  }

  return (
    <Card>
      <Space style={{ marginBottom: 16 }}>
        <Input.Search placeholder="检索客户姓名/证件号" allowClear style={{ width: 260 }} onSearch={(v) => { setKeyword(v); setPage(1) }} />
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setOpen(true)}>新建客户</Button>
      </Space>
      <Table<Client>
        rowKey="id"
        dataSource={store.list}
        pagination={{ current: page, pageSize, total: store.total, onChange: (p, ps) => { setPage(p); setPageSize(ps) } }}
        columns={[
          { title: 'ID', dataIndex: 'id' },
          { title: '姓名', dataIndex: 'name' },
          { title: '证件号', dataIndex: 'id_number' },
          { title: '联系方式', dataIndex: 'contact' },
          { title: '地址', dataIndex: 'address' },
          {
            title: '操作',
            render: (_, row) => (
              <>
                <a onClick={() => showDetail(row.id)}>详情</a>
                <Button type="link" danger onClick={() => onDelete(row.id)}>删除</Button>
              </>
            ),
          },
        ]}
      />
      <Modal title="新建客户" open={open} onOk={onCreate} onCancel={() => setOpen(false)}>
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="姓名" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="id_number" label="证件号"><Input /></Form.Item>
          <Form.Item name="contact" label="联系方式"><Input /></Form.Item>
          <Form.Item name="address" label="地址"><Input /></Form.Item>
          <Form.Item name="remark" label="备注"><Input.TextArea rows={3} /></Form.Item>
        </Form>
      </Modal>
      <Drawer title={detail?.name || '客户详情'} open={drawerOpen} onClose={() => setDrawerOpen(false)} width={560}>
        {detail && <ClientCard client={detail} />}
        <h3 style={{ marginTop: 16 }}>历史案件</h3>
        <CaseTable cases={cases} />
      </Drawer>
    </Card>
  )
}
