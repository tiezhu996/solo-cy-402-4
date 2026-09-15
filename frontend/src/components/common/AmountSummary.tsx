import { Card, Col, Row, Statistic } from 'antd'
import { formatAmountPlain } from '@/utils/amountFormatter'

export default function AmountSummary({ receivable, received, pending }: { receivable: number; received: number; pending: number }) {
  return (
    <Row gutter={16} style={{ marginBottom: 16 }}>
      <Col span={8}>
        <Card><Statistic title="本月应收" value={formatAmountPlain(receivable)} prefix="¥" /></Card>
      </Col>
      <Col span={8}>
        <Card><Statistic title="本月已收" value={formatAmountPlain(received)} prefix="¥" valueStyle={{ color: '#3f8600' }} /></Card>
      </Col>
      <Col span={8}>
        <Card><Statistic title="本月待收" value={formatAmountPlain(pending)} prefix="¥" valueStyle={{ color: '#cf1322' }} /></Card>
      </Col>
    </Row>
  )
}
