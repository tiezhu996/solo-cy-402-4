package dto

// RegisterRequest 注册请求。
type RegisterRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=50"`
	Password  string `json:"password" binding:"required,min=6,max=72"`
	RealName  string `json:"real_name" binding:"max=50"`
	Role      string `json:"role" binding:"omitempty,oneof=admin lawyer assistant"`
	LicenseNo string `json:"license_no" binding:"max=50"`
	Email     string `json:"email" binding:"omitempty,email,max=100"`
	Phone     string `json:"phone" binding:"max=20"`
}

// LoginRequest 登录请求。
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应。
type LoginResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

// UpdateProfileRequest 修改资料请求。
type UpdateProfileRequest struct {
	RealName  string `json:"real_name" binding:"max=50"`
	Avatar    string `json:"avatar" binding:"max=255"`
	Email     string `json:"email" binding:"omitempty,email,max=100"`
	Phone     string `json:"phone" binding:"max=20"`
	LicenseNo string `json:"license_no" binding:"max=50"`
}
