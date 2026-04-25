package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release, test
}

// PVEConfig Proxmox VE API 配置
type PVEConfig struct {
	BaseURL  string `mapstructure:"base_url"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Realm    string `mapstructure:"realm"` // pam, pve, etc.
	VerifyTLS bool  `mapstructure:"verify_tls"`
}

// CORSConfig CORS 跨域配置
type CORSConfig struct {
	AllowOrigins []string `mapstructure:"allow_origins"`
	AllowMethods []string `mapstructure:"allow_methods"`
	AllowHeaders []string `mapstructure:"allow_headers"`
}

// AESConfig AES 加密配置
type AESConfig struct {
	Key string `mapstructure:"key"` // AES-256 需要 32 字节密钥
}

// Config 全局应用配置
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	PVE    PVEConfig    `mapstructure:"pve"`
	CORS   CORSConfig   `mapstructure:"cors"`
	AES    AESConfig    `mapstructure:"aes"`
}

// LoadConfig 加载配置文件和环境变量
// 优先读取 config.yaml，未设置的值通过环境变量覆盖
// 环境变量前缀: PVE_MANAGER_
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// 设置配置文件路径
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 读取配置文件（如果存在）
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	// 配置环境变量
	v.SetEnvPrefix("PVE_MANAGER")
	v.AutomaticEnv()

	// 设置默认值
	setDefaults(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 验证必要配置
	if cfg.PVE.BaseURL == "" {
		return nil, fmt.Errorf("PVE BaseURL 不能为空")
	}

	return &cfg, nil
}

// setDefaults 设置默认配置值
func setDefaults(v *viper.Viper) {
	// Server 默认值
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "release")

	// PVE 默认值
	v.SetDefault("pve.verify_tls", true)
	v.SetDefault("pve.realm", "pam")

	// CORS 默认值
	v.SetDefault("cors.allow_origins", []string{"*"})
	v.SetDefault("cors.allow_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allow_headers", []string{"Origin", "Content-Type", "Accept", "Authorization"})
}

// GenerateDefaultConfig 生成默认配置文件到指定路径
// 用于首次启动时创建配置模板
func GenerateDefaultConfig(path string) error {
	defaultConfig := `server:
  port: 8080
  mode: debug # debug, release, test

pve:
  base_url: "https://192.168.1.100:8006/api2/json"
  username: "root"
  password: ""
  realm: "pam" # pam, pve, ldap
  verify_tls: false

cors:
  allow_origins:
    - "http://localhost:3000"
    - "http://localhost:5173"
  allow_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allow_headers:
    - "Origin"
    - "Content-Type"
    - "Accept"
    - "Authorization"

aes:
  key: "" # 留空则自动生成
`

	return os.WriteFile(path, []byte(defaultConfig), 0644)
}
