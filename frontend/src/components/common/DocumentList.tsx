import { Empty, List } from 'antd'
import DocumentCard from './DocumentCard'
import type { DocumentItem } from '@/types'

export default function DocumentList({ documents }: { documents: DocumentItem[] }) {
  if (documents.length === 0) return <Empty description="暂无文档" />
  return (
    <List
      grid={{ gutter: 16, column: 2 }}
      dataSource={documents}
      renderItem={(doc) => (
        <List.Item>
          <DocumentCard item={doc} />
        </List.Item>
      )}
    />
  )
}
