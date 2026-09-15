import { Button, DatePicker, Input, Select, Space } from 'antd'
import { useState } from 'react'

interface FilterBarProps {
  onSearch: (values: Record<string, unknown>) => void
  typeOptions?: { label: string; value: string }[]
  statusOptions?: { label: string; value: string }[]
  lawyerOptions?: { label: string; value: number }[]
  showKeyword?: boolean
  placeholder?: string
}

export default function FilterBar({ onSearch, typeOptions = [], statusOptions = [], lawyerOptions = [], showKeyword = true, placeholder = '搜索关键词' }: FilterBarProps) {
  const [values, setValues] = useState<Record<string, unknown>>({})
  return (
    <Space wrap style={{ marginBottom: 16 }}>
      {showKeyword && <Input placeholder={placeholder} allowClear style={{ width: 180 }} onChange={(e) => setValues((v) => ({ ...v, keyword: e.target.value }))} />}
      {typeOptions.length > 0 && (
        <Select placeholder="类型" allowClear style={{ width: 130 }} options={typeOptions} onChange={(v) => setValues((p) => ({ ...p, case_type: v }))} />
      )}
      {statusOptions.length > 0 && (
        <Select placeholder="状态" allowClear style={{ width: 130 }} options={statusOptions} onChange={(v) => setValues((p) => ({ ...p, status: v }))} />
      )}
      {lawyerOptions.length > 0 && (
        <Select placeholder="主办律师" allowClear style={{ width: 140 }} options={lawyerOptions} onChange={(v) => setValues((p) => ({ ...p, lead_lawyer_id: v }))} />
      )}
      <Button type="primary" onClick={() => onSearch(values)}>查询</Button>
    </Space>
  )
}
