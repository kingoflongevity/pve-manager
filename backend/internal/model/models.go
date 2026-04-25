package model

import (
	"time"

	"gorm.io/gorm"
)

// PVEConfig PVE 连接配置表
type PVEConfig struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Host      string         `gorm:"uniqueIndex;size:255;not null" json:"host"`
	Port      int            `gorm:"not null;default:8006" json:"port"`
	Realm     string         `gorm:"size:50;not null;default:pam" json:"realm"`
	Username  string         `gorm:"size:255;not null" json:"username"`
	Password  string         `gorm:"size:500;not null" json:"-"`
	Name      string         `gorm:"size:255" json:"name"`
	IsDefault bool           `gorm:"default:false" json:"is_default"`
	VerifyTLS bool           `gorm:"default:false" json:"verify_tls"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserSession 用户会话表
type UserSession struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	UserID     string         `gorm:"size:255;not null;index" json:"user_id"`
	Username   string         `gorm:"size:255;not null" json:"username"`
	Host       string         `gorm:"size:255;not null" json:"host"`
	Port       int            `gorm:"not null" json:"port"`
	Token      string         `gorm:"uniqueIndex;size:1000;not null" json:"-"`
	IP         string         `gorm:"size:45" json:"ip"`
	UserAgent  string         `gorm:"size:500" json:"user_agent"`
	ExpiresAt  time.Time      `gorm:"not null" json:"expires_at"`
	LastActive time.Time      `json:"last_active"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// AuditLog 审计日志表
type AuditLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    string    `gorm:"size:255;not null;index" json:"user_id"`
	Username  string    `gorm:"size:255;not null;json:" json:"username"`
	Action    string    `gorm:"size:100;not null;index" json:"action"`
	Resource  string    `gorm:"size:255" json:"resource"`
	Detail    string    `gorm:"type:text" json:"detail"`
	IP        string    `gorm:"size:45" json:"ip"`
	Status    string    `gorm:"size:50;default:success" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

// SystemConfig 系统配置表
type SystemConfig struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Key       string    `gorm:"uniqueIndex;size:100;not null" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	Remark    string    `gorm:"size:500" json:"remark"`
	UpdatedAt time.Time `json:"updated_at"`
}
