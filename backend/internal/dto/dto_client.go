package dto

// ClientCreateRequest 新建客户请求。
type ClientCreateRequest struct {
	Name     string `json:"name" binding:"required,max=100"`
	IDNumber string `json:"id_number" binding:"max=50"`
	Contact  string `json:"contact" binding:"max=50"`
	Address  string `json:"address" binding:"max=255"`
	Remark   string `json:"remark"`
}

// ClientUpdateRequest 编辑客户请求。
type ClientUpdateRequest struct {
	Name     string `json:"name" binding:"max=100"`
	IDNumber string `json:"id_number" binding:"max=50"`
	Contact  string `json:"contact" binding:"max=50"`
	Address  string `json:"address" binding:"max=255"`
	Remark   string `json:"remark"`
}
