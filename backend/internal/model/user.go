package model

import "time"

// User 用户/律师实体。
type User struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:100;not null" json:"-"`
	RealName     string    `gorm:"size:50;not null;default:''" json:"real_name"`
	Role         string    `gorm:"size:20;not null;default:lawyer" json:"role"`
	LicenseNo    string    `gorm:"size:50;not null;default:''" json:"license_no"`
	Email        string    `gorm:"size:100;not null;default:''" json:"email"`
	Phone        string    `gorm:"size:20;not null;default:''" json:"phone"`
	Avatar       string    `gorm:"size:255;not null;default:''" json:"avatar"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName 指定表名。
func (User) TableName() string { return "users" }
