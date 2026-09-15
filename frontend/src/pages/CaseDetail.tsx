import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { Alert, Card, Descriptions, Tabs, Button, Select, Space, Tag, message } from 'antd'
import { getCase, changeCaseStatus, assignLawyer } from '@/api/case'
import { getClient } from '@/api/client'
import DocumentList from '@/components/common/DocumentList'
import BillingCard from '@/components/common/BillingCard'
import StatusBadge from '@/components/common/StatusBadge'
import PermissionGuard from '@/components/common/PermissionGuard'
import TimelineItem from '@/components/common/TimelineItem'
import { useDocumentStore } from '@/stores/documentStore'
import { useBillingStore } from '@/stores/billingStore'
import { useUserStore } from '@/stores/userStore'
import { CaseStatusOptions, CaseTypeOptions, CaseStatus } from '@/constants/case'
import { BillingStatus } from '@/constants/billing'
import type { CaseItem, Client } from '@/types'

export default function CaseDetail() {
  const { id } = useParams()
  const caseId = Number(id)
  const [item, setItem] = useState<CaseItem | null>(null)
  const [client, setClient] = useState<Client | null>(null)
  const [status, setStatus] = useState('')
  const [lawyer, setLawyer] = useState<number>()
  const docStore = useDocumentStore()
  const billingStore = useBillingStore()
  const userStore = useUserStore()

  useEffect(() => {
    userStore.fetchLawyers()
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [caseId])

  async function load() {
    const res: any = await getCase(caseId)
    setItem(res.data)
    setStatus(res.data.status)
    if (res.data.client_id) {
      const cr: any = await getClient(res.data.client_id)
      setClient(cr.data.client)
    }
    docStore.fetchByCase(caseId)
    billingStore.fetchByCase(caseId)
  }

  const pendingBills = billingStore.byCase.filter((b) => b.status === BillingStatus.PENDING)

  async function onStatusChange() {
    // 归档前费用收口：前端先提示，后端在事务内再次强校验（并发情况下以后端为准）。
    if (status === CaseStatus.ARCHIVED && pendingBills.length > 0) {
      message.error(`仍有 ${pendingBills.length} 笔待支付账单，补齐支付或作废后才能归档`)
      return
    }
    try {
      await changeCaseStatus(caseId, status)
      message.success('状态已更新')
    } catch {
      // 被后端拒绝（如归档时仍有待支付账单）：回拉案件，状态与结案日期保持为服务端值。
      setStatus(item?.status ?? status)
      await load()
      return
    }
    load()
  }

  async function onAssign() {
    if (!lawyer) return
    await assignLawyer(caseId, { lead_lawyer_id: lawyer })
    message.success('律师已分配')
    load()
  }

  if (!item) return null

  return (
    <Card>
      <Space style={{ marginBottom: 16 }}>
        <h2 style={{ margin: 0 }}>{item.case_no} {item.title}</h2>
        <StatusBadge status={item.status} />
        <Tag>{CaseTypeOptions.find((o) => o.value === item.case_type)?.label || item.case_type}</Tag>
      </Space>
      <Tabs
        items={[
          {
            key: 'info',
            label: '案件信息',
            children: (
              <>
                <Descriptions bordered column={2} size="small">
                  <Descriptions.Item label="案号">{item.case_no}</Descriptions.Item>
                  <Descriptions.Item label="状态">{item.status}</Descriptions.Item>
                  <Descriptions.Item label="类型">{item.case_type}</Descriptions.Item>
                  <Descriptions.Item label="主办律师">#{item.lead_lawyer_id}</Descriptions.Item>
                  <Descriptions.Item label="受理日期">{item.accept_date || '-'}</Descriptions.Item>
                  <Descriptions.Item label="结案日期">{item.close_date || '-'}</Descriptions.Item>
                  <Descriptions.Item label="摘要" span={2}>{item.summary || '-'}</Descriptions.Item>
                </Descriptions>
                <PermissionGuard roles={['admin', 'lawyer']}>
                  {item.status === CaseStatus.CLOSED && pendingBills.length > 0 && (
                    <Alert
                      style={{ marginTop: 16 }}
                      type="warning"
                      showIcon
                      message={`本案有 ${pendingBills.length} 笔待支付账单`}
                      description="归档前必须补齐支付或作废全部待支付账单，否则案件会保持“已结案”，无法归档。"
                    />
                  )}
                  {item.status === CaseStatus.ARCHIVED && (
                    <Alert
                      style={{ marginTop: 16 }}
                      type="info"
                      showIcon
                      message="案件已归档，费用已收口"
                      description="归档为稳定终态，不能再新增待支付账单，状态与结案日期不再变化。"
                    />
                  )}
                  <Space style={{ marginTop: 16 }}>
                    <Select
                      value={status}
                      style={{ width: 150 }}
                      options={CaseStatusOptions}
                      onChange={setStatus}
                      disabled={item.status === CaseStatus.ARCHIVED}
                    />
                    <Button
                      type="primary"
                      onClick={onStatusChange}
                      disabled={item.status === CaseStatus.ARCHIVED}
                    >
                      更新状态
                    </Button>
                  </Space>
                  <Space style={{ marginTop: 8 }}>
                    <Select
                      placeholder="分配主办律师"
                      style={{ width: 180 }}
                      value={lawyer}
                      onChange={setLawyer}
                      options={userStore.lawyers.map((l) => ({ label: l.real_name || l.username, value: l.id }))}
                    />
                    <Button onClick={onAssign}>分配</Button>
                  </Space>
                </PermissionGuard>
              </>
            ),
          },
          {
            key: 'client',
            label: '关联客户',
            children: client ? (
              <Descriptions bordered column={1} size="small">
                <Descriptions.Item label="姓名">{client.name}</Descriptions.Item>
                <Descriptions.Item label="证件号">{client.id_number}</Descriptions.Item>
                <Descriptions.Item label="联系方式">{client.contact}</Descriptions.Item>
                <Descriptions.Item label="地址">{client.address}</Descriptions.Item>
              </Descriptions>
            ) : null,
          },
          {
            key: 'docs',
            label: '文档',
            children: <DocumentList documents={docStore.byCase} />,
          },
          {
            key: 'billings',
            label: '账单',
            children: (
              <>
                {item.status === CaseStatus.CLOSED && pendingBills.length > 0 && (
                  <Alert
                    style={{ marginBottom: 12 }}
                    type="warning"
                    showIcon
                    message={`${pendingBills.length} 笔待支付账单未结清`}
                    description="请在费用中心将待支付账单标记支付或作废后，再执行归档。"
                  />
                )}
                {item.status === CaseStatus.ARCHIVED && (
                  <Alert
                    style={{ marginBottom: 12 }}
                    type="info"
                    showIcon
                    message="案件已归档，账单只读收口"
                    description="归档后不得再新增待支付账单，仅可查看既有账单。"
                  />
                )}
                {billingStore.byCase.map((b) => <BillingCard key={b.id} item={b} />)}
              </>
            ),
          },
          {
            key: 'timeline',
            label: '时间线',
            children: (
              <TimelineItem
                items={[
                  { id: 1, time: item.created_at, text: `案件创建（${item.case_no}）` },
                  { id: 2, time: item.accept_date || item.created_at, text: '案件受理' },
                  { id: 3, time: item.close_date || item.created_at, text: item.close_date ? '案件结案' : '案件进行中' },
                ]}
              />
            ),
          },
        ]}
      />
    </Card>
  )
}
