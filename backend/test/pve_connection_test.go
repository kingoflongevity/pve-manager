package pve_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kingoflongevity/pve-manager/backend/internal/config"
	"github.com/kingoflongevity/pve-manager/backend/internal/client/pve"
	"go.uber.org/zap"
)

// TestPVEConnection 测试本地 PVE 连接
// 使用 root@pam / Qq1212112 凭据验证接口正确性
func TestPVEConnection(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg := config.PVEConfig{
		BaseURL:   "https://192.168.1.10:8006/api2/json",
		Username:  "root",
		Password:  "Qq1212112",
		Realm:     "pam",
		VerifyTLS: false,
	}

	client, err := pve.NewClient(cfg, logger)
	if err != nil {
		t.Fatalf("创建 PVE 客户端失败: %v", err)
	}

	ctx := context.Background()

	t.Run("登录测试", func(t *testing.T) {
		resp, err := client.Login(ctx, "root", "Qq1212112", "pam")
		if err != nil {
			t.Fatalf("登录失败: %v", err)
		}
		if resp == nil || resp.Ticket == "" {
			t.Fatal("登录成功但未获取到 ticket")
		}
		logger.Info("PVE 登录成功", zap.String("username", "root@pam"))
	})

	t.Run("获取集群资源", func(t *testing.T) {
		resources, err := client.GetClusterResources(ctx)
		if err != nil {
			t.Fatalf("获取集群资源失败: %v", err)
		}
		logger.Info("集群资源", zap.Int("count", len(resources)))
		for _, r := range resources {
			logger.Info("资源",
				zap.String("type", r.Type),
				zap.String("id", r.ID),
				zap.String("status", r.Status),
			)
		}
	})

	t.Run("获取节点状态", func(t *testing.T) {
		resources, err := client.GetClusterResources(ctx)
		if err != nil {
			t.Fatalf("获取集群资源失败: %v", err)
		}
		for _, r := range resources {
			if r.Type == "node" {
				status, err := client.GetNodeStatus(ctx, r.Node)
				if err != nil {
					t.Errorf("获取节点 %s 状态失败: %v", r.Node, err)
					continue
				}
				logger.Info("节点状态",
					zap.String("node", r.Node),
					zap.Float64("cpu", status.CPU),
					zap.Uint64("mem_used", status.Mem),
					zap.Uint64("mem_total", status.MaxMem),
				)
			}
		}
	})

	t.Run("获取节点版本", func(t *testing.T) {
		resources, err := client.GetClusterResources(ctx)
		if err != nil {
			t.Fatalf("获取集群资源失败: %v", err)
		}
		for _, r := range resources {
			if r.Type == "node" {
				version, err := client.GetNodeVersion(ctx, r.Node)
				if err != nil {
					t.Errorf("获取节点 %s 版本失败: %v", r.Node, err)
					continue
				}
				logger.Info("节点版本",
					zap.String("node", r.Node),
					zap.String("version", version.Version),
					zap.String("release", version.Release),
				)
			}
		}
	})

	t.Run("获取QEMU虚拟机列表", func(t *testing.T) {
		resources, err := client.GetClusterResources(ctx)
		if err != nil {
			t.Fatalf("获取集群资源失败: %v", err)
		}
		for _, r := range resources {
			if r.Type == "node" {
				vms, err := client.ListQEMU(ctx, r.Node)
				if err != nil {
					t.Errorf("获取节点 %s QEMU 列表失败: %v", r.Node, err)
					continue
				}
				logger.Info("QEMU 虚拟机",
					zap.String("node", r.Node),
					zap.Int("count", len(vms)),
				)
				for _, vm := range vms {
					logger.Info("VM",
						zap.Int("vmid", vm.VMID),
						zap.String("name", vm.Name),
						zap.String("status", vm.Status),
					)
				}
			}
		}
	})

	t.Run("获取LXC容器列表", func(t *testing.T) {
		resources, err := client.GetClusterResources(ctx)
		if err != nil {
			t.Fatalf("获取集群资源失败: %v", err)
		}
		for _, r := range resources {
			if r.Type == "node" {
				containers, err := client.ListLXC(ctx, r.Node)
				if err != nil {
					t.Errorf("获取节点 %s LXC 列表失败: %v", r.Node, err)
					continue
				}
				logger.Info("LXC 容器",
					zap.String("node", r.Node),
					zap.Int("count", len(containers)),
				)
				for _, ct := range containers {
					logger.Info("CT",
						zap.Int("vmid", ct.VMID),
						zap.String("name", ct.Name),
						zap.String("status", ct.Status),
					)
				}
			}
		}
	})

	t.Run("获取存储列表", func(t *testing.T) {
		resources, err := client.GetClusterResources(ctx)
		if err != nil {
			t.Fatalf("获取集群资源失败: %v", err)
		}
		for _, r := range resources {
			if r.Type == "node" {
				storages, err := client.ListStorage(ctx, r.Node)
				if err != nil {
					t.Errorf("获取节点 %s 存储列表失败: %v", r.Node, err)
					continue
				}
				logger.Info("存储",
					zap.String("node", r.Node),
					zap.Int("count", len(storages)),
				)
				for _, s := range storages {
					logger.Info("Storage",
						zap.String("storage", s.Storage),
						zap.String("type", s.Type),
						zap.Int("enabled", s.Enabled),
					)
				}
			}
		}
	})
}

func main() {
	fmt.Println("使用 go test -v ./test/pve_connection_test.go 运行测试")
}
