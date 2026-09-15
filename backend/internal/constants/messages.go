package constants

// messages.go 同时承载前端提示文案、后端返回文案与日志文案。
const (
	MsgOK                    = "ok"
	MsgInvalidParams         = "参数不合法"
	MsgUnauthorized          = "未登录或登录已过期"
	MsgForbidden             = "没有权限执行该操作"
	MsgNotFound              = "资源不存在"
	MsgTooManyRequests       = "请求过于频繁，请稍后再试"
	MsgInternalError         = "服务器内部错误"
	MsgUsernameExists        = "用户名已存在"
	MsgInvalidCredentials    = "用户名或密码错误"
	MsgCaseStatusConflict    = "案件状态流转冲突"
	MsgBillingStatusConflict = "账单状态流转冲突"
	MsgUploadTooLarge        = "上传文件过大"
	MsgUnsupportedFileType   = "不支持的文件类型"
	MsgLoginSuccess          = "登录成功"
	MsgClientCreated         = "客户创建成功"
	MsgCaseCreated           = "案件创建成功"
	MsgDocumentUploaded      = "文档上传成功"
	MsgBillingCreated        = "账单创建成功"
	MsgBillingPaid           = "账单已标记支付"
	MsgBillingInvoiced       = "账单已开票"
	MsgBillingVoided         = "账单已作废"
)
