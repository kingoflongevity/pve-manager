package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AccessHandler struct {
	logger *zap.Logger
}

func NewAccessHandler(logger *zap.Logger) *AccessHandler {
	return &AccessHandler{logger: logger}
}

var mockUsers = []map[string]interface{}{
	{"userid": "root@pam", "username": "root", "realm": "pam", "firstname": "", "lastname": "", "email": "", "enabled": 1},
	{"userid": "admin@pve", "username": "admin", "realm": "pve", "firstname": "Admin", "lastname": "", "email": "admin@pve.local", "enabled": 1},
	{"userid": "operator@pve", "username": "operator", "realm": "pve", "firstname": "Operator", "lastname": "User", "email": "op@pve.local", "enabled": 1},
	{"userid": "viewer@pve", "username": "viewer", "realm": "pve", "firstname": "View", "lastname": "Only", "email": "", "enabled": 0},
}

var mockGroups = []map[string]interface{}{
	{"groupid": "admins", "comment": "系统管理员组", "users": []string{"root@pam", "admin@pve"}},
	{"groupid": "operators", "comment": "运维操作组", "users": []string{"operator@pve"}},
	{"groupid": "viewers", "comment": "只读观察组", "users": []string{"viewer@pve"}},
}

var mockRoles = []map[string]interface{}{
	{"roleid": "Administrator", "privs": "Administrator"},
	{"roleid": "PVEAdmin", "privs": "Administrator,Audit,Sys.Audit,Sys.Modify,Sys.PowerMgmt,Sys.Console,Sys.Syslog"},
	{"roleid": "PVEAuditor", "privs": "Audit,Sys.Audit,Sys.Syslog,Datastore.Audit"},
	{"roleid": "PVEVMAdmin", "privs": "VM.Audit,VM.PowerMgmt,VM.Config.CPU,VM.Config.Memory,VM.Config.Disk,VM.Config.Network,VM.Config.Options,VM.Backup,VM.Console,VM.Monitor,VM.Snapshot"},
	{"roleid": "PVEVMUser", "privs": "VM.Audit,VM.PowerMgmt,VM.Console"},
	{"roleid": "NoAccess", "privs": ""},
}

var mockACLs = []map[string]interface{}{
	{"path": "/", "type": "user", "auth_id": "root@pam", "roleid": "Administrator", "propagate": 1},
	{"path": "/", "type": "group", "auth_id": "admins", "roleid": "Administrator", "propagate": 1},
	{"path": "/nodes", "type": "group", "auth_id": "operators", "roleid": "PVEAdmin", "propagate": 1},
	{"path": "/vms", "type": "group", "auth_id": "viewers", "roleid": "PVEAuditor", "propagate": 1},
}

var mockDomains = []map[string]interface{}{
	{"realm": "pam", "type": "pam", "comment": "Linux PAM 标准认证", "default": 0},
	{"realm": "pve", "type": "pve", "comment": "Proxmox VE 认证服务器", "default": 1},
	{"realm": "ldap", "type": "ldap", "comment": "LDAP 目录服务", "default": 0},
}

func (h *AccessHandler) ListUsers(c *gin.Context) {
	h.success(c, mockUsers)
}

func (h *AccessHandler) CreateUser(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		h.badRequest(c, "参数错误")
		return
	}
	newUser := map[string]interface{}{
		"userid":    body["userid"],
		"username":  "",
		"realm":     "pve",
		"firstname": body["firstname"],
		"lastname":  body["lastname"],
		"email":     body["email"],
		"enabled":   1,
	}
	if uid, ok := body["userid"].(string); ok {
		for i, ch := range uid {
			if ch == '@' {
				newUser["username"] = uid[:i]
				newUser["realm"] = uid[i+1:]
				break
			}
		}
		if newUser["username"] == "" {
			newUser["username"] = uid
		}
	}
	mockUsers = append(mockUsers, newUser)
	h.success(c, newUser)
}

func (h *AccessHandler) GetUser(c *gin.Context) {
	userid := c.Param("userid")
	for _, u := range mockUsers {
		if u["userid"] == userid {
			h.success(c, u)
			return
		}
	}
	h.notFound(c, "用户不存在")
}

func (h *AccessHandler) UpdateUser(c *gin.Context) {
	userid := c.Param("userid")
	for i, u := range mockUsers {
		if u["userid"] == userid {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				h.badRequest(c, "参数错误")
				return
			}
			for k, v := range body {
				mockUsers[i][k] = v
			}
			h.success(c, mockUsers[i])
			return
		}
	}
	h.notFound(c, "用户不存在")
}

func (h *AccessHandler) DeleteUser(c *gin.Context) {
	userid := c.Param("userid")
	for i, u := range mockUsers {
		if u["userid"] == userid {
			mockUsers = append(mockUsers[:i], mockUsers[i+1:]...)
			h.success(c, nil)
			return
		}
	}
	h.notFound(c, "用户不存在")
}

func (h *AccessHandler) UpdatePassword(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		h.badRequest(c, "参数错误")
		return
	}
	h.success(c, nil)
}

func (h *AccessHandler) ListGroups(c *gin.Context) {
	h.success(c, mockGroups)
}

func (h *AccessHandler) CreateGroup(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		h.badRequest(c, "参数错误")
		return
	}
	users := []string{}
	if u, ok := body["users"].([]interface{}); ok {
		for _, v := range u {
			if s, ok := v.(string); ok {
				users = append(users, s)
			}
		}
	}
	newGroup := map[string]interface{}{
		"groupid": body["groupid"],
		"comment": body["comment"],
		"users":   users,
	}
	mockGroups = append(mockGroups, newGroup)
	h.success(c, newGroup)
}

func (h *AccessHandler) GetGroup(c *gin.Context) {
	groupid := c.Param("groupid")
	for _, g := range mockGroups {
		if g["groupid"] == groupid {
			h.success(c, g)
			return
		}
	}
	h.notFound(c, "组不存在")
}

func (h *AccessHandler) UpdateGroup(c *gin.Context) {
	groupid := c.Param("groupid")
	for i, g := range mockGroups {
		if g["groupid"] == groupid {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				h.badRequest(c, "参数错误")
				return
			}
			for k, v := range body {
				mockGroups[i][k] = v
			}
			h.success(c, mockGroups[i])
			return
		}
	}
	h.notFound(c, "组不存在")
}

func (h *AccessHandler) DeleteGroup(c *gin.Context) {
	groupid := c.Param("groupid")
	for i, g := range mockGroups {
		if g["groupid"] == groupid {
			mockGroups = append(mockGroups[:i], mockGroups[i+1:]...)
			h.success(c, nil)
			return
		}
	}
	h.notFound(c, "组不存在")
}

func (h *AccessHandler) ListRoles(c *gin.Context) {
	h.success(c, mockRoles)
}

func (h *AccessHandler) CreateRole(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		h.badRequest(c, "参数错误")
		return
	}
	newRole := map[string]interface{}{
		"roleid": body["roleid"],
		"privs":  body["privs"],
	}
	mockRoles = append(mockRoles, newRole)
	h.success(c, newRole)
}

func (h *AccessHandler) GetRole(c *gin.Context) {
	roleid := c.Param("roleid")
	for _, r := range mockRoles {
		if r["roleid"] == roleid {
			h.success(c, r)
			return
		}
	}
	h.notFound(c, "角色不存在")
}

func (h *AccessHandler) UpdateRole(c *gin.Context) {
	roleid := c.Param("roleid")
	for i, r := range mockRoles {
		if r["roleid"] == roleid {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				h.badRequest(c, "参数错误")
				return
			}
			for k, v := range body {
				mockRoles[i][k] = v
			}
			h.success(c, mockRoles[i])
			return
		}
	}
	h.notFound(c, "角色不存在")
}

func (h *AccessHandler) DeleteRole(c *gin.Context) {
	roleid := c.Param("roleid")
	for i, r := range mockRoles {
		if r["roleid"] == roleid {
			mockRoles = append(mockRoles[:i], mockRoles[i+1:]...)
			h.success(c, nil)
			return
		}
	}
	h.notFound(c, "角色不存在")
}

func (h *AccessHandler) ListACLs(c *gin.Context) {
	h.success(c, mockACLs)
}

func (h *AccessHandler) SetACL(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		h.badRequest(c, "参数错误")
		return
	}

	if del, ok := body["delete"]; ok {
		delVal := 0
		switch v := del.(type) {
		case float64:
			delVal = int(v)
		case string:
			delVal, _ = strconv.Atoi(v)
		}
		if delVal == 1 {
			for i, a := range mockACLs {
				if a["path"] == body["path"] && a["auth_id"] == body["auth_id"] && a["roleid"] == body["roles"] {
					mockACLs = append(mockACLs[:i], mockACLs[i+1:]...)
					break
				}
			}
			h.success(c, nil)
			return
		}
	}

	aclType := "user"
	if _, ok := body["users"]; !ok {
		aclType = "group"
	}
	authID := ""
	if id, ok := body["users"].(string); ok {
		authID = id
	} else if id, ok := body["groups"].(string); ok {
		authID = id
	} else if id, ok := body["auth_id"].(string); ok {
		authID = id
	}
	propagate := 0
	if p, ok := body["propagate"]; ok {
		switch v := p.(type) {
		case float64:
			propagate = int(v)
		case int:
			propagate = v
		}
	}

	newACL := map[string]interface{}{
		"path":      body["path"],
		"type":      aclType,
		"auth_id":   authID,
		"roleid":    body["roles"],
		"propagate": propagate,
	}
	mockACLs = append(mockACLs, newACL)
	h.success(c, newACL)
}

func (h *AccessHandler) ListDomains(c *gin.Context) {
	h.success(c, mockDomains)
}

func (h *AccessHandler) GetDomain(c *gin.Context) {
	realm := c.Param("realm")
	for _, d := range mockDomains {
		if d["realm"] == realm {
			h.success(c, d)
			return
		}
	}
	h.notFound(c, "认证域不存在")
}

func (h *AccessHandler) success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

func (h *AccessHandler) badRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code":    400,
		"message": msg,
	})
}

func (h *AccessHandler) notFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, gin.H{
		"code":    404,
		"message": msg,
	})
}
