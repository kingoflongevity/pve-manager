package service

import (
	"context"

	"github.com/kingoflongevity/pve-manager/backend/internal/client/pve"
	"go.uber.org/zap"
)

// StorageService 存储服务
// 负责存储列表、状态、内容、ISO 下载等存储级操作
type StorageService struct {
	logger *zap.Logger
}

// NewStorageService 创建存储服务实例
func NewStorageService(logger *zap.Logger) *StorageService {
	return &StorageService{logger: logger}
}

// ListStorage 获取节点上所有存储列表
func (s *StorageService) ListStorage(ctx context.Context, client *pve.Client, node string) ([]pve.Storage, error) {
	return client.ListStorage(ctx, node)
}

// GetStorageStatus 获取指定存储的状态
func (s *StorageService) GetStorageStatus(ctx context.Context, client *pve.Client, node, storage string) (*pve.StorageStatus, error) {
	return client.GetStorageStatus(ctx, node, storage)
}

// GetStorageContent 获取存储内容列表
func (s *StorageService) GetStorageContent(ctx context.Context, client *pve.Client, node, storage, contentType string) ([]pve.StorageContent, error) {
	if contentType != "" {
		return client.GetStorageContentByType(ctx, node, storage, contentType)
	}
	return client.GetStorageContent(ctx, node, storage)
}

// DownloadISO 从 URL 下载 ISO 到存储
func (s *StorageService) DownloadISO(ctx context.Context, client *pve.Client, node, storage, fileURL, filename string) (string, error) {
	return client.DownloadISO(ctx, node, storage, fileURL, filename)
}

// CreateStorage 创建存储
func (s *StorageService) CreateStorage(ctx context.Context, client *pve.Client, node string, params pve.StorageCreateParams) (string, error) {
	return client.CreateStorage(ctx, node, params)
}

// UpdateStorage 更新存储配置
func (s *StorageService) UpdateStorage(ctx context.Context, client *pve.Client, node, storage string, params pve.StorageUpdateParams) (string, error) {
	return client.UpdateStorage(ctx, node, storage, params)
}

// DeleteStorage 删除存储
func (s *StorageService) DeleteStorage(ctx context.Context, client *pve.Client, node, storage string) (string, error) {
	return client.DeleteStorage(ctx, node, storage)
}
