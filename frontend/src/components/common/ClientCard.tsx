import { Card, Descriptions } from 'antd'
import type { Client } from '@/types'

export default function ClientCard({ client }: { client: Client }) {
  return (
    <Card title={client.name} size="small">
      <Descriptions column={1} size="small">
        <Descriptions.Item label="证件号">{client.id_number || '-'}</Descriptions.Item>
        <Descriptions.Item label="联系方式">{client.contact || '-'}</Descriptions.Item>
        <Descriptions.Item label="地址">{client.address || '-'}</Descriptions.Item>
        <Descriptions.Item label="备注">{client.remark || '-'}</Descriptions.Item>
      </Descriptions>
    </Card>
  )
}
