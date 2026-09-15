import request from '@/utils/request'

export function listAuditLogs(params: { page?: number; page_size?: number }) {
  return request.get('/audit-logs', { params })
}
