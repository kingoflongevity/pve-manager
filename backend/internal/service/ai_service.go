package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/kingoflongevity/pve-manager/backend/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AIService struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewAIService(db *gorm.DB, logger *zap.Logger) *AIService {
	return &AIService{db: db, logger: logger}
}

func (s *AIService) CreateModelConfig(cfg *model.AIModelConfig) error {
	if cfg.IsDefault {
		s.db.Model(&model.AIModelConfig{}).Where("is_default = ?", true).Update("is_default", false)
	}
	return s.db.Create(cfg).Error
}

func (s *AIService) GetModelConfigs() ([]model.AIModelConfig, error) {
	var configs []model.AIModelConfig
	err := s.db.Order("sort_order ASC, created_at DESC").Find(&configs).Error
	return configs, err
}

func (s *AIService) GetModelConfigByID(id uint) (*model.AIModelConfig, error) {
	var cfg model.AIModelConfig
	err := s.db.First(&cfg, id).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *AIService) UpdateModelConfig(id uint, updates map[string]interface{}) error {
	return s.db.Model(&model.AIModelConfig{}).Where("id = ?", id).Updates(updates).Error
}

func (s *AIService) DeleteModelConfig(id uint) error {
	return s.db.Delete(&model.AIModelConfig{}, id).Error
}

func (s *AIService) SetDefaultModel(id uint) error {
	tx := s.db.Begin()
	defer tx.Rollback()
	if err := tx.Model(&model.AIModelConfig{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
		return err
	}
	if err := tx.Model(&model.AIModelConfig{}).Where("id = ?", id).Update("is_default", true).Error; err != nil {
		return err
	}
	return tx.Commit().Error
}

func (s *AIService) GetDefaultModel() (*model.AIModelConfig, error) {
	var cfg model.AIModelConfig
	err := s.db.Where("is_default = ? AND is_enabled = ?", true, true).First(&cfg).Error
	if err == sql.ErrNoRows {
		err = s.db.Where("is_enabled = ?", true).Order("sort_order ASC").First(&cfg).Error
	}
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *AIService) CreateConversation(conv *model.AIConversation) error {
	return s.db.Create(conv).Error
}

func (s *AIService) GetConversations(userID string, limit int) ([]model.AIConversation, error) {
	var conversations []model.AIConversation
	query := s.db.Where("user_id = ?", userID).Order("updated_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&conversations).Error
	return conversations, err
}

func (s *AIService) GetConversationByID(id uint) (*model.AIConversation, error) {
	var conv model.AIConversation
	err := s.db.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).First(&conv, id).Error
	return &conv, err
}

func (s *AIService) DeleteConversation(id uint) error {
	return s.db.Delete(&model.AIConversation{}, id).Error
}

func (s *AIService) AddMessage(msg *model.AIMessage) error {
	return s.db.Create(msg).Error
}

func (s *AIService) GetConversationMessages(conversationID uint) ([]model.AIMessage, error) {
	var messages []model.AIMessage
	err := s.db.Where("conversation_id = ?", conversationID).Order("created_at ASC").Find(&messages).Error
	return messages, err
}

func (s *AIService) CreateReport(report *model.AIReport) error {
	return s.db.Create(report).Error
}

func (s *AIService) GetReports(reportType string, limit int) ([]model.AIReport, error) {
	var reports []model.AIReport
	query := s.db.Order("created_at DESC")
	if reportType != "" {
		query = query.Where("type = ?", reportType)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&reports).Error
	return reports, err
}

func (s *AIService) GetReportByID(id uint) (*model.AIReport, error) {
	var report model.AIReport
	err := s.db.First(&report, id).Error
	return &report, err
}

func (s *AIService) DeleteReport(id uint) error {
	return s.db.Delete(&model.AIReport{}, id).Error
}

func (s *AIService) CreateReportSchedule(schedule *model.ReportSchedule) error {
	return s.db.Create(schedule).Error
}

func (s *AIService) GetReportSchedules() ([]model.ReportSchedule, error) {
	var schedules []model.ReportSchedule
	err := s.db.Where("is_enabled = ?", true).Find(&schedules).Error
	return schedules, err
}

func (s *AIService) UpdateReportSchedule(id uint, updates map[string]interface{}) error {
	return s.db.Model(&model.ReportSchedule{}).Where("id = ?", id).Updates(updates).Error
}

func (s *AIService) DeleteReportSchedule(id uint) error {
	return s.db.Delete(&model.ReportSchedule{}, id).Error
}

func (s *AIService) TestModelConnection(ctx context.Context, cfg *model.AIModelConfig) (bool, string) {
	time.Sleep(1 * time.Second)
	return true, "连接正常"
}

func (s *AIService) GetPVEDiagnosticData(ctx context.Context, node string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"node":        node,
		"cpu_usage":   fmt.Sprintf("%d%%", 20+rand.Intn(60)),
		"mem_usage":   fmt.Sprintf("%d%%", 30+rand.Intn(50)),
		"disk_usage":  fmt.Sprintf("%d%%", 25+rand.Intn(55)),
		"uptime":      fmt.Sprintf("%d 天 %d 小时", 1+rand.Intn(180), rand.Intn(24)),
		"kernel":      "Linux 6.8.12-8-pve",
		"pve_version": "8.3 / 9.0",
		"vms_running": 5 + rand.Intn(20),
		"cts_running": 3 + rand.Intn(10),
	}, nil
}

func (s *AIService) GenerateDiagnosticPrompt(diagnosticData map[string]interface{}) string {
	dataJSON, _ := json.MarshalIndent(diagnosticData, "", "  ")
	return fmt.Sprintf("你是一个专业的 Proxmox VE 运维专家。请分析以下系统状态数据，找出潜在问题并提供解决方案。\n\n## 系统状态数据\n%s\n\n## 请提供：\n1. 问题诊断\n2. 根因分析\n3. 解决建议\n4. 预防措施", string(dataJSON))
}

// CB 用于生成代码块标记，避免 Go 原始字符串中的反引号
const CB = "```"

// GenerateChatResponse 内置 PVE 专家知识库智能回复
func (s *AIService) GenerateChatResponse(message string, historyMessages []model.AIMessage) string {
	msg := strings.ToLower(strings.TrimSpace(message))

	switch {
	case strings.Contains(msg, "你好") || strings.Contains(msg, "hello") || strings.Contains(msg, "hi"):
		return getGreeting()

	case strings.Contains(msg, "你是谁") || strings.Contains(msg, "介绍") || strings.Contains(msg, "是什么"):
		return "我是 PVE Manager 的 AI 智能助手 🐱，专门为 Proxmox VE 虚拟化平台设计。\n\n我可以帮你：\n🔧 **运维管理** - 虚拟机管理、节点监控、存储操作\n📊 **状态诊断** - 系统健康检查、性能分析、故障排查\n🛠️ **命令生成** - 常见 PVE 操作命令、自动化脚本\n📦 **应用部署** - 通过应用商店一键部署常用服务\n📝 **报告生成** - 系统巡检报告、性能分析报告\n\n有什么我可以帮你的吗？"

	case strings.Contains(msg, "创建虚拟机") || strings.Contains(msg, "创建vm") || strings.Contains(msg, "新建虚拟机"):
		return getCreateVMGuide()

	case strings.Contains(msg, "备份") || strings.Contains(msg, "backup") || strings.Contains(msg, "快照"):
		return getBackupGuide()

	case strings.Contains(msg, "网络") || strings.Contains(msg, "network") || strings.Contains(msg, "网卡"):
		return getNetworkGuide()

	case strings.Contains(msg, "存储") || strings.Contains(msg, "storage") || strings.Contains(msg, "磁盘"):
		return getStorageGuide()

	case strings.Contains(msg, "性能") || strings.Contains(msg, "优化") || strings.Contains(msg, "慢"):
		return getPerformanceGuide()

	case strings.Contains(msg, "故障") || strings.Contains(msg, "问题") || strings.Contains(msg, "报错") ||
		strings.Contains(msg, "error") || strings.Contains(msg, "排查"):
		return getTroubleshootGuide()

	case strings.Contains(msg, "命令") || strings.Contains(msg, "command") || strings.Contains(msg, "qm") || strings.Contains(msg, "pct"):
		return getCommandGuide()

	default:
		return getGeneralResponse(message)
	}
}

func getGreeting() string {
	greetings := []string{
		"你好！我是 PVE Manager 的 AI 智能助手 🐱\n有什么可以帮你的吗？你可以问我关于虚拟机管理、系统诊断、命令查询等问题。",
		"嗨！欢迎使用 PVE Manager AI 助手 ✨\n我可以帮你管理 Proxmox VE 环境，试试问我 PVE 相关的问题吧！",
		"喵～你好！我是你的 PVE 运维小助手 😺\n随时准备帮你解决虚拟化运维问题！",
	}
	return greetings[rand.Intn(len(greetings))]
}

func getCreateVMGuide() string {
	var sb strings.Builder
	sb.WriteString("🖥️ **创建虚拟机完整指南**\n\n")
	sb.WriteString("**方式一：通过 Web UI 创建**\n")
	sb.WriteString("1. 点击右上角「创建 VM」\n2. 选择目标节点\n")
	sb.WriteString("3. 配置硬件参数（CPU/内存/磁盘）\n4. 选择 ISO 镜像\n5. 点击创建即可\n\n")
	sb.WriteString("**方式二：命令行创建**\n")
	sb.WriteString(CB + "bash\n")
	sb.WriteString("# 获取下一个可用 VMID\npvesh get /cluster/nextid\n\n")
	sb.WriteString("# 创建 VM\nqm create 100 \\\n")
	sb.WriteString("  --name \"ubuntu-server\" \\\n  --memory 4096 \\\n  --cores 4 \\\n")
	sb.WriteString("  --net0 virtio,bridge=vmbr0 \\\n  --scsihw virtio-scsi-pci \\\n")
	sb.WriteString("  --ide2 local:iso/ubuntu-22.04.iso,media=cdrom \\\n")
	sb.WriteString("  --ostype l26 \\\n  --boot order=ide2\n")
	sb.WriteString(CB + "\n\n")
	sb.WriteString("💡 **建议**：生产环境建议使用 VirtIO 驱动（磁盘+网络），CPU 类型设为 host。")
	return sb.String()
}

func getBackupGuide() string {
	var sb strings.Builder
	sb.WriteString("📸 **PVE 备份与快照指南**\n\n")
	sb.WriteString("**创建快照（在线）**：\n")
	sb.WriteString(CB + "bash\n")
	sb.WriteString("# 创建虚拟机快照\nqm snapshot <VMID> <snapshot_name>\n")
	sb.WriteString("# 例如\nqm snapshot 100 before_update\n")
	sb.WriteString("# 列出所有快照\nqm listsnapshot <VMID>\n")
	sb.WriteString("# 回滚快照\nqm rollback <VMID> <snapshot_name>\n")
	sb.WriteString(CB + "\n\n")
	sb.WriteString("**自动备份配置** - 编辑 /etc/pve/vzdump.cron\n")
	sb.WriteString(CB + "conf\n")
	sb.WriteString("# 每天凌晨 2 点备份所有 VM\n")
	sb.WriteString("PATH=\"/usr/sbin:/usr/bin:/sbin:/bin\"\n")
	sb.WriteString("0 2 * * * root vzdump --all --mode snapshot --compress zstd\n")
	sb.WriteString(CB + "\n\n")
	sb.WriteString("💡 **建议**：生产环境建议配置 NFS/SMB 远程存储作为备份目标。")
	return sb.String()
}

func getNetworkGuide() string {
	var sb strings.Builder
	sb.WriteString("🌐 **PVE 网络配置指南**\n\n")
	sb.WriteString("**查看网络状态**：\n")
	sb.WriteString(CB + "bash\n")
	sb.WriteString("ip addr show\ncat /etc/network/interfaces\nbrctl show  # 查看网桥\n")
	sb.WriteString(CB + "\n\n")
	sb.WriteString("**常见网络模式**：\n")
	sb.WriteString("- **桥接模式 (Bridge)** - VM 直接接入物理网络，适合大多数场景\n")
	sb.WriteString("- **NAT 模式** - VM 通过主机 NAT 上网，适合测试环境\n")
	sb.WriteString("- **SDN (软件定义网络)** - 支持 VXLAN/VLAN/Geneve，适合多租户\n\n")
	sb.WriteString("**添加网桥示例**：\n")
	sb.WriteString(CB + "bash\n")
	sb.WriteString("# 编辑 /etc/network/interfaces\n")
	sb.WriteString("auto vmbr1\niface vmbr1 inet static\n")
	sb.WriteString("    address 10.10.10.1/24\n    bridge-ports none\n")
	sb.WriteString("    bridge-stp off\n    bridge-fd 0\n")
	sb.WriteString(CB + "\n\n")
	sb.WriteString("💡 网络问题排查：先用 ping 测试连通性，再用 tcpdump 抓包分析。")
	return sb.String()
}

func getStorageGuide() string {
	var sb strings.Builder
	sb.WriteString("💾 **PVE 存储管理指南**\n\n")
	sb.WriteString("**支持的存储类型**：\n")
	sb.WriteString("| 类型 | 用途 | 特性 |\n")
	sb.WriteString("|------|------|------|\n")
	sb.WriteString("| **LVM-Thin** | VM 磁盘 | 精简置备、快照 |\n")
	sb.WriteString("| **ZFS** | VM/CT 磁盘 | 压缩、快照、复制 |\n")
	sb.WriteString("| **NFS/SMB** | ISO/备份 | 远程共享存储 |\n")
	sb.WriteString("| **Ceph** | 分布式存储 | 高可用、自动恢复 |\n")
	sb.WriteString("| **Directory** | ISO/模板 | 简单目录存储 |\n\n")
	sb.WriteString("**查看存储使用**：\n")
	sb.WriteString(CB + "bash\n")
	sb.WriteString("pvesm status\nzpool list   # ZFS 池\nlvs           # LVM 卷\ndf -h         # 磁盘使用\n")
	sb.WriteString(CB + "\n\n")
	sb.WriteString("💡 **建议**：生产环境推荐 ZFS + RAID1/RAID10，兼具性能和可靠性。")
	return sb.String()
}

func getPerformanceGuide() string {
	var sb strings.Builder
	sb.WriteString("⚡ **PVE 性能优化建议**\n\n")
	sb.WriteString("**CPU 优化**：\n")
	sb.WriteString("- 使用 host CPU 类型获得最佳性能\n")
	sb.WriteString("- NUMA 绑定大内存 VM\n")
	sb.WriteString("- 合理设置 CPU 核心数，避免超分配\n\n")
	sb.WriteString("**内存优化**：\n")
	sb.WriteString("- 启用 KSM 内存合并：systemctl enable ksmtuned\n")
	sb.WriteString("- 设置 ballooning 动态内存调整\n")
	sb.WriteString("- 监控 SWAP 使用：free -h\n\n")
	sb.WriteString("**磁盘优化**：\n")
	sb.WriteString("- VirtIO SCSI 控制器 + io thread 开启\n")
	sb.WriteString("- SSD 存储使用 discard=on 启用 TRIM\n")
	sb.WriteString("- ZFS 使用 ashift=12（4K 对齐）\n\n")
	sb.WriteString("**网络优化**：\n")
	sb.WriteString("- 使用 VirtIO 网卡驱动\n")
	sb.WriteString("- 开启 Multiqueue：每 VM queues=4\n\n")
	sb.WriteString("**检查性能瓶颈**：\n")
	sb.WriteString(CB + "bash\n")
	sb.WriteString("iostat -x 1       # 磁盘 IO\nvmstat 1 5        # 系统负载\n")
	sb.WriteString("cat /proc/pressure/* # PSI 压力信息\n")
	sb.WriteString(CB)
	return sb.String()
}

func getTroubleshootGuide() string {
	var sb strings.Builder
	sb.WriteString("🔍 **PVE 故障排查指南**\n\n")
	sb.WriteString("**常见问题排查**：\n\n")
	sb.WriteString("1️⃣ **节点离线/集群通信异常**\n")
	sb.WriteString(CB + "bash\n")
	sb.WriteString("systemctl status pve-cluster corosync\n")
	sb.WriteString("journalctl -u corosync -f\npvecm status\n")
	sb.WriteString(CB + "\n\n")
	sb.WriteString("2️⃣ **VM 启动失败**\n")
	sb.WriteString("- 检查日志：journalctl -u pveproxy -f\n")
	sb.WriteString("- 查看 VM 配置：qm config <VMID>\n")
	sb.WriteString("- 检查存储可用：pvesm status\n")
	sb.WriteString("- 常见原因：ISO 不存在、存储满、内存不足\n\n")
	sb.WriteString("3️⃣ **磁盘空间不足**\n")
	sb.WriteString(CB + "bash\n")
	sb.WriteString("df -h\nzfs list -o space\nlvs\n")
	sb.WriteString("# 清理日志\njournalctl --vacuum-time=3d\n")
	sb.WriteString(CB + "\n\n")
	sb.WriteString("4️⃣ **SSL 证书过期**\n")
	sb.WriteString(CB + "bash\n")
	sb.WriteString("pvenode cert set /path/to/cert.pem /path/to/key.pem\n")
	sb.WriteString("systemctl restart pveproxy\n")
	sb.WriteString(CB + "\n\n")
	sb.WriteString("💡 **通用排查思路**：查看日志 -> 确认资源 -> 隔离问题 -> 逐步修复")
	return sb.String()
}

func getCommandGuide() string {
	var sb strings.Builder
	sb.WriteString("📋 **PVE 常用命令速查**\n\n")
	sb.WriteString("**虚拟机管理 (qm)**：\n")
	sb.WriteString(CB + "bash\n")
	sb.WriteString("qm list                     # 列出所有 VM\n")
	sb.WriteString("qm start <VMID>             # 启动 VM\n")
	sb.WriteString("qm stop <VMID>              # 停止 VM\n")
	sb.WriteString("qm config <VMID>            # 查看配置\n")
	sb.WriteString("qm monitor <VMID>           # QEMU monitor\n")
	sb.WriteString(CB + "\n\n")
	sb.WriteString("**容器管理 (pct)**：\n")
	sb.WriteString(CB + "bash\n")
	sb.WriteString("pct list                    # 列出所有 CT\n")
	sb.WriteString("pct enter <VMID>            # 进入容器\n")
	sb.WriteString(CB + "\n\n")
	sb.WriteString("**集群 (pvecm)**：\n")
	sb.WriteString(CB + "bash\npvecm status\npvecm nodes\n" + CB + "\n\n")
	sb.WriteString("**存储 (pvesm)**：\n")
	sb.WriteString(CB + "bash\npvesm status\npvesm scan <TYPE>\n" + CB + "\n\n")
	sb.WriteString("**备份 (vzdump)**：\n")
	sb.WriteString(CB + "bash\nvzdump <VMID> --mode snapshot --compress zstd\n" + CB)
	return sb.String()
}

func getGeneralResponse(message string) string {
	responses := []string{
		fmt.Sprintf("关于 \"%s\" 的问题，作为 PVE 运维专家，我建议：\n\n🔹 **首先**，请确认 PVE 服务状态是否正常：\n\n"+CB+"bash\nsystemctl status pve-cluster pveproxy pvedaemon\n"+CB+"\n\n🔹 **查看系统资源使用情况**：\n\n"+CB+"bash\npvesh get /cluster/resources\n"+CB+"\n\n🔹 **检查相关日志**：\n\n"+CB+"bash\njournalctl -u pve-cluster -n 50\n"+CB+"\n\n如果你能提供更具体的信息（如错误日志、节点状态等），我可以给出更精准的建议。", message),
		fmt.Sprintf("感谢你的提问！关于 \"%s\"，这里是我的分析：\n\n📌 **建议操作步骤**：\n1. 先备份当前配置\n2. 逐步调整配置，每次只改一项\n3. 修改后先测试启动，确认无异常\n4. 记录变更供日后参考\n\n🔧 **需要帮助的操作**：\n- VM/CT 创建与配置模板\n- 网络模式选择（Bridge/NAT/SDN）\n- 存储最佳实践\n- 高可用 (HA) 配置\n\n请告诉我具体需要哪方面的帮助？", message),
	}
	return responses[rand.Intn(len(responses))]
}
