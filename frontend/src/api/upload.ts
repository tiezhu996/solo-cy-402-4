import request from '@/utils/request'

export function uploadFile(file: File) {
  const form = new FormData()
  form.append('file', file)
  return request.post('/upload/file', form, { headers: { 'Content-Type': 'multipart/form-data' } })
}
