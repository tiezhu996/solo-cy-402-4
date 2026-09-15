package model

import "time"

// Client 客户实体。
type Client struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	IDNumber  string    `gorm:"size:50;not null;default:''" json:"id_number"`
	Contact   string    `gorm:"size:50;not null;default:''" json:"contact"`
	Address   string    `gorm:"size:255;not null;default:''" json:"address"`
	Remark    string    `gorm:"type:text" json:"remark"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名。
func (Client) TableName() string { return "clients" }
