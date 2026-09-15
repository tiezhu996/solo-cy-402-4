import { Card, Typography } from 'antd'
import { DocumentTypeText } from '@/constants/document'
import { formatDateTime } from '@/utils/dateFormat'
import type { DocumentItem } from '@/types'

export default function DocumentCard({ item }: { item: DocumentItem }) {
  return (
    <Card size="small" title={item.title}>
      <Typography.Paragraph type="secondary" style={{ marginBottom: 4 }}>
        类型：{DocumentTypeText[item.file_type] || item.file_type}
      </Typography.Paragraph>
      <Typography.Paragraph type="secondary" style={{ marginBottom: 4 }}>
        上传时间：{formatDateTime(item.upload_time)}
      </Typography.Paragraph>
      <Typography.Link href={item.file_url} target="_blank">查看/下载</Typography.Link>
    </Card>
  )
}
