import { useState } from 'react'
import { message } from 'antd'
import { uploadFile } from '@/api/upload'

export function useFileUpload() {
  const [uploading, setUploading] = useState(false)
  const [url, setUrl] = useState('')

  async function upload(file: File) {
    setUploading(true)
    try {
      const res: any = await uploadFile(file)
      setUrl(res.data.url)
      message.success('上传成功')
      return res.data.url as string
    } finally {
      setUploading(false)
    }
  }

  return { uploading, url, upload }
}
