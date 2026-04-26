package service

import (
	"context"
	"fmt"

	"github.com/kingoflongevity/pve-manager/backend/internal/client/pve"
	"go.uber.org/zap"
)

// VMService QEMU 虚拟机服务
// 负责虚拟机的创建、配置、操作、快照、克隆、迁移等
type VMService struct {
	logger *zap.Logger
}

// NewVMService 创建虚拟机服务实例
func NewVMService(logger *zap.Logger) *VMService {
	return &VMService{logger: logger}
}

// ListVMs 获取节点上所有 QEMU 虚拟机列表
func (s *VMService) ListVMs(ctx context.Context, client *pve.Client, node string) ([]pve.QEMUVM, error) {
	return client.ListQEMU(ctx, node)
}

// GetVMConfig 获取虚拟机完整配置
func (s *VMService) GetVMConfig(ctx context.Context, client *pve.Client, node string, vmid int) (map[string]interface{}, error) {
	return client.GetQEMUConfig(ctx, node, vmid)
}

// CreateVM 创建虚拟机
func (s *VMService) CreateVM(ctx context.Context, client *pve.Client, node string, params *pve.QEMUCreateParams) (string, error) {
	return client.CreateQEMU(ctx, node, params)
}

// VMAction 执行虚拟机操作 (start, stop, shutdown, reboot, suspend, resume, reset)
func (s *VMService) VMAction(ctx context.Context, client *pve.Client, node string, vmid int, action string) (string, error) {
	switch action {
	case "start":
		return client.StartQEMU(ctx, node, vmid)
	case "stop":
		return client.StopQEMU(ctx, node, vmid)
	case "shutdown":
		return client.ShutdownQEMU(ctx, node, vmid)
	case "reboot":
		return client.RebootQEMU(ctx, node, vmid)
	case "suspend":
		return client.SuspendQEMU(ctx, node, vmid)
	case "resume":
		return client.ResumeQEMU(ctx, node, vmid)
	case "reset":
		return client.ResetQEMU(ctx, node, vmid)
	default:
		return "", nil
	}
}

// SetVMConfig 更新虚拟机配置
func (s *VMService) SetVMConfig(ctx context.Context, client *pve.Client, node string, vmid int, config pve.QEMUConfigParams) (string, error) {
	return client.SetQEMUConfig(ctx, node, vmid, config)
}

// DeleteVM 删除虚拟机
func (s *VMService) DeleteVM(ctx context.Context, client *pve.Client, node string, vmid int) (string, error) {
	return client.DeleteQEMU(ctx, node, vmid)
}

// GetSnapshots 获取虚拟机快照列表
func (s *VMService) GetSnapshots(ctx context.Context, client *pve.Client, node string, vmid int) ([]pve.Snapshot, error) {
	return client.ListQEMUSnapshots(ctx, node, vmid)
}

// CreateSnapshot 创建虚拟机快照
func (s *VMService) CreateSnapshot(ctx context.Context, client *pve.Client, node string, vmid int, name, description string) (string, error) {
	return client.CreateQEMUSnapshot(ctx, node, vmid, name, description)
}

// DeleteSnapshot 删除虚拟机快照
func (s *VMService) DeleteSnapshot(ctx context.Context, client *pve.Client, node string, vmid int, snapname string) (string, error) {
	return client.DeleteQEMUSnapshot(ctx, node, vmid, snapname)
}

// CloneVM 克隆虚拟机
func (s *VMService) CloneVM(ctx context.Context, client *pve.Client, node string, vmid int, params *pve.QEMUCloneParams) (string, error) {
	return client.CloneQEMU(ctx, node, vmid, params)
}

// MigrateVM 迁移虚拟机
func (s *VMService) MigrateVM(ctx context.Context, client *pve.Client, node string, vmid int, params *pve.QEMUMigrateParams) (string, error) {
	return client.MigrateQEMU(ctx, node, vmid, params)
}

// GetRRD 获取虚拟机 RRD 性能数据
func (s *VMService) GetRRD(ctx context.Context, client *pve.Client, node string, vmid int, timeframe, dataset string) ([]pve.RRDPoint, error) {
	return client.GetQEMURRD(ctx, node, vmid, timeframe, dataset)
}

// GetPending 获取虚拟机待处理配置
func (s *VMService) GetPending(ctx context.Context, client *pve.Client, node string, vmid int) ([]pve.PendingConfig, error) {
	return client.GetQEMUPending(ctx, node, vmid)
}

// GetCurrent 获取虚拟机当前状态
func (s *VMService) GetCurrent(ctx context.Context, client *pve.Client, node string, vmid int) (*pve.QEMUVM, error) {
	return client.GetQEMUCurrent(ctx, node, vmid)
}

// VNCProxy 获取 VNC 代理票据
func (s *VMService) VNCProxy(ctx context.Context, client *pve.Client, node string, vmid int) (interface{}, error) {
	path := fmt.Sprintf("nodes/%s/qemu/%d/vncproxy", node, vmid)
	resp, err := client.Do(ctx, "POST", path, map[string]interface{}{
		"websocket": 1,
	})
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
