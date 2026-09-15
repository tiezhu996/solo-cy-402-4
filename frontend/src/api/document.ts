import request from '@/utils/request'

export function listDocuments(params: { page?: number; page_size?: number; file_type?: string; keyword?: string }) {
  return request.get('/documents', { params })
}

export function listDocumentsByCase(caseId: number) {
  return request.get(`/documents/by-case/${caseId}`)
}

export function createDocument(data: { case_id: number; title: string; file_type: string; file_url: string }) {
  return request.post('/documents', data)
}

export function deleteDocument(id: number) {
  return request.delete(`/documents/${id}`)
}
