import { Timeline } from 'antd'
import { formatDateTime } from '@/utils/dateFormat'

export interface TimelineData {
  id: number
  time: string
  text: string
  color?: string
}

export default function TimelineItem({ items }: { items: TimelineData[] }) {
  return (
    <Timeline
      items={items.map((it) => ({
        color: it.color || 'blue',
        children: (
          <div>
            <div>{it.text}</div>
            <div style={{ color: '#999', fontSize: 12 }}>{formatDateTime(it.time)}</div>
          </div>
        ),
      }))}
    />
  )
}
