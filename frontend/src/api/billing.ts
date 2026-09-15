import request from '@/utils/request'

export function listBillings(params: { page?: number; page_size?: number; case_id?: number; client_id?: number; status?: string }) {
  return request.get('/billings', { params })
}

export function listBillingsByCase(caseId: number) {
  return request.get(`/billings/by-case/${caseId}`)
}

export function createBilling(data: { case_id: number; client_id: number; billing_type: string; amount: number; invoice_info?: string }) {
  return request.post('/billings', data)
}

export function markPaid(id: number) {
  return request.post(`/billings/${id}/paid`)
}

export function markInvoiced(id: number, invoiceInfo?: string) {
  return request.post(`/billings/${id}/invoiced`, { invoice_info: invoiceInfo })
}

export function voidBilling(id: number) {
  return request.post(`/billings/${id}/void`)
}

export function getBillingSummary() {
  return request.get('/billings/summary')
}
