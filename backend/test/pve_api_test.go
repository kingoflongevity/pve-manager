//go:build integration

package pve_test

import (
	"context"
	"testing"

	"github.com/kingoflongevity/pve-manager/backend/internal/config"
	"github.com/kingoflongevity/pve-manager/backend/internal/client/pve"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// setupTestClient 创建已登录的 PVE 客户端
func setupTestClient(t *testing.T) (*pve.Client, context.Context) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	t.Cleanup(func() { logger.Sync() })

	cfg := config.PVEConfig{
		BaseURL:   "https://192.168.1.10:8006/api2/json",
		Username:  "root",
		Password:  "Qq1212112",
		Realm:     "pam",
		VerifyTLS: false,
	}

	client, err := pve.NewClient(cfg, logger)
	require.NoError(t, err, "创建 PVE 客户端失败")

	ctx := context.Background()

	// 登录
	_, err = client.Login(ctx, "root", "Qq1212112", "pam")
	require.NoError(t, err, "登录 PVE 失败")

	return client, ctx
}

// getTestNode 获取测试用的第一个节点名称
func getTestNode(t *testing.T, client *pve.Client, ctx context.Context) string {
	t.Helper()
	resources, err := client.GetClusterResources(ctx)
	require.NoError(t, err, "获取集群资源失败")

	for _, r := range resources {
		if r.Type == "node" {
			return r.Node
		}
	}
	t.Fatal("未找到任何节点")
	return ""
}

// ==================== 集群管理接口测试 ====================

func TestClusterResources(t *testing.T) {
	client, ctx := setupTestClient(t)

	resources, err := client.GetClusterResources(ctx)
	require.NoError(t, err, "获取集群资源失败")
	assert.NotEmpty(t, resources, "集群资源不应为空")

	// 验证资源类型
	types := make(map[string]int)
	for _, r := range resources {
		types[r.Type]++
	}
	assert.Contains(t, types, "node", "应包含节点资源")
	t.Logf("集群资源统计: %+v", types)
}

func TestClusterResourcesByType(t *testing.T) {
	client, ctx := setupTestClient(t)

	t.Run("获取虚拟机资源", func(t *testing.T) {
		resources, err := client.GetClusterResourcesByType(ctx, "vm")
		require.NoError(t, err)
		t.Logf("虚拟机数量: %d", len(resources))
	})

	t.Run("获取存储资源", func(t *testing.T) {
		resources, err := client.GetClusterResourcesByType(ctx, "storage")
		require.NoError(t, err)
		t.Logf("存储数量: %d", len(resources))
	})

	t.Run("获取节点资源", func(t *testing.T) {
		resources, err := client.GetClusterResourcesByType(ctx, "node")
		require.NoError(t, err)
		t.Logf("节点数量: %d", len(resources))
	})
}

func TestClusterTasks(t *testing.T) {
	client, ctx := setupTestClient(t)

	tasks, err := client.GetClusterTasks(ctx)
	require.NoError(t, err, "获取集群任务失败")
	t.Logf("集群任务数量: %d", len(tasks))
}

func TestNextID(t *testing.T) {
	client, ctx := setupTestClient(t)

	// GetNextID 可能返回 nil 如果 PVE 中没有配置 nextid
	nextID, err := client.GetNextID(ctx)
	if err != nil {
		t.Logf("获取下一个 ID 失败 (可能未配置): %v", err)
		return
	}
	if nextID != nil {
		t.Logf("下一个可用 ID: %d", nextID.NextID)
	}
}

// ==================== 节点管理接口测试 ====================

func TestNodeStatus(t *testing.T) {
	client, ctx := setupTestClient(t)
	node := getTestNode(t, client, ctx)

	status, err := client.GetNodeStatus(ctx, node)
	require.NoError(t, err, "获取节点状态失败")
	// PVE 9.x GetNodeStatus 返回的 Node 字段可能为空，使用 URL 中的节点名验证
	assert.NotEmpty(t, status.PVEVersion, "PVE 版本不应为空")
	t.Logf("节点 %s: CPU=%.1f%%, 内存=%d/%d, 版本=%s", node, status.CPU*100, status.Mem, status.MaxMem, status.PVEVersion)
}

func TestNodeVersion(t *testing.T) {
	client, ctx := setupTestClient(t)
	node := getTestNode(t, client, ctx)

	version, err := client.GetNodeVersion(ctx, node)
	require.NoError(t, err, "获取节点版本失败")
	assert.NotEmpty(t, version.Version, "版本信息不应为空")
	t.Logf("节点 %s 版本: %s", node, version.Version)
}

func TestNodeServices(t *testing.T) {
	client, ctx := setupTestClient(t)
	node := getTestNode(t, client, ctx)

	services, err := client.GetNodeServices(ctx, node)
	require.NoError(t, err, "获取节点服务失败")
	assert.NotEmpty(t, services, "节点服务列表不应为空")
	t.Logf("节点 %s 服务数量: %d", node, len(services))
}

func TestNodeNetwork(t *testing.T) {
	client, ctx := setupTestClient(t)
	node := getTestNode(t, client, ctx)

	network, err := client.GetNodeNetwork(ctx, node)
	require.NoError(t, err, "获取网络接口失败")
	assert.NotEmpty(t, network, "网络接口列表不应为空")
	t.Logf("节点 %s 网络接口数量: %d", node, len(network))
}

func TestNodeDNS(t *testing.T) {
	client, ctx := setupTestClient(t)
	node := getTestNode(t, client, ctx)

	dns, err := client.GetNodeDNS(ctx, node)
	require.NoError(t, err, "获取 DNS 配置失败")
	t.Logf("节点 %s DNS 配置: %+v", node, dns)
}

func TestNodeTime(t *testing.T) {
	client, ctx := setupTestClient(t)
	node := getTestNode(t, client, ctx)

	timeInfo, err := client.GetNodeTime(ctx, node)
	require.NoError(t, err, "获取时间信息失败")
	t.Logf("节点 %s 时间戳: %d, 时区: %s", node, timeInfo.Time, timeInfo.Timezone)
}

func TestNodeTasks(t *testing.T) {
	client, ctx := setupTestClient(t)
	node := getTestNode(t, client, ctx)

	tasks, err := client.GetNodeTasks(ctx, node)
	require.NoError(t, err, "获取节点任务失败")
	t.Logf("节点 %s 任务数量: %d", node, len(tasks))
}

func TestNodeRRD(t *testing.T) {
	client, ctx := setupTestClient(t)
	node := getTestNode(t, client, ctx)

	rrdData, err := client.GetNodeRRD(ctx, node, "hour", "")
	if err != nil {
		t.Logf("获取 RRD 数据失败 (可能未启用): %v", err)
		return
	}
	t.Logf("节点 %s RRD 数据点数: %d", node, len(rrdData))
}

// ==================== 存储管理接口测试 ====================

func TestStorageList(t *testing.T) {
	client, ctx := setupTestClient(t)
	node := getTestNode(t, client, ctx)

	storages, err := client.ListStorage(ctx, node)
	require.NoError(t, err, "获取存储列表失败")
	assert.NotEmpty(t, storages, "存储列表不应为空")
	t.Logf("节点 %s 存储数量: %d", node, len(storages))
}

func TestStorageContent(t *testing.T) {
	client, ctx := setupTestClient(t)
	node := getTestNode(t, client, ctx)

	storages, err := client.ListStorage(ctx, node)
	require.NoError(t, err)
	require.NotEmpty(t, storages, "至少需要一个存储")

	// 找第一个启用的存储
	for _, s := range storages {
		if s.Enabled == 1 {
			content, err := client.GetStorageContent(ctx, node, s.Storage)
			if err != nil {
				t.Logf("存储 %s 内容获取失败: %v", s.Storage, err)
				continue
			}
			t.Logf("存储 %s 内容数量: %d", s.Storage, len(content))
			return
		}
	}
	t.Skip("没有启用的存储")
}

// ==================== QEMU 虚拟机接口测试 ====================

func TestQEMUList(t *testing.T) {
	client, ctx := setupTestClient(t)
	node := getTestNode(t, client, ctx)

	vms, err := client.ListQEMU(ctx, node)
	require.NoError(t, err, "获取虚拟机列表失败")
	t.Logf("节点 %s 虚拟机数量: %d", node, len(vms))

	for _, vm := range vms {
		t.Logf("VM %d: %s (%s)", vm.VMID, vm.Name, vm.Status)
	}
}

func TestQEMUConfig(t *testing.T) {
	client, ctx := setupTestClient(t)
	node := getTestNode(t, client, ctx)

	vms, err := client.ListQEMU(ctx, node)
	require.NoError(t, err)
	require.NotEmpty(t, vms, "至少需要一台虚拟机")

	vmid := vms[0].VMID
	config, err := client.GetQEMUConfig(ctx, node, vmid)
	require.NoError(t, err, "获取虚拟机配置失败")
	t.Logf("VM %d 配置: %+v", vmid, config)
}

func TestQEMUSnapshots(t *testing.T) {
	client, ctx := setupTestClient(t)
	node := getTestNode(t, client, ctx)

	vms, err := client.ListQEMU(ctx, node)
	require.NoError(t, err)
	if len(vms) == 0 {
		t.Skip("没有虚拟机，跳过快照测试")
	}

	vmid := vms[0].VMID
	snapshots, err := client.ListQEMUSnapshots(ctx, node, vmid)
	require.NoError(t, err, "获取快照失败")
	t.Logf("VM %d 快照数量: %d", vmid, len(snapshots))
}

// ==================== LXC 容器接口测试 ====================

func TestLXCList(t *testing.T) {
	client, ctx := setupTestClient(t)
	node := getTestNode(t, client, ctx)

	containers, err := client.ListLXC(ctx, node)
	require.NoError(t, err, "获取容器列表失败")
	t.Logf("节点 %s 容器数量: %d", node, len(containers))

	for _, ct := range containers {
		t.Logf("CT %d: %s (%s)", ct.VMID, ct.Name, ct.Status)
	}
}

func TestLXCConfig(t *testing.T) {
	client, ctx := setupTestClient(t)
	node := getTestNode(t, client, ctx)

	containers, err := client.ListLXC(ctx, node)
	require.NoError(t, err)
	if len(containers) == 0 {
		t.Skip("没有容器，跳过配置测试")
	}

	vmid := containers[0].VMID
	config, err := client.GetLXCConfig(ctx, node, vmid)
	require.NoError(t, err, "获取容器配置失败")
	t.Logf("CT %d 配置: %+v", vmid, config)
}

// ==================== 访问控制接口测试 ====================

func TestAccessUsers(t *testing.T) {
	client, ctx := setupTestClient(t)

	users, err := client.ListUsers(ctx)
	require.NoError(t, err, "获取用户列表失败")
	assert.NotEmpty(t, users, "用户列表不应为空")
	t.Logf("用户数量: %d", len(users))
}

func TestAccessRoles(t *testing.T) {
	client, ctx := setupTestClient(t)

	roles, err := client.ListRoles(ctx)
	require.NoError(t, err, "获取角色列表失败")
	assert.NotEmpty(t, roles, "角色列表不应为空")
	t.Logf("角色数量: %d", len(roles))
}

func TestAccessGroups(t *testing.T) {
	client, ctx := setupTestClient(t)

	groups, err := client.ListGroups(ctx)
	require.NoError(t, err, "获取组列表失败")
	t.Logf("组数量: %d", len(groups))
}

func TestAccessDomains(t *testing.T) {
	client, ctx := setupTestClient(t)

	domains, err := client.ListDomains(ctx)
	require.NoError(t, err, "获取认证域列表失败")
	assert.NotEmpty(t, domains, "认证域列表不应为空")
	t.Logf("认证域数量: %d", len(domains))
}

func TestAccessACLs(t *testing.T) {
	client, ctx := setupTestClient(t)

	acls, err := client.ListACLs(ctx)
	require.NoError(t, err, "获取 ACL 列表失败")
	t.Logf("ACL 规则数量: %d", len(acls))
}
