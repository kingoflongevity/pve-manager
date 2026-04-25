package pve

// APIResponse PVE API 通用响应结构
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

// ClusterStatus 集群状态信息
type ClusterStatus struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// NodeInfo 节点信息
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

// VMInfo 虚拟机信息
type VMInfo struct {
	VMID      int     `json:"vmid"`
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	CPU       float64 `json:"cpu"`
	Mem       uint64  `json:"mem"`
	MaxMem    uint64  `json:"maxmem"`
	DiskRead  uint64  `json:"diskread"`
	DiskWrite uint64  `json:"diskwrite"`
	NetIn     uint64  `json:"netin"`
	NetOut    uint64  `json:"netout"`
	Template  int     `json:"template"`
	Type      string  `json:"type"` // qemu, lxc
	Node      string  `json:"node"`
	Uptime    uint64  `json:"uptime"`
	CPUs      int     `json:"cpus"`
}

// LXCInfo LXC 容器信息
type LXCInfo struct {
	VMID      int     `json:"vmid"`
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	CPU       float64 `json:"cpu"`
	Mem       uint64  `json:"mem"`
	MaxMem    uint64  `json:"maxmem"`
	DiskRead  uint64  `json:"diskread"`
	DiskWrite uint64  `json:"diskwrite"`
	NetIn     uint64  `json:"netin"`
	NetOut    uint64  `json:"netout"`
	Template  int     `json:"template"`
	Type      string  `json:"type"`
	Node      string  `json:"node"`
	Uptime    uint64  `json:"uptime"`
}

// StorageInfo 存储信息
type StorageInfo struct {
	Storage string `json:"storage"`
	Active  int    `json:"active"`
	Enabled int    `json:"enabled"`
	Shared  int    `json:"shared"`
	Total   uint64 `json:"total"`
	Used    uint64 `json:"used"`
	Avail   uint64 `json:"avail"`
	Type    string `json:"type"`
	Content string `json:"content"`
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
