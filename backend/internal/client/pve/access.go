package pve

import (
	"context"
	"fmt"
	"net/url"
)

// ListUsers 获取所有用户列表
// 返回系统中所有用户的信息
// 使用 interface{} 返回值防止 PVE 9.x 返回意外数据格式时 json.Unmarshal 失败
func (c *Client) ListUsers(ctx context.Context) (interface{}, error) {
	var result interface{}
	if err := c.Get(ctx, "access/users", &result); err != nil {
		return nil, fmt.Errorf("获取用户列表失败: %w", err)
	}
	return result, nil
}

// GetUser 获取指定用户的详细信息
// userid: 用户 ID (格式: username@realm)
// 使用 interface{} 返回值防止 PVE 9.x 返回意外数据格式时 json.Unmarshal 失败
func (c *Client) GetUser(ctx context.Context, userid string) (interface{}, error) {
	var result interface{}
	path := fmt.Sprintf("access/users/%s", userid)
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	return result, nil
}

// CreateUser 创建新用户
// params: 用户创建参数
// 返回异步任务 ID (UPID)
func (c *Client) CreateUser(ctx context.Context, params *UserCreateParams) (string, error) {
	var upid string
	if err := c.Post(ctx, "access/users", params, &upid); err != nil {
		return "", fmt.Errorf("创建用户失败: %w", err)
	}
	return upid, nil
}

// UpdateUser 更新用户信息
// userid: 用户 ID, params: 更新参数
// 返回异步任务 ID (UPID)
func (c *Client) UpdateUser(ctx context.Context, userid string, params map[string]interface{}) (string, error) {
	var upid string
	path := fmt.Sprintf("access/users/%s", userid)
	if err := c.Put(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("更新用户失败: %w", err)
	}
	return upid, nil
}

// DeleteUser 删除用户
// userid: 用户 ID
// 返回异步任务 ID (UPID)
func (c *Client) DeleteUser(ctx context.Context, userid string) (string, error) {
	var upid string
	path := fmt.Sprintf("access/users/%s", userid)
	if err := c.Delete(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("删除用户失败: %w", err)
	}
	return upid, nil
}

// SetUserPassword 设置用户密码
// userid: 用户 ID, password: 新密码
// 返回异步任务 ID (UPID)
func (c *Client) SetUserPassword(ctx context.Context, userid, password string) (string, error) {
	var upid string
	path := fmt.Sprintf("access/password/%s", userid)
	params := map[string]interface{}{"password": password}
	if err := c.Put(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("设置密码失败: %w", err)
	}
	return upid, nil
}

// ListGroups 获取所有组列表
// 返回系统中所有用户组的信息
// 使用 interface{} 返回值防止 PVE 9.x 返回意外数据格式时 json.Unmarshal 失败
func (c *Client) ListGroups(ctx context.Context) (interface{}, error) {
	var result interface{}
	if err := c.Get(ctx, "access/groups", &result); err != nil {
		return nil, fmt.Errorf("获取组列表失败: %w", err)
	}
	return result, nil
}

// CreateGroup 创建新组
// params: 组创建参数
// 返回异步任务 ID (UPID)
func (c *Client) CreateGroup(ctx context.Context, params *GroupCreateParams) (string, error) {
	var upid string
	if err := c.Post(ctx, "access/groups", params, &upid); err != nil {
		return "", fmt.Errorf("创建组失败: %w", err)
	}
	return upid, nil
}

// UpdateGroup 更新组信息
// groupid: 组 ID, params: 更新参数
// 返回异步任务 ID (UPID)
func (c *Client) UpdateGroup(ctx context.Context, groupid string, params map[string]interface{}) (string, error) {
	var upid string
	path := fmt.Sprintf("access/groups/%s", groupid)
	if err := c.Put(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("更新组失败: %w", err)
	}
	return upid, nil
}

// DeleteGroup 删除组
// groupid: 组 ID
// 返回异步任务 ID (UPID)
func (c *Client) DeleteGroup(ctx context.Context, groupid string) (string, error) {
	var upid string
	path := fmt.Sprintf("access/groups/%s", groupid)
	if err := c.Delete(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("删除组失败: %w", err)
	}
	return upid, nil
}

// ListRoles 获取所有角色列表
// 返回系统中所有权限角色的信息
// 使用 interface{} 返回值防止 PVE 9.x 返回意外数据格式时 json.Unmarshal 失败
func (c *Client) ListRoles(ctx context.Context) (interface{}, error) {
	var result interface{}
	if err := c.Get(ctx, "access/roles", &result); err != nil {
		return nil, fmt.Errorf("获取角色列表失败: %w", err)
	}
	return result, nil
}

// CreateRole 创建新角色
// params: 角色创建参数
// 返回异步任务 ID (UPID)
func (c *Client) CreateRole(ctx context.Context, params *RoleCreateParams) (string, error) {
	var upid string
	if err := c.Post(ctx, "access/roles", params, &upid); err != nil {
		return "", fmt.Errorf("创建角色失败: %w", err)
	}
	return upid, nil
}

// UpdateRole 更新角色权限
// roleid: 角色 ID, privs: 权限列表（逗号分隔）
// 返回异步任务 ID (UPID)
func (c *Client) UpdateRole(ctx context.Context, roleid, privs string) (string, error) {
	var upid string
	path := fmt.Sprintf("access/roles/%s", roleid)
	params := map[string]interface{}{"privs": privs}
	if err := c.Put(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("更新角色失败: %w", err)
	}
	return upid, nil
}

// DeleteRole 删除角色
// roleid: 角色 ID
// 返回异步任务 ID (UPID)
func (c *Client) DeleteRole(ctx context.Context, roleid string) (string, error) {
	var upid string
	path := fmt.Sprintf("access/roles/%s", roleid)
	if err := c.Delete(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("删除角色失败: %w", err)
	}
	return upid, nil
}

// ListACLs 获取所有访问控制列表
// 返回系统中所有 ACL 条目
// 使用 interface{} 返回值防止 PVE 9.x 返回意外数据格式时 json.Unmarshal 失败
func (c *Client) ListACLs(ctx context.Context) (interface{}, error) {
	var result interface{}
	if err := c.Get(ctx, "access/acl", &result); err != nil {
		return nil, fmt.Errorf("获取 ACL 列表失败: %w", err)
	}
	return result, nil
}

// SetACL 设置访问控制
// params: ACL 设置参数
// 返回异步任务 ID (UPID)
func (c *Client) SetACL(ctx context.Context, params *ACLParams) (string, error) {
	var upid string
	if err := c.Put(ctx, "access/acl", params, &upid); err != nil {
		return "", fmt.Errorf("设置 ACL 失败: %w", err)
	}
	return upid, nil
}

// ListDomains 获取所有认证域列表
// 返回系统中所有认证域（PVE、PAM、LDAP 等）
// 使用 interface{} 返回值防止 PVE 9.x 返回意外数据格式时 json.Unmarshal 失败
func (c *Client) ListDomains(ctx context.Context) (interface{}, error) {
	var result interface{}
	if err := c.Get(ctx, "access/domains", &result); err != nil {
		return nil, fmt.Errorf("获取认证域列表失败: %w", err)
	}
	return result, nil
}

// GetDomain 获取指定认证域的详细信息
// realm: 认证域 ID
// 使用 interface{} 返回值防止 PVE 9.x 返回意外数据格式时 json.Unmarshal 失败
func (c *Client) GetDomain(ctx context.Context, realm string) (interface{}, error) {
	var result interface{}
	path := fmt.Sprintf("access/domains/%s", realm)
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("获取认证域信息失败: %w", err)
	}
	return result, nil
}

// CreateDomain 创建认证域
// params: 认证域创建参数
// 返回异步任务 ID (UPID)
func (c *Client) CreateDomain(ctx context.Context, params map[string]interface{}) (string, error) {
	var upid string
	if err := c.Post(ctx, "access/domains", params, &upid); err != nil {
		return "", fmt.Errorf("创建认证域失败: %w", err)
	}
	return upid, nil
}

// UpdateDomain 更新认证域配置
// realm: 认证域 ID, params: 更新参数
// 返回异步任务 ID (UPID)
func (c *Client) UpdateDomain(ctx context.Context, realm string, params map[string]interface{}) (string, error) {
	var upid string
	path := fmt.Sprintf("access/domains/%s", realm)
	if err := c.Put(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("更新认证域失败: %w", err)
	}
	return upid, nil
}

// DeleteDomain 删除认证域
// realm: 认证域 ID
// 返回异步任务 ID (UPID)
func (c *Client) DeleteDomain(ctx context.Context, realm string) (string, error) {
	var upid string
	path := fmt.Sprintf("access/domains/%s", realm)
	if err := c.Delete(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("删除认证域失败: %w", err)
	}
	return upid, nil
}

// ListAPITokens 获取指定用户的 API Token 列表
// userid: 用户 ID
// 返回该用户的所有 API Token
func (c *Client) ListAPITokens(ctx context.Context, userid string) ([]map[string]interface{}, error) {
	var tokens []map[string]interface{}
	path := fmt.Sprintf("access/users/%s/token", userid)
	if err := c.Get(ctx, path, &tokens); err != nil {
		return nil, fmt.Errorf("获取 API Token 列表失败: %w", err)
	}
	return tokens, nil
}

// CreateAPIToken 创建 API Token
// userid: 用户 ID, tokenid: Token ID, params: 创建参数
func (c *Client) CreateAPIToken(ctx context.Context, userid, tokenid string, params map[string]interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("access/users/%s/token/%s", userid, tokenid)
	if err := c.Put(ctx, path, params, &result); err != nil {
		return nil, fmt.Errorf("创建 API Token 失败: %w", err)
	}
	return result, nil
}

// DeleteAPIToken 删除 API Token
// userid: 用户 ID, tokenid: Token ID
// 返回异步任务 ID (UPID)
func (c *Client) DeleteAPIToken(ctx context.Context, userid, tokenid string) (string, error) {
	var upid string
	path := fmt.Sprintf("access/users/%s/token/%s", userid, tokenid)
	if err := c.Delete(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("删除 API Token 失败: %w", err)
	}
	return upid, nil
}

// GetPermissions 获取权限树
// 返回当前用户的权限树（路径 -> 权限列表）
// 使用 interface{} 返回值防止 PVE 9.x 返回意外数据格式时 json.Unmarshal 失败
func (c *Client) GetPermissions(ctx context.Context) (interface{}, error) {
	var result interface{}
	if err := c.Get(ctx, "access", &result); err != nil {
		return nil, fmt.Errorf("获取权限树失败: %w", err)
	}
	return result, nil
}

// VerifyTicket 验证 ticket 有效性
// 返回当前 ticket 关联的用户信息
// 使用 interface{} 返回值防止 PVE 9.x 返回意外数据格式时 json.Unmarshal 失败
func (c *Client) VerifyTicket(ctx context.Context) (interface{}, error) {
	var result interface{}
	if err := c.Get(ctx, "access/ticket", &result); err != nil {
		return nil, fmt.Errorf("验证 ticket 失败: %w", err)
	}
	return result, nil
}

// GetAPIToken 获取 API Token 信息
// tokenid: 完整 Token ID (格式: USER@REALM!TOKENID)
func (c *Client) GetAPIToken(ctx context.Context, tokenid string) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("access/tokens/%s", tokenid)
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("获取 API Token 信息失败: %w", err)
	}
	return result, nil
}

// SearchUsers 搜索用户
// params: 搜索参数（enable, groups 等）
// 返回匹配的用户列表
// 使用 interface{} 返回值防止 PVE 9.x 返回意外数据格式时 json.Unmarshal 失败
func (c *Client) SearchUsers(ctx context.Context, params url.Values) (interface{}, error) {
	var result interface{}
	if err := c.GetWithParams(ctx, "access/users", params, &result); err != nil {
		return nil, fmt.Errorf("搜索用户失败: %w", err)
	}
	return result, nil
}
