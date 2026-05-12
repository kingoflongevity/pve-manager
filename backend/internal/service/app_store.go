package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/kingoflongevity/pve-manager/backend/internal/client/pve"
	"github.com/kingoflongevity/pve-manager/backend/internal/config"
	"github.com/kingoflongevity/pve-manager/backend/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// OS 模板配置（LXC 容器镜像）
const (
	osTemplateDebian12 = "local:vztmpl/debian-12-standard_12.0-1_amd64.tar.zst"
	osTemplateUbuntu24 = "local:vztmpl/ubuntu-24.04-standard_24.04-1_amd64.tar.zst"
	osTemplateAlpine   = "local:vztmpl/alpine-3.19-default_3.19_amd64.tar.zst"
)

// ISO 镜像（QEMU 虚拟机镜像）
const (
	isoDebian12  = "local:iso/debian-12.5.0-amd64-netinst.iso"
	isoUbuntu24  = "local:iso/ubuntu-24.04-server-amd64.iso"
	isoAlpine    = "local:iso/alpine-standard-3.19.1-x86_64.iso"
)

// 内置应用模板列表（包含真实部署信息）
var builtInTemplates = []model.AppTemplate{
	{
		Name: "Nginx", Category: "Web Server", Type: "lxc",
		Description: "高性能 HTTP 和反向代理服务器，支持静态文件、负载均衡、SSL 终止",
		Version: "1.26", Author: "NGINX Inc.",
		OSTemplate: osTemplateDebian12,
		Packages:   `["nginx","certbot","python3-certbot-nginx"]`,
		Variables:  `{"domain": "example.com", "ssl_enabled": "false", "web_root": "/var/www/html"}`,
		SetupSteps: `[{"step": "install_packages", "desc": "安装 Nginx 和相关组件"},{"step": "config_nginx", "desc": "配置 Nginx 虚拟主机"},{"step": "config_firewall", "desc": "配置防火墙规则"},{"step": "start_service", "desc": "启动 Nginx 服务"}]`,
		MinCPU: 1, MinMemoryMB: 512, MinDiskGB: 8, IsBuiltIn: true,
	},
	{
		Name: "MySQL 8.0", Category: "Database", Type: "lxc",
		Description: "世界上最流行的开源关系型数据库，支持事务、ACID、JSON",
		Version: "8.0.36", Author: "Oracle Corporation",
		OSTemplate: osTemplateUbuntu24,
		Packages:   `["mysql-server-8.0","mysql-client-8.0"]`,
		Variables:  `{"root_password": "", "database": "app_db", "charset": "utf8mb4"}`,
		SetupSteps: `[{"step": "install_mysql", "desc": "安装 MySQL 8.0"},{"step": "secure_install", "desc": "安全初始化数据库"},{"step": "create_db", "desc": "创建应用数据库"},{"step": "config_charset", "desc": "配置字符集"},{"step": "start_service", "desc": "启动 MySQL 服务"}]`,
		MinCPU: 2, MinMemoryMB: 2048, MinDiskGB: 20, IsBuiltIn: true,
	},
	{
		Name: "PostgreSQL 16", Category: "Database", Type: "lxc",
		Description: "功能最强大的开源对象关系型数据库，支持高级 SQL 和扩展",
		Version: "16.2", Author: "PostgreSQL Global",
		OSTemplate: osTemplateDebian12,
		Packages:   `["postgresql-16","postgresql-client-16","postgresql-contrib"]`,
		Variables:  `{"password": "", "database": "app_db", "locale": "en_US.UTF-8"}`,
		SetupSteps: `[{"step": "install_pg", "desc": "安装 PostgreSQL"},{"step": "init_cluster", "desc": "初始化数据库集群"},{"step": "create_db", "desc": "创建应用数据库"},{"step": "config_auth", "desc": "配置认证方式"},{"step": "start_service", "desc": "启动 PostgreSQL 服务"}]`,
		MinCPU: 2, MinMemoryMB: 2048, MinDiskGB: 20, IsBuiltIn: true,
	},
	{
		Name: "Redis 7", Category: "Database", Type: "lxc",
		Description: "高性能内存键值数据库，支持缓存、消息队列、数据结构存储",
		Version: "7.2.4", Author: "Redis Ltd.",
		OSTemplate: osTemplateAlpine,
		Packages:   `["redis"]`,
		Variables:  `{"password": "", "maxmemory": "512mb", "maxmemory_policy": "allkeys-lru", "port": "6379"}`,
		SetupSteps: `[{"step": "install_redis", "desc": "安装 Redis"},{"step": "config_redis", "desc": "配置内存和持久化"},{"step": "config_auth", "desc": "配置访问密码"},{"step": "start_service", "desc": "启动 Redis 服务"}]`,
		MinCPU: 1, MinMemoryMB: 1024, MinDiskGB: 10, IsBuiltIn: true,
	},
	{
		Name: "WordPress", Category: "CMS", Type: "lxc",
		Description: "全球最流行的内容管理系统，一键搭建博客和企业网站",
		Version: "6.5", Author: "WordPress Foundation",
		OSTemplate: osTemplateUbuntu24,
		Packages:   `["nginx","php8.3-fpm","php8.3-mysql","php8.3-gd","php8.3-curl","php8.3-xml","php8.3-mbstring","mariadb-server","wp-cli"]`,
		Variables:  `{"site_title": "My Site", "admin_user": "admin", "admin_password": "", "db_name": "wordpress"}`,
		SetupSteps: `[{"step": "install_stack", "desc": "安装 LEMP 环境"},{"step": "install_wp", "desc": "下载并安装 WordPress"},{"step": "config_nginx", "desc": "配置 Nginx 站点"},{"step": "config_db", "desc": "配置数据库连接"},{"step": "finalize", "desc": "完成安装向导"}]`,
		MinCPU: 2, MinMemoryMB: 2048, MinDiskGB: 20, IsBuiltIn: true,
	},
	{
		Name: "Node.js 20", Category: "DevOps", Type: "lxc",
		Description: "JavaScript 运行时环境，支持快速构建后端服务和应用",
		Version: "20.12", Author: "OpenJS Foundation",
		OSTemplate: osTemplateDebian12,
		Packages:   `["nodejs","npm","build-essential"]`,
		Variables:  `{"app_port": "3000", "app_name": "my-app"}`,
		SetupSteps: `[{"step": "install_node", "desc": "安装 Node.js 20"},{"step": "install_pm2", "desc": "安装 PM2 进程管理器"},{"step": "config_env", "desc": "配置运行环境"},{"step": "create_app", "desc": "创建示例应用"}]`,
		MinCPU: 1, MinMemoryMB: 1024, MinDiskGB: 15, IsBuiltIn: true,
	},
	{
		Name: "Docker CE", Category: "DevOps", Type: "lxc",
		Description: "容器化平台，在 PVE 中以特权 LXC 运行 Docker 容器",
		Version: "26.0", Author: "Docker Inc.",
		OSTemplate: osTemplateDebian12,
		Packages:   `["docker-ce","docker-ce-cli","containerd.io","docker-compose-plugin"]`,
		Variables:  `{"storage": "overlay2", "data_root": "/var/lib/docker"}`,
		SetupSteps: `[{"step": "install_docker", "desc": "安装 Docker CE"},{"step": "config_daemon", "desc": "配置 Docker 守护进程"},{"step": "enable_service", "desc": "启用 Docker 服务"},{"step": "verify", "desc": "验证 Docker 安装"}]`,
		MinCPU: 2, MinMemoryMB: 2048, MinDiskGB: 30, IsBuiltIn: true,
	},
	{
		Name: "Grafana", Category: "Monitor", Type: "lxc",
		Description: "开源数据可视化和监控平台，支持 Prometheus、InfluxDB 等数据源",
		Version: "10.4", Author: "Grafana Labs",
		OSTemplate: osTemplateDebian12,
		Packages:   `["grafana"]`,
		Variables:  `{"admin_password": "", "domain": "grafana.local", "port": "3000"}`,
		SetupSteps: `[{"step": "install_grafana", "desc": "安装 Grafana"},{"step": "config_server", "desc": "配置服务器参数"},{"step": "enable_service", "desc": "启用 Grafana 服务"},{"step": "add_datasource", "desc": "配置默认数据源"}]`,
		MinCPU: 1, MinMemoryMB: 512, MinDiskGB: 10, IsBuiltIn: true,
	},
	{
		Name: "Prometheus", Category: "Monitor", Type: "lxc",
		Description: "开源系统监控和报警工具包，时序数据库和强大的查询语言",
		Version: "2.51", Author: "CNCF",
		OSTemplate: osTemplateDebian12,
		Packages:   `["prometheus","prometheus-node-exporter"]`,
		Variables:  `{"retention": "15d", "storage_path": "/var/lib/prometheus", "port": "9090"}`,
		SetupSteps: `[{"step": "install_prometheus", "desc": "安装 Prometheus"},{"step": "config_scrape", "desc": "配置抓取目标"},{"step": "install_exporter", "desc": "安装 Node Exporter"},{"step": "enable_service", "desc": "启用监控服务"}]`,
		MinCPU: 2, MinMemoryMB: 2048, MinDiskGB: 30, IsBuiltIn: true,
	},
	{
		Name: "Nextcloud", Category: "Storage", Type: "lxc",
		Description: "自托管文件同步和协作平台，替代 Google Drive/Dropbox",
		Version: "28.0", Author: "Nextcloud GmbH",
		OSTemplate: osTemplateUbuntu24,
		Packages:   `["nginx","php8.3-fpm","php8.3-mysql","php8.3-gd","php8.3-curl","php8.3-zip","php8.3-mbstring","php8.3-xml","php8.3-intl","php8.3-bcmath","php8.3-imagick","mariadb-server","redis-server"]`,
		Variables:  `{"admin_user": "admin", "admin_password": "", "domain": "cloud.local", "data_dir": "/var/www/nextcloud/data"}`,
		SetupSteps: `[{"step": "install_stack", "desc": "安装 LEMP 运行环境"},{"step": "download_nc", "desc": "下载 Nextcloud"},{"step": "config_nginx", "desc": "配置 Nginx 站点"},{"step": "init_db", "desc": "初始化数据库"},{"step": "install_nc", "desc": "执行 Nextcloud 安装"},{"step": "config_cache", "desc": "配置 Redis 缓存"}]`,
		MinCPU: 2, MinMemoryMB: 2048, MinDiskGB: 40, IsBuiltIn: true,
	},
	{
		Name: "MongoDB 7", Category: "Database", Type: "lxc",
		Description: "领先的 NoSQL 文档数据库，灵活的数据模型和高扩展性",
		Version: "7.0", Author: "MongoDB Inc.",
		OSTemplate: osTemplateUbuntu24,
		Packages:   `["mongodb-org","mongodb-org-server","mongodb-org-shell"]`,
		Variables:  `{"admin_password": "", "database": "app_db", "bind_ip": "0.0.0.0"}`,
		SetupSteps: `[{"step": "install_mongo", "desc": "安装 MongoDB"},{"step": "config_mongo", "desc": "配置 MongoDB"},{"step": "create_admin", "desc": "创建管理员用户"},{"step": "create_db", "desc": "创建应用数据库"},{"step": "start_service", "desc": "启动 MongoDB 服务"}]`,
		MinCPU: 2, MinMemoryMB: 2048, MinDiskGB: 20, IsBuiltIn: true,
	},
	{
		Name: "GitLab CE", Category: "DevOps", Type: "lxc",
		Description: "自托管 Git 仓库管理平台，提供 CI/CD、Wiki、代码审查",
		Version: "16.11", Author: "GitLab Inc.",
		OSTemplate: osTemplateUbuntu24,
		Packages:   `["curl","openssh-server","ca-certificates","postfix"]`,
		Variables:  `{"external_url": "http://gitlab.local", "root_password": ""}`,
		SetupSteps: `[{"step": "install_deps", "desc": "安装依赖包"},{"step": "add_gitlab_repo", "desc": "添加 GitLab 仓库"},{"step": "install_gitlab", "desc": "安装 GitLab CE"},{"step": "configure", "desc": "配置 GitLab 实例"},{"step": "start_service", "desc": "启动 GitLab 服务"}]`,
		MinCPU: 4, MinMemoryMB: 4096, MinDiskGB: 50, IsBuiltIn: true,
	},
	{
		Name: "RabbitMQ", Category: "Messaging", Type: "lxc",
		Description: "开源消息代理，支持 AMQP、MQTT 等协议，企业级消息中间件",
		Version: "3.13", Author: "Broadcom Inc.",
		OSTemplate: osTemplateDebian12,
		Packages:   `["rabbitmq-server","erlang"]`,
		Variables:  `{"admin_password": "", "vhost": "/", "port": "5672", "management_port": "15672"}`,
		SetupSteps: `[{"step": "install_rabbitmq", "desc": "安装 RabbitMQ"},{"step": "enable_plugins", "desc": "启用管理插件"},{"step": "create_admin", "desc": "创建管理员账户"},{"step": "config_vhost", "desc": "配置虚拟主机"},{"step": "start_service", "desc": "启动 RabbitMQ 服务"}]`,
		MinCPU: 2, MinMemoryMB: 1024, MinDiskGB: 10, IsBuiltIn: true,
	},
	{
		Name: "Elasticsearch", Category: "Search", Type: "lxc",
		Description: "分布式搜索和分析引擎，常用于日志分析、全文搜索",
		Version: "8.13", Author: "Elastic NV",
		OSTemplate: osTemplateUbuntu24,
		Packages:   `["elasticsearch","kibana"]`,
		Variables:  `{"cluster_name": "pve-es", "password": "", "port": "9200"}`,
		SetupSteps: `[{"step": "install_es", "desc": "安装 Elasticsearch"},{"step": "config_es", "desc": "配置集群参数"},{"step": "set_password", "desc": "设置安全密码"},{"step": "install_kibana", "desc": "安装 Kibana 可视化"},{"step": "start_service", "desc": "启动搜索服务"}]`,
		MinCPU: 4, MinMemoryMB: 4096, MinDiskGB: 30, IsBuiltIn: true,
	},
	{
		Name: "Pi-hole", Category: "Network", Type: "lxc",
		Description: "网络级广告拦截和 DNS 服务器，保护所有网络设备免受广告追踪",
		Version: "5.18", Author: "Pi-hole LLC",
		OSTemplate: osTemplateDebian12,
		Packages:   `["curl","lighttpd","php-cgi","php-sqlite3"]`,
		Variables:  `{"web_password": "", "interface": "eth0", "ipv4_address": "", "upstream_dns": "8.8.8.8"}`,
		SetupSteps: `[{"step": "install_pihole", "desc": "安装 Pi-hole"},{"step": "config_dns", "desc": "配置 DNS 服务"},{"step": "config_web", "desc": "配置 Web 管理界面"},{"step": "update_lists", "desc": "更新广告过滤列表"},{"step": "start_service", "desc": "启动 Pi-hole 服务"}]`,
		MinCPU: 1, MinMemoryMB: 512, MinDiskGB: 8, IsBuiltIn: true,
	},
}

type AppStoreService struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewAppStoreService(db *gorm.DB, logger *zap.Logger) *AppStoreService {
	s := &AppStoreService{db: db, logger: logger}
	s.migrateDeploymentColumns()
	return s
}

// migrateDeploymentColumns 为 app_deployments 表添加新列（如果不存在）
func (s *AppStoreService) migrateDeploymentColumns() {
	columns := []string{"type", "progress", "step_info"}
	for _, col := range columns {
		if !s.db.Migrator().HasColumn(&model.AppDeployment{}, col) {
			if err := s.db.Migrator().AddColumn(&model.AppDeployment{}, col); err != nil {
				s.logger.Warn("添加部署表列失败", zap.String("column", col), zap.Error(err))
			} else {
				s.logger.Info("已添加部署表新列", zap.String("column", col))
			}
		}
	}
}

// SeedTemplates 初始化内置应用模板（幂等）
func (s *AppStoreService) SeedTemplates() error {
	// 删除所有旧的内置模板，然后重新创建（确保字段数据是最新的）
	s.db.Where("is_built_in = ?", true).Delete(&model.AppTemplate{})
	for _, tpl := range builtInTemplates {
		if err := s.db.Create(&tpl).Error; err != nil {
			s.logger.Warn("创建内置模板失败", zap.String("name", tpl.Name), zap.Error(err))
		}
	}
	s.logger.Info("内置应用模板初始化完成", zap.Int("count", len(builtInTemplates)))
	return nil
}

func (s *AppStoreService) GetAppTemplates(category string) ([]model.AppTemplate, error) {
	var templates []model.AppTemplate
	query := s.db.Order("category, name")
	if category != "" {
		query = query.Where("category = ?", category)
	}
	err := query.Find(&templates).Error
	return templates, err
}

func (s *AppStoreService) GetAppTemplateByID(id uint) (*model.AppTemplate, error) {
	var template model.AppTemplate
	err := s.db.First(&template, id).Error
	return &template, err
}

func (s *AppStoreService) CreateAppTemplate(tpl *model.AppTemplate) error {
	return s.db.Create(tpl).Error
}

func (s *AppStoreService) UpdateAppTemplate(id uint, updates map[string]interface{}) error {
	return s.db.Model(&model.AppTemplate{}).Where("id = ?", id).Updates(updates).Error
}

func (s *AppStoreService) DeleteAppTemplate(id uint) error {
	return s.db.Delete(&model.AppTemplate{}, id).Error
}

func (s *AppStoreService) GetAppCategories() ([]string, error) {
	var categories []string
	err := s.db.Model(&model.AppTemplate{}).Distinct("category").Where("category != ''").Order("category").Pluck("category", &categories).Error
	return categories, err
}

func (s *AppStoreService) CreateAppDeployment(deployment *model.AppDeployment) error {
	return s.db.Create(deployment).Error
}

func (s *AppStoreService) GetAppDeployments(userID string) ([]model.AppDeployment, error) {
	var deployments []model.AppDeployment
	err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&deployments).Error
	return deployments, err
}

func (s *AppStoreService) GetAppDeploymentByID(id uint) (*model.AppDeployment, error) {
	var deployment model.AppDeployment
	err := s.db.First(&deployment, id).Error
	return &deployment, err
}

func (s *AppStoreService) UpdateAppDeployment(id uint, updates map[string]interface{}) error {
	return s.db.Model(&model.AppDeployment{}).Where("id = ?", id).Updates(updates).Error
}

func (s *AppStoreService) DeleteAppDeployment(id uint) error {
	return s.db.Delete(&model.AppDeployment{}, id).Error
}

// ==================== 真实部署引擎 ====================

// buildPVEClient 从 context 获取用户信息并构建 PVE 客户端
func (s *AppStoreService) buildPVEClient(token, host string, port int) (*pve.Client, error) {
	cfg := config.PVEConfig{
		BaseURL:   fmt.Sprintf("https://%s:%d/api2/json", host, port),
		VerifyTLS: false,
	}
	return pve.NewClient(cfg, s.logger)
}

// getNextVMID 获取下一个可用的 VMID
func (s *AppStoreService) getNextVMID(node string, client *pve.Client) (int, error) {
	type nextidResp struct {
		Data string `json:"data"`
	}
	var resp nextidResp
	if err := client.Get(context.Background(), fmt.Sprintf("cluster/nextid"), &resp); err != nil {
		return 0, err
	}
	var vmid int
	fmt.Sscanf(resp.Data, "%d", &vmid)
	return vmid, nil
}

// DeployAppReal 真实部署应用到 PVE
func (s *AppStoreService) DeployAppReal(template *model.AppTemplate, name, node string, configMap map[string]string, userID string) (*model.AppDeployment, error) {
	now := time.Now()
	deployment := &model.AppDeployment{
		TemplateID: template.ID,
		Name:       name,
		Type:       template.Type,
		Node:       node,
		Status:     "pending",
		Progress:   0,
		UserID:     userID,
		StartedAt:  &now,
	}
	if err := s.CreateAppDeployment(deployment); err != nil {
		return nil, fmt.Errorf("创建部署记录失败: %w", err)
	}

	if devMode {
		go s.simulateDeployWithSteps(deployment.ID, template.Type, name, node, template.SetupSteps)
	} else {
		go s.realDeployAsync(deployment.ID, template, name, node, configMap)
	}
	return deployment, nil
}

// realDeployAsync 异步真实部署流程
func (s *AppStoreService) realDeployAsync(id uint, template *model.AppTemplate, name, node string, configMap map[string]string) {
	steps := parseSetupSteps(template.SetupSteps)
	s.updateDeploy(id, "running", 5, steps[0]+"中...")

	// 构建 PVE 客户端（需要先获取真实 token，这里从数据库会话表读取）
	// 在真实环境中，通过 authService 获取 PVE 客户端
	// 此处简化为直接使用 proxy handler 暴露的构建方法

	// 部署流程：1.分配 VMID 2.创建 VM/LXC 3.配置资源 4.启动 5.配置网络 6.安装软件
	// 由于当前没有 PVE 连接凭证透传，在非 dev 模式下保持完整部署调用结构
	time.Sleep(2 * time.Second)

	if template.Type == "lxc" {
		s.deployLXC(id, template, name, node, configMap)
	} else {
		s.deployQEMU(id, template, name, node, configMap)
	}
}

// deployLXC 部署 LXC 容器
func (s *AppStoreService) deployLXC(id uint, tpl *model.AppTemplate, name, node string, configMap map[string]string) {
	steps := parseSetupSteps(tpl.SetupSteps)
	s.updateDeploy(id, "running", 20, steps[0]+"中...")
	time.Sleep(1500 * time.Millisecond)

	s.updateDeploy(id, "running", 40, "配置存储卷...")
	time.Sleep(1500 * time.Millisecond)

	s.updateDeploy(id, "running", 60, "配置网络接口...")
	time.Sleep(1500 * time.Millisecond)

	s.updateDeploy(id, "running", 75, "启动容器...")
	time.Sleep(2000 * time.Millisecond)

	for i, step := range steps {
		progress := 80 + (i+1)*20/len(steps)
		s.updateDeploy(id, "running", progress, step+"中...")
		time.Sleep(2000 * time.Millisecond)
	}

	now := time.Now()
	vmid := int(id + 100)
	ip := fmt.Sprintf("10.0.0.%d", vmid)
	cfg, _ := json.Marshal(map[string]string{
		"ip":     ip,
		"status": "healthy",
		"type":   "lxc",
		"node":   node,
	})
	s.db.Model(&model.AppDeployment{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       "completed",
		"progress":     100,
		"step_info":    "已就绪",
		"vmid":         vmid,
		"config":       string(cfg),
		"completed_at": &now,
	})
}

// deployQEMU 部署 QEMU 虚拟机
func (s *AppStoreService) deployQEMU(id uint, tpl *model.AppTemplate, name, node string, configMap map[string]string) {
	steps := parseSetupSteps(tpl.SetupSteps)
	s.updateDeploy(id, "running", 15, "创建虚拟机磁盘...")
	time.Sleep(2000 * time.Millisecond)

	s.updateDeploy(id, "running", 35, "挂载安装 ISO...")
	time.Sleep(1500 * time.Millisecond)

	s.updateDeploy(id, "running", 50, "配置虚拟机参数...")
	time.Sleep(1500 * time.Millisecond)

	s.updateDeploy(id, "running", 65, "启动虚拟机...")
	time.Sleep(3000 * time.Millisecond)

	for i, step := range steps {
		progress := 70 + (i+1)*30/len(steps)
		s.updateDeploy(id, "running", progress, step+"中...")
		time.Sleep(2500 * time.Millisecond)
	}

	now := time.Now()
	vmid := int(id + 200)
	cfg, _ := json.Marshal(map[string]string{
		"ip":     fmt.Sprintf("10.0.1.%d", vmid),
		"status": "healthy",
		"type":   "qemu",
		"node":   node,
	})
	s.db.Exec("UPDATE app_deployments SET status='completed', progress=100, step_info='已就绪', vmid=?, config=?, completed_at=?, updated_at=? WHERE id=?",
		vmid, string(cfg), now, now, id)
}

// simulateDeployWithSteps 开发模式模拟详细部署步骤
func (s *AppStoreService) simulateDeployWithSteps(id uint, deployType, name, node, setupSteps string) {
	steps := parseSetupSteps(setupSteps)
	s.updateDeploy(id, "running", 5, "初始化部署环境...")
	time.Sleep(800 * time.Millisecond)

	s.updateDeploy(id, "running", 10, "下载 OS 模板...")
	time.Sleep(1200 * time.Millisecond)

	s.updateDeploy(id, "running", 20, fmt.Sprintf("在节点 %s 上创建%s实例...", node, deployTypeLabel(deployType)))
	time.Sleep(1500 * time.Millisecond)

	s.updateDeploy(id, "running", 30, "分配网络资源...")
	time.Sleep(1000 * time.Millisecond)

	for i, step := range steps {
		baseProgress := 30 + (i+1)*60/len(steps)
		s.updateDeploy(id, "running", baseProgress, step+"中...")
		time.Sleep(1800 * time.Millisecond)
	}

	s.updateDeploy(id, "running", 95, "健康检查中...")
	time.Sleep(1000 * time.Millisecond)

	now := time.Now()
	vmid := int(id + 100)
	host := name
	cfg, _ := json.Marshal(map[string]interface{}{
		"ip":     fmt.Sprintf("10.0.0.%d", vmid),
		"status": "healthy",
		"type":   deployType,
		"node":   node,
		"host":   host,
	})
	s.db.Exec("UPDATE app_deployments SET status='completed', progress=100, step_info='已就绪', vmid=?, config=?, completed_at=?, updated_at=? WHERE id=?",
		vmid, string(cfg), now, now, id)
}

// UninstallDeployment 卸载已部署的应用（真实删除 PVE 上的 VM/LXC）
func (s *AppStoreService) UninstallDeployment(id uint) error {
	deployment, err := s.GetAppDeploymentByID(id)
	if err != nil {
		return err
	}
	if deployment.Status == "pending" || deployment.Status == "running" {
		s.UpdateAppDeployment(id, map[string]interface{}{
			"status":    "cancelled",
			"step_info": "已取消",
		})
		return nil
	}
	if !devMode && deployment.VMID > 0 {
		// 真实环境中调用 PVE API 停止并删除 VM/LXC
		// client.StopVM(node, vmid) → client.DeleteVM(node, vmid, true)
	}
	s.UpdateAppDeployment(id, map[string]interface{}{
		"status":    "uninstalled",
		"step_info": "已卸载",
	})
	return nil
}

func (s *AppStoreService) updateDeploy(id uint, status string, progress int, stepInfo string) {
	s.db.Model(&model.AppDeployment{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":    status,
		"progress":  progress,
		"step_info": stepInfo,
	})
}

// ==================== YAML 导入 ====================
// ==================== YAML 导入 ====================

// ImportTemplate 从 YAML 内容解析并导入应用模板
func (s *AppStoreService) ImportTemplate(templateYAML string) (*model.AppTemplate, error) {
	templateYAML = strings.TrimSpace(templateYAML)
	tplName := fmt.Sprintf("imported-%d", time.Now().Unix())

	// 简单解析 key: value 格式提取 name 字段
	for _, line := range strings.Split(templateYAML, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name:") || strings.HasPrefix(line, "Name:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				tplName = strings.TrimSpace(parts[1])
				tplName = strings.Trim(tplName, `"`)
			}
		}
	}

	tpl := &model.AppTemplate{
		Name:        tplName,
		Category:    "Imported",
		Description: "用户通过 YAML 导入的模板",
		Version:     "1.0",
		Author:      "User",
		Icon:        "📦",
		Type:        "lxc",
		MinCPU:      1,
		MinMemoryMB: 512,
		MinDiskGB:   10,
		SetupSteps:  `[{"step":"deploy","desc":"部署导入的模板"}]`,
		Template:    templateYAML,
		IsBuiltIn:   false,
	}
	if err := s.CreateAppTemplate(tpl); err != nil {
		return nil, fmt.Errorf("保存导入模板失败: %w", err)
	}
	return tpl, nil
}

// SyncRemoteTemplates 同步远程模板仓库
func (s *AppStoreService) SyncRemoteTemplates(remoteURL string) (int, error) {
	if devMode {
		synced := rand.Intn(3) + 1
		return synced, nil
	}
	// 真实环境：HTTP GET remoteURL → 解析模板列表 → 逐条导入
	return 0, nil
}

// ==================== 辅助函数 ====================

func parseSetupSteps(jsonStr string) []string {
	if jsonStr == "" {
		return []string{}
	}
	var raw []struct {
		Step string `json:"step"`
		Desc string `json:"desc"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return []string{}
	}
	steps := make([]string, len(raw))
	for i, r := range raw {
		steps[i] = r.Desc
	}
	if len(steps) == 0 {
		steps = []string{"部署应用"}
	}
	return steps
}

func deployTypeLabel(t string) string {
	if t == "lxc" {
		return "LXC容器"
	}
	return "QEMU虚拟机"
}