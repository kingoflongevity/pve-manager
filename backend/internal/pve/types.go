package pve

// ============================================================
// API 基础类型
// ============================================================

// APIResponse PVE API 通用响应结构
// PVE 所有 API 返回格式: { "data": { ... } }
type APIResponse struct {
	Data interface{} `json:"data"`
}

// TicketResponse 登录成功后返回的 ticket 响应
type TicketResponse struct {
	Username    string `json:"username"`
	CSRFToken   string `json:"CSRFPreventionToken"`
	ClusterName string `json:"clustername"`
	Ticket      string `json:"ticket"`
}

// UPID Proxmox 异步任务 ID 格式
// 格式: UPID:node:pid:pstart:starttime:type:id:user:
type UPID struct {
	Node      string `json:"node"`
	PID       uint64 `json:"pid"`
	PStart    uint64 `json:"pstart"`
	StartTime uint64 `json:"starttime"`
	Type      string `json:"type"`
	ID        string `json:"id"`
	User      string `json:"user"`
}

// ============================================================
// QEMU 虚拟机类型
// ============================================================

// QEMUVM QEMU 虚拟机简要信息（列表返回）
type QEMUVM struct {
	VMID      int      `json:"vmid"`
	Name      string   `json:"name,omitempty"`
	Status    string   `json:"status,omitempty"`
	CPU       float64  `json:"cpu,omitempty"`
	CPUs      int      `json:"cpus,omitempty"`
	Mem       uint64   `json:"mem,omitempty"`
	MaxMem    uint64   `json:"maxmem,omitempty"`
	DiskRead  uint64   `json:"diskread,omitempty"`
	DiskWrite uint64   `json:"diskwrite,omitempty"`
	NetIn     uint64   `json:"netin,omitempty"`
	NetOut    uint64   `json:"netout,omitempty"`
	Uptime    uint64   `json:"uptime,omitempty"`
	Template  int      `json:"template,omitempty"`
	Tags      string   `json:"tags,omitempty"`
	PID       int      `json:"pid,omitempty"`
	Node      string   `json:"node,omitempty"`
	Type      string   `json:"type,omitempty"`
	Lock      string   `json:"lock,omitempty"`
	HA        *HAState `json:"ha,omitempty"`
}

// QEMUConfig QEMU 虚拟机完整配置
type QEMUConfig struct {
	VMID           int    `json:"vmid"`
	Name           string `json:"name,omitempty"`
	Description    string `json:"description,omitempty"`
	Agent          string `json:"agent,omitempty"`
	Arch           string `json:"arch,omitempty"`
	Args           string `json:"args,omitempty"`
	Audio0         string `json:"audio0,omitempty"`
	Autostart      int    `json:"autostart,omitempty"`
	Backup         string `json:"backup,omitempty"`
	Balloon        int    `json:"balloon,omitempty"`
	Bios           string `json:"bios,omitempty"`
	Boot           string `json:"boot,omitempty"`
	BootDisk       string `json:"bootdisk,omitempty"`
	CICustom       string `json:"cicustom,omitempty"`
	CIUser         string `json:"ciuser,omitempty"`
	CIPassword     string `json:"cipassword,omitempty"`
	Cores          int    `json:"cores,omitempty"`
	CPU            string `json:"cpu,omitempty"`
	CPULimit       int    `json:"cpulimit,omitempty"`
	CPUUnits       int    `json:"cpuunits,omitempty"`
	Deleted        string `json:"deleted,omitempty"`
	EFIDisk0       string `json:"efidisk0,omitempty"`
	Freeze         int    `json:"freeze,omitempty"`
	HookScript     string `json:"hookscript,omitempty"`
	Hotplug        string `json:"hotplug,omitempty"`
	HugePages      string `json:"hugepages,omitempty"`
	Ivf            int    `json:"ivf,omitempty"`
	Keyboard       string `json:"keyboard,omitempty"`
	KVM            int    `json:"kvm,omitempty"`
	LocalTime      int    `json:"localtime,omitempty"`
	Lock           string `json:"lock,omitempty"`
	Machine        string `json:"machine,omitempty"`
	Memory         int    `json:"memory,omitempty"`
	Net0           string `json:"net0,omitempty"`
	Net1           string `json:"net1,omitempty"`
	Net2           string `json:"net2,omitempty"`
	Net3           string `json:"net3,omitempty"`
	Numa           int    `json:"numa,omitempty"`
	OnBoot         int    `json:"onboot,omitempty"`
	OSType         string `json:"ostype,omitempty"`
	Protection     int    `json:"protection,omitempty"`
	Reboot         int    `json:"reboot,omitempty"`
	RNG0           string `json:"rng0,omitempty"`
	SCSIHW         string `json:"scsihw,omitempty"`
	Serial0        string `json:"serial0,omitempty"`
	Shares         int    `json:"shares,omitempty"`
	SMBios1        string `json:"smbios1,omitempty"`
	Sockets        int    `json:"sockets,omitempty"`
	Spice          string `json:"spice,omitempty"`
	Startup        string `json:"startup,omitempty"`
	Tablet         int    `json:"tablet,omitempty"`
	Tags           string `json:"tags,omitempty"`
	TCPU           int    `json:"vcpus,omitempty"`
	TGA            int    `json:"vga,omitempty"`
	TPMState0      string `json:"tpmstate0,omitempty"`
	Unused0        string `json:"unused0,omitempty"`
	Unused1        string `json:"unused1,omitempty"`
	Unused2        string `json:"unused2,omitempty"`
	Unused3        string `json:"unused3,omitempty"`
	USB0           string `json:"usb0,omitempty"`
	VirtIO0        string `json:"virtio0,omitempty"`
	VMGenerationID string `json:"vmgenid,omitempty"`
	VCPUs          int    `json:"vcpus,omitempty"`
	VGA            string `json:"vga,omitempty"`
	Watchdog       string `json:"watchdog,omitempty"`
	IDE0           string `json:"ide0,omitempty"`
	IDE1           string `json:"ide1,omitempty"`
	IDE2           string `json:"ide2,omitempty"`
	IDE3           string `json:"ide3,omitempty"`
	SCSI0          string `json:"scsi0,omitempty"`
	SCSI1          string `json:"scsi1,omitempty"`
	SCSI2          string `json:"scsi2,omitempty"`
	SCSI3          string `json:"scsi3,omitempty"`
	SCSI4          string `json:"scsi4,omitempty"`
	SCSI5          string `json:"scsi5,omitempty"`
	SCSI6          string `json:"scsi6,omitempty"`
	SCSI7          string `json:"scsi7,omitempty"`
	SCSI8          string `json:"scsi8,omitempty"`
	SCSI9          string `json:"scsi9,omitempty"`
	SCSI10         string `json:"scsi10,omitempty"`
	SCSI11         string `json:"scsi11,omitempty"`
	SCSI12         string `json:"scsi12,omitempty"`
	SCSI13         string `json:"scsi13,omitempty"`
	SCSI14         string `json:"scsi14,omitempty"`
	SCSI15         string `json:"scsi15,omitempty"`
	SCSI16         string `json:"scsi16,omitempty"`
	SCSI17         string `json:"scsi17,omitempty"`
	SCSI18         string `json:"scsi18,omitempty"`
	SCSI19         string `json:"scsi19,omitempty"`
	SCSI20         string `json:"scsi20,omitempty"`
	SCSI21         string `json:"scsi21,omitempty"`
	SCSI22         string `json:"scsi22,omitempty"`
	SCSI23         string `json:"scsi23,omitempty"`
	SCSI24         string `json:"scsi24,omitempty"`
	SCSI25         string `json:"scsi25,omitempty"`
	SCSI26         string `json:"scsi26,omitempty"`
	SCSI27         string `json:"scsi27,omitempty"`
	SCSI28         string `json:"scsi28,omitempty"`
	SCSI29         string `json:"scsi29,omitempty"`
	SCSI30         string `json:"scsi30,omitempty"`
}

// QEMUCreateParams 创建 QEMU 虚拟机的参数
type QEMUCreateParams struct {
	VMID         int    `json:"vmid,omitempty"`
	Node         string `json:"node,omitempty"`
	Name         string `json:"name,omitempty"`
	Memory       int    `json:"memory,omitempty"`
	Cores        int    `json:"cores,omitempty"`
	Sockets      int    `json:"sockets,omitempty"`
	OSType       string `json:"ostype,omitempty"`
	Net0         string `json:"net0,omitempty"`
	SCSI0        string `json:"scsi0,omitempty"`
	Ide2         string `json:"ide2,omitempty"`
	OnBoot       int    `json:"onboot,omitempty"`
	Agent        int    `json:"agent,omitempty"`
	Bios         string `json:"bios,omitempty"`
	Boot         string `json:"boot,omitempty"`
	CPU          string `json:"cpu,omitempty"`
	Numa         int    `json:"numa,omitempty"`
	Tablet       int    `json:"tablet,omitempty"`
	VGA          string `json:"vga,omitempty"`
	Arch         string `json:"arch,omitempty"`
	Machine      string `json:"machine,omitempty"`
	SMBios1      string `json:"smbios1,omitempty"`
	CIUser       string `json:"ciuser,omitempty"`
	CIPassword   string `json:"cipassword,omitempty"`
	SSHKeys      string `json:"sshkeys,omitempty"`
	Nameserver   string `json:"nameserver,omitempty"`
	Searchdomain string `json:"searchdomain,omitempty"`
	IPConfig0    string `json:"ipconfig0,omitempty"`
	Pool         string `json:"pool,omitempty"`
	Tags         string `json:"tags,omitempty"`
	Protection   int    `json:"protection,omitempty"`
}

// QEMUCloneParams 克隆虚拟机参数
type QEMUCloneParams struct {
	NewID    int    `json:"newid"`
	Name     string `json:"name,omitempty"`
	Disk     string `json:"disk,omitempty"`
	Format   string `json:"format,omitempty"`
	Full     int    `json:"full,omitempty"`
	Pool     string `json:"pool,omitempty"`
	Target   string `json:"target,omitempty"`
	SnapName string `json:"snapname,omitempty"`
}

// QEMUMigrateParams 迁移虚拟机参数
type QEMUMigrateParams struct {
	Target           string `json:"target"`
	Online           int    `json:"online,omitempty"`
	MigrationNetwork string `json:"migration_network,omitempty"`
	MigrationType    string `json:"migration_type,omitempty"`
	BW               int    `json:"bw,omitempty"`
}

// QEMUConfigParams 更新虚拟机配置参数
type QEMUConfigParams map[string]interface{}

// ============================================================
// LXC 容器类型
// ============================================================

// LXCContainer LXC 容器简要信息（列表返回）
type LXCContainer struct {
	VMID      int      `json:"vmid"`
	Name      string   `json:"name,omitempty"`
	Status    string   `json:"status,omitempty"`
	CPU       float64  `json:"cpu,omitempty"`
	CPUs      int      `json:"cpus,omitempty"`
	Mem       uint64   `json:"mem,omitempty"`
	MaxMem    uint64   `json:"maxmem,omitempty"`
	DiskRead  uint64   `json:"diskread,omitempty"`
	DiskWrite uint64   `json:"diskwrite,omitempty"`
	NetIn     uint64   `json:"netin,omitempty"`
	NetOut    uint64   `json:"netout,omitempty"`
	Uptime    uint64   `json:"uptime,omitempty"`
	Template  int      `json:"template,omitempty"`
	Tags      string   `json:"tags,omitempty"`
	PID       int      `json:"pid,omitempty"`
	Node      string   `json:"node,omitempty"`
	Type      string   `json:"type,omitempty"`
	Lock      string   `json:"lock,omitempty"`
	Swap      uint64   `json:"swap,omitempty"`
	MaxSwap   uint64   `json:"maxswap,omitempty"`
	HA        *HAState `json:"ha,omitempty"`
}

// LXCConfig LXC 容器完整配置
type LXCConfig struct {
	VMID         int    `json:"vmid"`
	Arch         string `json:"arch,omitempty"`
	Backup       string `json:"backup,omitempty"`
	Console      int    `json:"console,omitempty"`
	Cores        int    `json:"cores,omitempty"`
	CPULimit     int    `json:"cpulimit,omitempty"`
	CPUUnits     int    `json:"cpuunits,omitempty"`
	Description  string `json:"description,omitempty"`
	Features     string `json:"features,omitempty"`
	Force        int    `json:"force,omitempty"`
	HookScript   string `json:"hookscript,omitempty"`
	Hostname     string `json:"hostname,omitempty"`
	Lock         string `json:"lock,omitempty"`
	Memory       int    `json:"memory,omitempty"`
	Net0         string `json:"net0,omitempty"`
	Net1         string `json:"net1,omitempty"`
	Net2         string `json:"net2,omitempty"`
	OnBoot       int    `json:"onboot,omitempty"`
	OSType       string `json:"ostype,omitempty"`
	RootFs       string `json:"rootfs,omitempty"`
	Startup      string `json:"startup,omitempty"`
	Swap         int    `json:"swap,omitempty"`
	Template     int    `json:"template,omitempty"`
	TTY          int    `json:"tty,omitempty"`
	Unprivileged int    `json:"unprivileged,omitempty"`
	CIUser       string `json:"ciuser,omitempty"`
	CIPassword   string `json:"cipassword,omitempty"`
	SSHKeys      string `json:"sshkeys,omitempty"`
	Nameserver   string `json:"nameserver,omitempty"`
	Searchdomain string `json:"searchdomain,omitempty"`
	IPConfig0    string `json:"ipconfig0,omitempty"`
	Tags         string `json:"tags,omitempty"`
	Protection   int    `json:"protection,omitempty"`
	MountPoint0  string `json:"mp0,omitempty"`
	MountPoint1  string `json:"mp1,omitempty"`
	MountPoint2  string `json:"mp2,omitempty"`
	MountPoint3  string `json:"mp3,omitempty"`
	MountPoint4  string `json:"mp4,omitempty"`
	MountPoint5  string `json:"mp5,omitempty"`
	MountPoint6  string `json:"mp6,omitempty"`
	MountPoint7  string `json:"mp7,omitempty"`
	MountPoint8  string `json:"mp8,omitempty"`
	MountPoint9  string `json:"mp9,omitempty"`
}

// LXCCreateParams 创建 LXC 容器的参数
type LXCCreateParams struct {
	VMID          int    `json:"vmid"`
	Node          string `json:"node,omitempty"`
	OSTemplate    string `json:"ostemplate,omitempty"`
	RootFs        string `json:"rootfs"`
	Hostname      string `json:"hostname,omitempty"`
	Memory        int    `json:"memory,omitempty"`
	Swap          int    `json:"swap,omitempty"`
	Cores         int    `json:"cores,omitempty"`
	CPUUnits      int    `json:"cpuunits,omitempty"`
	Net0          string `json:"net0,omitempty"`
	Password      string `json:"password,omitempty"`
	Storage       string `json:"storage,omitempty"`
	Console       int    `json:"console,omitempty"`
	Description   string `json:"description,omitempty"`
	Features      string `json:"features,omitempty"`
	Nameserver    string `json:"nameserver,omitempty"`
	Searchdomain  string `json:"searchdomain,omitempty"`
	SSHPublicKeys string `json:"ssh-public-keys,omitempty"`
	Tags          string `json:"tags,omitempty"`
	Unprivileged  int    `json:"unprivileged,omitempty"`
	Pool          string `json:"pool,omitempty"`
}

// LXCConfigParams 更新 LXC 容器配置参数
type LXCConfigParams map[string]interface{}

// LXCCloneParams 克隆 LXC 容器参数
type LXCCloneParams struct {
	NewID    int    `json:"newid"`
	Name     string `json:"name,omitempty"`
	Full     int    `json:"full,omitempty"`
	Pool     string `json:"pool,omitempty"`
	Target   string `json:"target,omitempty"`
	SnapName string `json:"snapname,omitempty"`
	Storage  string `json:"storage,omitempty"`
}

// LXCMigrateParams 迁移 LXC 容器参数
type LXCMigrateParams struct {
	Target  string `json:"target"`
	Online  int    `json:"online,omitempty"`
	Restart int    `json:"restart,omitempty"`
}

// ============================================================
// 节点管理类型
// ============================================================

// NodeStatus 节点状态信息
type NodeStatus struct {
	Node       string  `json:"node"`
	Status     string  `json:"status"`
	CPU        float64 `json:"cpu"`
	CPUs       int     `json:"cpus"`
	Mem        uint64  `json:"mem"`
	MaxMem     uint64  `json:"maxmem"`
	Swap       uint64  `json:"swap"`
	MaxSwap    uint64  `json:"maxswap"`
	Disk       uint64  `json:"disk"`
	MaxDisk    uint64  `json:"maxdisk"`
	KVersion   string  `json:"kversion,omitempty"`
	PVEVersion string  `json:"pveversion,omitempty"`
	LoadAvg    string  `json:"loadavg,omitempty"`
	Uptime     uint64  `json:"uptime"`
	Idle       int     `json:"idle,omitempty"`
	RootFs     struct {
		Free  uint64 `json:"free"`
		Total uint64 `json:"total"`
		Used  uint64 `json:"used"`
	} `json:"rootfs,omitempty"`
	SwapInfo struct {
		Free  uint64 `json:"free"`
		Total uint64 `json:"total"`
		Used  uint64 `json:"used"`
	} `json:"swap,omitempty"`
}

// VersionInfo PVE 版本信息
type VersionInfo struct {
	Release string `json:"release"`
	Repoid  string `json:"repoid"`
	Version string `json:"version"`
}

// Service 节点服务信息
type Service struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	State       string `json:"state,omitempty"`
	Desc        string `json:"desc"`
	Node        string `json:"node,omitempty"`
	Port        int    `json:"port,omitempty"`
	EnableState string `json:"enable-state,omitempty"`
	Watchdog    string `json:"watchdog,omitempty"`
	Error       string `json:"error,omitempty"`
}

// LogEntry 日志条目
type LogEntry struct {
	N    int    `json:"n"`
	Text string `json:"t"`
	Prio string `json:"p,omitempty"`
	Dev  string `json:"dev,omitempty"`
	Time uint64 `json:"time,omitempty"`
}

// Task 任务信息
type Task struct {
	ID        string `json:"id"`
	PID       uint64 `json:"pid,omitempty"`
	Node      string `json:"node,omitempty"`
	StartTime uint64 `json:"starttime"`
	User      string `json:"user"`
	Type      string `json:"type"`
	Status    string `json:"status,omitempty"`
	ExitCode  string `json:"exitstatus,omitempty"`
	EndTime   uint64 `json:"endtime,omitempty"`
	UPID      string `json:"upid,omitempty"`
}

// TaskStatus 任务状态
type TaskStatus struct {
	ID        string `json:"id"`
	Node      string `json:"node"`
	Type      string `json:"type"`
	User      string `json:"user"`
	Status    string `json:"status"`
	ExitCode  string `json:"exitstatus"`
	StartTime uint64 `json:"starttime"`
	EndTime   uint64 `json:"endtime"`
}

// TaskLogLine 任务日志行
type TaskLogLine struct {
	N    int    `json:"n"`
	Text string `json:"t"`
}

// NetInterface 网络接口信息
type NetInterface struct {
	Iface     string `json:"iface"`
	Type      string `json:"type"`
	Active    int    `json:"active,omitempty"`
	Address   string `json:"address,omitempty"`
	Netmask   int    `json:"netmask,omitempty"`
	Gateway   string `json:"gateway,omitempty"`
	Autostart int    `json:"autostart,omitempty"`
	Method    string `json:"method,omitempty"`
	IPv6      string `json:"ipv6,omitempty"`
	IPv6CIDR  int    `json:"cidr6,omitempty"`
	Gateway6  string `json:"gw6,omitempty"`
	BondMode  string `json:"bond_mode,omitempty"`
	Slaves    string `json:"slaves,omitempty"`
	VlanId    int    `json:"vlan_id,omitempty"`
	MTU       int    `json:"mtu,omitempty"`
	Priority  int    `json:"priority,omitempty"`
}

// NetInterfaceConfig 网络接口配置参数
type NetInterfaceConfig map[string]interface{}

// PackageUpdate 软件包更新信息
type PackageUpdate struct {
	Name    string `json:"package"`
	Title   string `json:"title,omitempty"`
	Version string `json:"version,omitempty"`
	Prio    string `json:"prio,omitempty"`
	Source  string `json:"source,omitempty"`
}

// DNSConfig DNS 配置
type DNSConfig struct {
	Search     string `json:"search,omitempty"`
	Nameserver string `json:"nameserver,omitempty"`
}

// TimeInfo 时间信息
type TimeInfo struct {
	Timezone  string `json:"timezone,omitempty"`
	Time      uint64 `json:"time"`
	Localtime uint64 `json:"localtime"`
}

// ============================================================
// 存储管理类型
// ============================================================

// Storage 存储简要信息
type Storage struct {
	Storage string `json:"storage"`
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Shared  int    `json:"shared,omitempty"`
	Enabled int    `json:"enabled,omitempty"`
	Active  int    `json:"active,omitempty"`
}

// StorageStatus 存储状态详情
type StorageStatus struct {
	Storage string `json:"storage"`
	Type    string `json:"type"`
	Content string `json:"content"`
	Active  int    `json:"active"`
	Enabled int    `json:"enabled"`
	Shared  int    `json:"shared"`
	Total   uint64 `json:"total"`
	Used    uint64 `json:"used"`
	Avail   uint64 `json:"avail"`
}

// StorageContent 存储内容条目
type StorageContent struct {
	Volid       string `json:"volid"`
	ContentType string `json:"content"`
	Format      string `json:"format,omitempty"`
	Size        uint64 `json:"size,omitempty"`
	VSize       uint64 `json:"vsize,omitempty"`
	Used        uint64 `json:"used,omitempty"`
	Name        string `json:"name,omitempty"`
	VMID        int    `json:"vmid,omitempty"`
	CTime       uint64 `json:"ctime,omitempty"`
	Parent      string `json:"parent,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

// StorageCreateParams 创建存储参数
type StorageCreateParams map[string]interface{}

// StorageUpdateParams 更新存储参数
type StorageUpdateParams map[string]interface{}

// ============================================================
// 集群管理类型
// ============================================================

// ClusterResource 集群资源
type ClusterResource struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	CGroup    string  `json:"cgroup,omitempty"`
	CPU       float64 `json:"cpu,omitempty"`
	HA        int     `json:"h,omitempty"`
	Level     string  `json:"level,omitempty"`
	MaxCPU    int     `json:"maxcpu,omitempty"`
	MaxDisk   uint64  `json:"maxdisk,omitempty"`
	MaxMem    uint64  `json:"maxmem,omitempty"`
	Mem       uint64  `json:"mem,omitempty"`
	Name      string  `json:"name,omitempty"`
	Node      string  `json:"node,omitempty"`
	PlugLevel string  `json:"pluglevel,omitempty"`
	PlugState string  `json:"plugstate,omitempty"`
	Pool      string  `json:"pool,omitempty"`
	Status    string  `json:"status,omitempty"`
	Storage   string  `json:"storage,omitempty"`
	Template  int     `json:"template,omitempty"`
	Uptime    uint64  `json:"uptime,omitempty"`
	VMID      int     `json:"vmid,omitempty"`
	Tags      string  `json:"tags,omitempty"`
}

// ClusterTask 集群任务
type ClusterTask struct {
	ID        string `json:"id"`
	Node      string `json:"node"`
	Type      string `json:"type"`
	User      string `json:"user"`
	Status    string `json:"status,omitempty"`
	ExitCode  string `json:"exitstatus,omitempty"`
	StartTime uint64 `json:"starttime"`
	EndTime   uint64 `json:"endtime,omitempty"`
	UPID      string `json:"upid,omitempty"`
}

// NextVMID 下一个可用的 VMID
type NextVMID struct {
	NextID int `json:"nextid"`
}

// HAConfig HA 配置
type HAConfig struct {
	Groups    []HAGroup    `json:"groups"`
	Resources []HAResource `json:"resources"`
	Fence     []HAFence    `json:"fence,omitempty"`
}

// HAGroup HA 组
type HAGroup struct {
	Group      string   `json:"group"`
	Comment    string   `json:"comment,omitempty"`
	Nodelist   []string `json:"nodes"`
	Restrict   int      `json:"restricted,omitempty"`
	NoFailback int      `json:"nofailback,omitempty"`
}

// HAResource HA 资源
type HAResource struct {
	SID         string `json:"sid"`
	Type        string `json:"type"`
	Group       string `json:"group,omitempty"`
	Comment     string `json:"comment,omitempty"`
	MaxRelocate int    `json:"max_relocate,omitempty"`
	MaxRestart  int    `json:"max_restart,omitempty"`
	Exclusive   int    `json:"exclusive,omitempty"`
	Prefered    int    `json:"prefered,omitempty"`
	State       string `json:"state,omitempty"`
}

// HAFence HA 隔离配置
type HAFence struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Port   string `json:"port,omitempty"`
	Target string `json:"target,omitempty"`
}

// SDNZone SDN 区域
type SDNZone struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Nodes   string `json:"nodes,omitempty"`
	Comment string `json:"comment,omitempty"`
}

// SDNVNET SDN 虚拟网络
type SDNVNET struct {
	ID      string `json:"id"`
	Zone    string `json:"zone"`
	VNI     int    `json:"vni,omitempty"`
	Comment string `json:"comment,omitempty"`
	Tag     int    `json:"tag,omitempty"`
}

// Pool 池简要信息
type Pool struct {
	PoolID  string `json:"poolid"`
	Comment string `json:"comment,omitempty"`
}

// PoolDetail 池详细信息
type PoolDetail struct {
	PoolID  string       `json:"poolid"`
	Comment string       `json:"comment,omitempty"`
	Members []PoolMember `json:"members"`
}

// PoolMember 池成员
type PoolMember struct {
	Node    string `json:"node"`
	Type    string `json:"type"`
	VMID    int    `json:"vmid,omitempty"`
	Storage string `json:"storage,omitempty"`
}

// ============================================================
// 访问控制类型
// ============================================================

// User 用户信息
type User struct {
	UserID    string   `json:"userid"`
	Comment   string   `json:"comment,omitempty"`
	Email     string   `json:"email,omitempty"`
	Enable    int      `json:"enable"`
	Expire    int      `json:"expire,omitempty"`
	Firstname string   `json:"firstname,omitempty"`
	Lastname  string   `json:"lastname,omitempty"`
	Groups    []string `json:"groups,omitempty"`
	Keys      string   `json:"keys,omitempty"`
	Tokens    string   `json:"tokens,omitempty"`
	RealmType string   `json:"realm-type,omitempty"`
}

// UserCreateParams 创建用户参数
type UserCreateParams struct {
	UserID    string `json:"userid"`
	Comment   string `json:"comment,omitempty"`
	Email     string `json:"email,omitempty"`
	Enable    int    `json:"enable,omitempty"`
	Expire    int    `json:"expire,omitempty"`
	Firstname string `json:"firstname,omitempty"`
	Lastname  string `json:"lastname,omitempty"`
	Groups    string `json:"groups,omitempty"`
	Password  string `json:"password,omitempty"`
	Keys      string `json:"keys,omitempty"`
}

// Group 组信息
type Group struct {
	GroupID string   `json:"groupid"`
	Comment string   `json:"comment,omitempty"`
	Users   []string `json:"users,omitempty"`
}

// GroupCreateParams 创建组参数
type GroupCreateParams struct {
	GroupID string `json:"groupid"`
	Comment string `json:"comment,omitempty"`
	Users   string `json:"users,omitempty"`
}

// Role 角色信息
type Role struct {
	RoleID string   `json:"roleid"`
	Privs  []string `json:"privs,omitempty"`
}

// RoleCreateParams 创建角色参数
type RoleCreateParams struct {
	RoleID string `json:"roleid"`
	Privs  string `json:"privs"`
}

// ACL 访问控制列表
type ACL struct {
	Path      string `json:"path"`
	RoleID    string `json:"roleid"`
	Type      string `json:"type"`
	UserID    string `json:"userid,omitempty"`
	GroupID   string `json:"groupid,omitempty"`
	Propagate int    `json:"propagate,omitempty"`
}

// ACLParams 设置 ACL 参数
type ACLParams struct {
	Path      string `json:"path"`
	Roles     string `json:"roles"`
	Users     string `json:"users,omitempty"`
	Groups    string `json:"groups,omitempty"`
	Propagate int    `json:"propagate,omitempty"`
	Delete    int    `json:"delete,omitempty"`
}

// AuthDomain 认证域信息
type AuthDomain struct {
	Realm   string `json:"realm"`
	Type    string `json:"type"`
	Comment string `json:"comment,omitempty"`
	Default int    `json:"default,omitempty"`
}

// ============================================================
// 快照类型
// ============================================================

// Snapshot 快照信息
type Snapshot struct {
	Name        string `json:"name"`
	VMASize     uint64 `json:"vmstate-size,omitempty"`
	Description string `json:"description,omitempty"`
	SnapTime    uint64 `json:"snaptime,omitempty"`
	Parent      string `json:"parent,omitempty"`
	Children    string `json:"children,omitempty"`
}

// ============================================================
// 待处理配置类型
// ============================================================

// PendingConfig 待处理配置
type PendingConfig struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Delete string `json:"delete,omitempty"`
	Digest string `json:"digest,omitempty"`
}

// ============================================================
// RRD 性能数据类型
// ============================================================

// RRDData RRD 性能图数据
type RRDData struct {
	DataSource string     `json:"datasource"`
	Timeframe  string     `json:"timeframe"`
	Data       []RRDPoint `json:"data"`
}

// RRDPoint RRD 数据点
type RRDPoint struct {
	Time   uint64  `json:"time"`
	CPU    float64 `json:"cpu,omitempty"`
	Mem    float64 `json:"mem,omitempty"`
	Disk   float64 `json:"disk,omitempty"`
	NetIn  float64 `json:"netin,omitempty"`
	NetOut float64 `json:"netout,omitempty"`
}

// ============================================================
// 辅助类型
// ============================================================

// HAState HA 状态信息
type HAState struct {
	Managed int    `json:"managed"`
	State   string `json:"state,omitempty"`
}

// APIError PVE API 错误响应
type APIError struct {
	Errors  interface{} `json:"errors,omitempty"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// VNCConfig VNC 代理配置
type VNCConfig struct {
	Ticket     string `json:"ticket"`
	Port       int    `json:"port"`
	Node       string `json:"node"`
	VMID       int    `json:"vmid"`
	VNCTicket  string `json:"vncticket"`
	VMName     string `json:"vmname"`
	Upid       string `json:"upid"`
	User       string `json:"user"`
	IsTemplate int    `json:"isTemplate"`
}

// ============================================================
// 向后兼容类型（用于旧代码）
// ============================================================

// ClusterStatus 集群状态信息（兼容旧代码）
type ClusterStatus struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// NodeInfo 节点信息（兼容旧代码）
type NodeInfo struct {
	Node    string  `json:"node"`
	Status  string  `json:"status"`
	CPU     float64 `json:"cpu"`
	Mem     uint64  `json:"mem"`
	MaxMem  uint64  `json:"maxmem"`
	Disk    uint64  `json:"disk"`
	MaxDisk uint64  `json:"maxdisk"`
	Level   string  `json:"level"`
}

// VMInfo 虚拟机信息（兼容旧代码）
type VMInfo = QEMUVM

// LXCInfo LXC 容器信息（兼容旧代码）
type LXCInfo = LXCContainer

// StorageInfo 存储信息（兼容旧代码）
type StorageInfo = Storage
