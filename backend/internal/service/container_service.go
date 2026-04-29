package service

import (
	"context"

	"github.com/kingoflongevity/pve-manager/backend/internal/client/pve"
	"go.uber.org/zap"
)

// ContainerService LXC 容器服务
// 负责容器的创建、配置、操作、快照、克隆、迁移等
type ContainerService struct {
	logger *zap.Logger
}

// NewContainerService 创建容器服务实例
func NewContainerService(logger *zap.Logger) *ContainerService {
	return &ContainerService{logger: logger}
}

// ListContainers 获取节点上所有 LXC 容器列表
func (s *ContainerService) ListContainers(ctx context.Context, client *pve.Client, node string) (interface{}, error) {
	return client.ListLXC(ctx, node)
}

// GetContainerConfig 获取容器完整配置
func (s *ContainerService) GetContainerConfig(ctx context.Context, client *pve.Client, node string, vmid int) (map[string]interface{}, error) {
	return client.GetLXCConfig(ctx, node, vmid)
}

// CreateContainer 创建容器
func (s *ContainerService) CreateContainer(ctx context.Context, client *pve.Client, node string, params *pve.LXCCreateParams) (string, error) {
	return client.CreateLXC(ctx, node, params)
}

// ContainerAction 执行容器操作 (start, stop, shutdown, reboot, freeze, unfreeze)
func (s *ContainerService) ContainerAction(ctx context.Context, client *pve.Client, node string, vmid int, action string) (string, error) {
	switch action {
	case "start":
		return client.StartLXC(ctx, node, vmid)
	case "stop":
		return client.StopLXC(ctx, node, vmid)
	case "shutdown":
		return client.ShutdownLXC(ctx, node, vmid)
	case "reboot":
		return client.RebootLXC(ctx, node, vmid)
	case "freeze":
		return client.FreezeLXC(ctx, node, vmid)
	case "unfreeze":
		return client.UnfreezeLXC(ctx, node, vmid)
	default:
		return "", nil
	}
}

// SetContainerConfig 更新容器配置
func (s *ContainerService) SetContainerConfig(ctx context.Context, client *pve.Client, node string, vmid int, config pve.LXCConfigParams) (string, error) {
	return client.SetLXCConfig(ctx, node, vmid, config)
}

// DeleteContainer 删除容器
func (s *ContainerService) DeleteContainer(ctx context.Context, client *pve.Client, node string, vmid int) (string, error) {
	return client.DeleteLXC(ctx, node, vmid)
}

// GetSnapshots 获取容器快照列表
func (s *ContainerService) GetSnapshots(ctx context.Context, client *pve.Client, node string, vmid int) (interface{}, error) {
	return client.ListLXCSnapshots(ctx, node, vmid)
}

// CreateSnapshot 创建容器快照
func (s *ContainerService) CreateSnapshot(ctx context.Context, client *pve.Client, node string, vmid int, name, description string) (string, error) {
	return client.CreateLXCSnapshot(ctx, node, vmid, name, description)
}

// DeleteSnapshot 删除容器快照
func (s *ContainerService) DeleteSnapshot(ctx context.Context, client *pve.Client, node string, vmid int, snapname string) (string, error) {
	return client.DeleteLXCSnapshot(ctx, node, vmid, snapname)
}

// CloneContainer 克隆容器
func (s *ContainerService) CloneContainer(ctx context.Context, client *pve.Client, node string, vmid int, params *pve.LXCCloneParams) (string, error) {
	return client.CloneLXC(ctx, node, vmid, params)
}

// MigrateContainer 迁移容器
func (s *ContainerService) MigrateContainer(ctx context.Context, client *pve.Client, node string, vmid int, params *pve.LXCMigrateParams) (string, error) {
	return client.MigrateLXC(ctx, node, vmid, params)
}

// GetRRD 获取容器 RRD 性能数据
func (s *ContainerService) GetRRD(ctx context.Context, client *pve.Client, node string, vmid int, timeframe, dataset string) (interface{}, error) {
	return client.GetLXCURRD(ctx, node, vmid, timeframe, dataset)
}

// GetPending 获取容器待处理配置
func (s *ContainerService) GetPending(ctx context.Context, client *pve.Client, node string, vmid int) (interface{}, error) {
	return client.ListLXCPending(ctx, node, vmid)
}

// GetCurrent 获取容器当前状态
func (s *ContainerService) GetCurrent(ctx context.Context, client *pve.Client, node string, vmid int) (interface{}, error) {
	return client.GetLXCCurrent(ctx, node, vmid)
}
