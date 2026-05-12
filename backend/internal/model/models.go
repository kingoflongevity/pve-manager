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

// AIModelConfig AI 模型配置表
type AIModelConfig struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	Name         string         `gorm:"size:100;not null" json:"name"`
	Provider     string         `gorm:"size:50;not null;index" json:"provider"`
	BaseURL      string         `gorm:"size:500;not null" json:"base_url"`
	APIKey       string         `gorm:"size:1000;not null" json:"-"`
	Model        string         `gorm:"size:100;not null" json:"model"`
	MaxTokens    int            `gorm:"default:4096" json:"max_tokens"`
	Temperature  float64        `gorm:"default:0.7" json:"temperature"`
	Timeout      int            `gorm:"default:60" json:"timeout"`
	IsEnabled    bool           `gorm:"default:true" json:"is_enabled"`
	IsDefault    bool           `gorm:"default:false" json:"is_default"`
	SortOrder    int            `gorm:"default:0" json:"sort_order"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// AIConversation AI 对话会话表
type AIConversation struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Title       string         `gorm:"size:255" json:"title"`
	Scene       string         `gorm:"size:50;not null;index" json:"scene"`
	ModelConfigID uint          `gorm:"index" json:"model_config_id"`
	UserID      string         `gorm:"size:255;not null;index" json:"user_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Messages []AIMessage `gorm:"foreignKey:ConversationID;constraint:OnDelete:CASCADE" json:"messages,omitempty"`
}

// AIMessage AI 对话消息表
type AIMessage struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	ConversationID uint           `gorm:"not null;index" json:"conversation_id"`
	Role           string         `gorm:"size:20;not null" json:"role"`
	Content        string         `gorm:"type:text;not null" json:"content"`
	ToolCalls      string         `gorm:"type:text" json:"tool_calls"`
	ToolCallID     string         `gorm:"size:100" json:"tool_call_id"`
	CreatedAt      time.Time      `json:"created_at"`
}

// AIReport AI 生成报告表
type AIReport struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	Title         string         `gorm:"size:255;not null" json:"title"`
	Type          string         `gorm:"size:50;not null;index" json:"type"`
	Content       string         `gorm:"type:text;not null" json:"content"`
	ModelConfigID uint           `json:"model_config_id"`
	UserID        string         `gorm:"size:255;not null;index" json:"user_id"`
	ScheduleID    uint           `gorm:"index" json:"schedule_id"`
	CreatedAt     time.Time      `json:"created_at"`
}

// ReportSchedule 报告定时任务表
type ReportSchedule struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	Name          string         `gorm:"size:100;not null" json:"name"`
	Type          string         `gorm:"size:50;not null" json:"type"`
	Schedule      string         `gorm:"size:50;not null" json:"schedule"`
	ModelConfigID uint           `json:"model_config_id"`
	Recipients    string         `gorm:"type:text" json:"recipients"`
	IsEnabled     bool           `gorm:"default:true" json:"is_enabled"`
	LastRunAt     time.Time      `json:"last_run_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// AppTemplate 应用商店模板表
type AppTemplate struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"size:100;not null;index" json:"name"`
	Category    string         `gorm:"size:50;not null;index" json:"category"`
	Description string         `gorm:"type:text" json:"description"`
	Icon        string         `gorm:"size:500" json:"icon"`
	Version     string         `gorm:"size:20;not null" json:"version"`
	Author      string         `gorm:"size:100" json:"author"`
	MinCPU      int            `gorm:"default:1" json:"min_cpu"`
	MinMemoryMB int            `gorm:"default:512" json:"min_memory_mb"`
	MinDiskGB   int            `gorm:"default:8" json:"min_disk_gb"`
	Type        string         `gorm:"size:10;default:lxc" json:"type"`
	OSTemplate  string         `gorm:"size:255" json:"os_template"`
	Packages    string         `gorm:"type:text" json:"packages"`
	Variables   string         `gorm:"type:text" json:"variables"`
	SetupSteps  string         `gorm:"type:text" json:"setup_steps"`
	Template    string         `gorm:"type:text" json:"-"`
	IsBuiltIn   bool           `gorm:"default:false" json:"is_built_in"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// AppDeployment 应用部署记录表
type AppDeployment struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	TemplateID  uint           `gorm:"not null;index" json:"template_id"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Node        string         `gorm:"size:100;not null" json:"node"`
	Type        string         `gorm:"size:10" json:"type"`
	VMID        int            `gorm:"column:vmid;index" json:"vmid"`
	Status      string         `gorm:"size:20;not null;index" json:"status"`
	Progress    int            `gorm:"default:0" json:"progress"`
	StepInfo    string         `gorm:"size:255" json:"step_info"`
	Config      string         `gorm:"type:text" json:"config"`
	ErrorMsg    string         `gorm:"type:text" json:"error_msg"`
	UserID      string         `gorm:"size:255;not null" json:"user_id"`
	StartedAt   *time.Time     `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
