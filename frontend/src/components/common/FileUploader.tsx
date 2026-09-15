import { Button, Upload } from 'antd'
import { UploadOutlined } from '@ant-design/icons'
import { useFileUpload } from '@/hooks/useFileUpload'

export default function FileUploader({ onUploaded }: { onUploaded: (url: string) => void }) {
  const { uploading, upload } = useFileUpload()
  return (
    <Upload
      showUploadList={false}
      beforeUpload={(file) => {
        upload(file).then((url) => onUploaded(url))
        return false
      }}
    >
      <Button icon={<UploadOutlined />} loading={uploading}>选择文件上传</Button>
    </Upload>
  )
}
