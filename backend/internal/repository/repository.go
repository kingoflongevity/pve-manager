package repository

import (
	"time"

	"github.com/kingoflongevity/pve-manager/backend/internal/database"
	"github.com/kingoflongevity/pve-manager/backend/internal/model"
	"gorm.io/gorm"
)

// PVEConfigRepo PVE 配置仓库
type PVEConfigRepo struct {
	db *gorm.DB
}

func NewPVEConfigRepo() *PVEConfigRepo {
	return &PVEConfigRepo{db: database.GetDB()}
}

func (r *PVEConfigRepo) Create(cfg *model.PVEConfig) error {
	return r.db.Create(cfg).Error
}

func (r *PVEConfigRepo) GetByHost(host string, port int) (*model.PVEConfig, error) {
	var cfg model.PVEConfig
	err := r.db.Where("host = ? AND port = ?", host, port).First(&cfg).Error
	return &cfg, err
}

func (r *PVEConfigRepo) GetByID(id uint) (*model.PVEConfig, error) {
	var cfg model.PVEConfig
	err := r.db.First(&cfg, id).Error
	return &cfg, err
}

func (r *PVEConfigRepo) GetAll() ([]model.PVEConfig, error) {
	var configs []model.PVEConfig
	err := r.db.Order("created_at DESC").Find(&configs).Error
	return configs, err
}

func (r *PVEConfigRepo) Update(cfg *model.PVEConfig) error {
	return r.db.Save(cfg).Error
}

func (r *PVEConfigRepo) Delete(id uint) error {
	return r.db.Delete(&model.PVEConfig{}, id).Error
}

func (r *PVEConfigRepo) SetDefault(id uint) error {
	return r.db.Model(&model.PVEConfig{}).Updates(map[string]interface{}{
		"is_default": gorm.Expr("CASE WHEN id = ? THEN 1 ELSE 0 END", id),
	}).Error
}

// SessionRepo 用户会话仓库
type SessionRepo struct {
	db *gorm.DB
}

func NewSessionRepo() *SessionRepo {
	return &SessionRepo{db: database.GetDB()}
}

func (r *SessionRepo) Create(session *model.UserSession) error {
	return r.db.Create(session).Error
}

func (r *SessionRepo) GetByToken(token string) (*model.UserSession, error) {
	var session model.UserSession
	err := r.db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&session).Error
	return &session, err
}

func (r *SessionRepo) UpdateLastActive(token string) error {
	return r.db.Model(&model.UserSession{}).
		Where("token = ?", token).
		Update("last_active", time.Now()).Error
}

func (r *SessionRepo) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&model.UserSession{}).Error
}

func (r *SessionRepo) DeleteByToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&model.UserSession{}).Error
}

func (r *SessionRepo) GetUserSessions(userID string) ([]model.UserSession, error) {
	var sessions []model.UserSession
	err := r.db.Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").Find(&sessions).Error
	return sessions, err
}

// AuditLogRepo 审计日志仓库
type AuditLogRepo struct {
	db *gorm.DB
}

func NewAuditLogRepo() *AuditLogRepo {
	return &AuditLogRepo{db: database.GetDB()}
}

func (r *AuditLogRepo) Create(log *model.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *AuditLogRepo) List(page, pageSize int, userID, action string) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	query := r.db.Model(&model.AuditLog{})
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

func (r *AuditLogRepo) GetRecent(limit int) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	err := r.db.Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

// SystemConfigRepo 系统配置仓库
type SystemConfigRepo struct {
	db *gorm.DB
}

func NewSystemConfigRepo() *SystemConfigRepo {
	return &SystemConfigRepo{db: database.GetDB()}
}

func (r *SystemConfigRepo) Get(key string) (string, error) {
	var cfg model.SystemConfig
	err := r.db.Where("key = ?", key).First(&cfg).Error
	return cfg.Value, err
}

func (r *SystemConfigRepo) Set(key, value, remark string) error {
	cfg := model.SystemConfig{Key: key, Value: value, Remark: remark}
	return r.db.Where("key = ?", key).Assign(cfg).FirstOrCreate(&cfg).Error
}

func (r *SystemConfigRepo) GetAll() (map[string]string, error) {
	var configs []model.SystemConfig
	if err := r.db.Find(&configs).Error; err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, c := range configs {
		result[c.Key] = c.Value
	}
	return result, nil
}

func (r *SystemConfigRepo) Delete(key string) error {
	return r.db.Where("key = ?", key).Delete(&model.SystemConfig{}).Error
}
