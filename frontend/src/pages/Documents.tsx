import { useEffect, useState } from 'react'
import { Card, Form, Input, message, Modal, Select, Space, Table, Button } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useDocumentStore } from '@/stores/documentStore'
import { createDocument, deleteDocument } from '@/api/document'
import { DocumentTypeOptions } from '@/constants/document'
import FileUploader from '@/components/common/FileUploader'
import { formatDateTime } from '@/utils/dateFormat'
import type { DocumentItem } from '@/types'

export default function Documents() {
  const store = useDocumentStore()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [filters, setFilters] = useState<Record<string, unknown>>({})
  const [open, setOpen] = useState(false)
  const [url, setUrl] = useState('')
  const [form] = Form.useForm()

  useEffect(() => {
    store.fetchList({ page, page_size: pageSize, ...filters })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, pageSize, filters])

  async function onCreate() {
    const values = await form.validateFields()
    if (!url) {
      message.warning('请先上传文件')
      return
    }
    await createDocument({ ...values, file_url: url })
    message.success('文档上传成功')
    setOpen(false)
    setUrl('')
    form.resetFields()
    setPage(1)
  }

  return (
    <Card>
      <Space style={{ marginBottom: 16 }}>
        <Input.Search placeholder="搜索文档标题" allowClear style={{ width: 240 }} onSearch={(v) => { setFilters({ keyword: v }); setPage(1) }} />
        <Select placeholder="文件类型" allowClear style={{ width: 140 }} options={DocumentTypeOptions} onChange={(v) => { setFilters({ file_type: v }); setPage(1) }} />
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setOpen(true)}>上传文档</Button>
      </Space>
      <Table<DocumentItem>
        rowKey="id"
        dataSource={store.list}
        pagination={{ current: page, pageSize, total: store.total, onChange: (p, ps) => { setPage(p); setPageSize(ps) } }}
        columns={[
          { title: 'ID', dataIndex: 'id' },
          { title: '标题', dataIndex: 'title' },
          { title: '案件ID', dataIndex: 'case_id' },
          { title: '类型', dataIndex: 'file_type', render: (v) => DocumentTypeOptions.find((o) => o.value === v)?.label || v },
          { title: '上传时间', dataIndex: 'upload_time', render: (v) => formatDateTime(v) },
          {
            title: '操作',
            render: (_, row) => (
              <Space>
                <a href={row.file_url} target="_blank" rel="noreferrer">查看</a>
                <Button type="link" danger onClick={async () => { await deleteDocument(row.id); message.success('已删除'); store.fetchList({ page, page_size: pageSize, ...filters }) }}>删除</Button>
              </Space>
            ),
          },
        ]}
      />
      <Modal title="上传文档" open={open} onOk={onCreate} onCancel={() => setOpen(false)}>
        <Form form={form} layout="vertical">
          <Form.Item name="case_id" label="案件ID" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="title" label="文档标题" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="file_type" label="文件类型" rules={[{ required: true }]}><Select options={DocumentTypeOptions} /></Form.Item>
          <Form.Item label="文件">
            <FileUploader onUploaded={(u) => setUrl(u)} />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  )
}
