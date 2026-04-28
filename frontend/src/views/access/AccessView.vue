<template>
  <div class="access-view">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2>访问管理</h2>
        <span class="header-subtitle">用户、组、角色、权限及认证域管理</span>
      </div>
      <div class="header-right">
        <el-button :icon="Refresh" :loading="loading" @click="refreshAll">刷新</el-button>
      </div>
    </div>

    <!-- Tab 切换 -->
    <el-tabs v-model="activeTab" class="access-tabs">
      <!-- Tab 1: 用户管理 -->
      <el-tab-pane label="用户管理" name="users">
        <div class="tab-toolbar">
          <el-input
            v-model="userSearch"
            placeholder="搜索用户名"
            clearable
            style="width: 200px"
          />
          <el-select v-model="userRealmFilter" placeholder="认证域" clearable style="width: 120px">
            <el-option label="全部" value="" />
            <el-option label="PVE" value="pve" />
            <el-option label="PAM" value="pam" />
            <el-option label="LDAP" value="ldap" />
          </el-select>
          <el-button type="primary" :icon="Plus" @click="handleCreateUser">创建用户</el-button>
        </div>

        <el-table :data="filteredUsers" stripe border v-loading="loading">
          <el-table-column prop="username" label="用户名" width="160">
            <template #default="{ row }">{{ row.username }}@{{ row.realm }}</template>
          </el-table-column>
          <el-table-column label="全名" width="150">
            <template #default="{ row }">
              {{ [row.firstname, row.lastname].filter(Boolean).join(' ') || '--' }}
            </template>
          </el-table-column>
          <el-table-column prop="email" label="邮箱" width="200" />
          <el-table-column label="启用" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="row.enabled ? 'success' : 'info'" size="small">
                {{ row.enabled ? '是' : '否' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="realm" label="认证域" width="100" />
          <el-table-column label="操作" width="280" align="center" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="handleEditUser(row)">编辑</el-button>
              <el-button link type="warning" size="small" @click="handleResetPassword(row)">重置密码</el-button>
              <el-button link :type="row.enabled ? 'warning' : 'success'" size="small" @click="handleToggleUser(row)">
                {{ row.enabled ? '禁用' : '启用' }}
              </el-button>
              <el-button link type="danger" size="small" @click="handleDeleteUser(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Tab 2: 组管理 -->
      <el-tab-pane label="组管理" name="groups">
        <div class="tab-toolbar">
          <el-input v-model="groupSearch" placeholder="搜索组名称" clearable style="width: 200px" />
          <el-button type="primary" :icon="Plus" @click="handleCreateGroup">创建组</el-button>
        </div>

        <el-table :data="filteredGroups" stripe border v-loading="loading">
          <el-table-column prop="groupid" label="组名称" width="200" />
          <el-table-column prop="comment" label="注释" width="250" />
          <el-table-column label="成员" min-width="200">
            <template #default="{ row }">
              <el-tag
                v-for="user in (row.users || [])"
                :key="user"
                size="small"
                style="margin-right: 4px; margin-bottom: 4px"
              >
                {{ user }}
              </el-tag>
              <span v-if="!row.users || row.users.length === 0" style="color: #909399">暂无成员</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="180" align="center" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="handleEditGroup(row)">编辑</el-button>
              <el-button link type="danger" size="small" @click="handleDeleteGroup(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Tab 3: 角色管理 -->
      <el-tab-pane label="角色管理" name="roles">
        <div class="tab-toolbar">
          <el-input v-model="roleSearch" placeholder="搜索角色名称" clearable style="width: 200px" />
          <el-button type="primary" :icon="Plus" @click="handleCreateRole">创建角色</el-button>
        </div>

        <el-table :data="filteredRoles" stripe border v-loading="loading">
          <el-table-column prop="roleid" label="角色名称" width="200" />
          <el-table-column label="权限" min-width="400">
            <template #default="{ row }">
              <el-tag
                v-for="priv in row.privs.split(',')"
                :key="priv"
                size="small"
                type="info"
                style="margin-right: 4px; margin-bottom: 4px"
              >
                {{ priv.trim() }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="180" align="center" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="handleEditRole(row)">编辑</el-button>
              <el-button link type="danger" size="small" @click="handleDeleteRole(row)" :disabled="isSystemRole(row.roleid)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Tab 4: 权限管理 (ACL) -->
      <el-tab-pane label="权限管理" name="acls">
        <div class="tab-toolbar">
          <el-input v-model="aclSearch" placeholder="搜索路径" clearable style="width: 200px" />
          <el-button type="primary" :icon="Plus" @click="handleAddACL">添加权限</el-button>
        </div>

        <el-table :data="filteredACLs" stripe border v-loading="loading">
          <el-table-column prop="path" label="路径" width="250" />
          <el-table-column label="类型" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="row.type === 'user' ? '' : 'warning'" size="small">
                {{ row.type === 'user' ? '用户' : '组' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="auth_id" label="用户/组" width="200" />
          <el-table-column prop="roleid" label="角色" width="150" />
          <el-table-column label="继承" width="80" align="center">
            <template #default="{ row }">
              <el-icon v-if="row.propagate"><Check /></el-icon>
              <span v-else style="color: #dcdfe6">-</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120" align="center" fixed="right">
            <template #default="{ row, $index }">
              <el-button link type="danger" size="small" @click="handleDeleteACL(row, $index)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Tab 5: 认证域 -->
      <el-tab-pane label="认证域" name="domains">
        <el-table :data="domains" stripe border v-loading="loading">
          <el-table-column prop="realm" label="认证域" width="150" />
          <el-table-column label="类型" width="120">
            <template #default="{ row }">
              <el-tag :type="getRealmType(row.type)" size="small">
                {{ getRealmLabel(row.type) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="comment" label="注释" min-width="200" />
          <el-table-column label="默认" width="80" align="center">
            <template #default="{ row }">
              <el-tag v-if="row.default" type="success" size="small">是</el-tag>
              <span v-else style="color: #dcdfe6">-</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120" align="center" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="handleViewDomain(row)">查看配置</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- 用户创建/编辑对话框 -->
    <el-dialog v-model="userDialogVisible" :title="editingUser ? '编辑用户' : '创建用户'" width="520px">
      <el-form ref="userFormRef" :model="userForm" :rules="userRules" label-width="100px">
        <el-form-item v-if="!editingUser" label="用户名" prop="username">
          <el-input v-model="userForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item v-if="!editingUser" label="认证域" prop="realm">
          <el-select v-model="userForm.realm" style="width: 100%">
            <el-option label="PVE 认证" value="pve" />
            <el-option label="PAM 认证" value="pam" />
            <el-option label="LDAP 认证" value="ldap" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="!editingUser" label="密码" prop="password">
          <el-input v-model="userForm.password" type="password" show-password placeholder="请输入密码" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="userForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="名" prop="firstname">
          <el-input v-model="userForm.firstname" placeholder="名" />
        </el-form-item>
        <el-form-item label="姓" prop="lastname">
          <el-input v-model="userForm.lastname" placeholder="姓" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="userForm.enabled" />
        </el-form-item>
        <el-form-item label="所属组">
          <el-select v-model="userForm.groups" multiple filterable style="width: 100%">
            <el-option v-for="g in groups" :key="g.groupid" :label="g.groupid" :value="g.groupid" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="userDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleUserSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 组创建/编辑对话框 -->
    <el-dialog v-model="groupDialogVisible" :title="editingGroup ? '编辑组' : '创建组'" width="480px">
      <el-form ref="groupFormRef" :model="groupForm" :rules="groupRules" label-width="100px">
        <el-form-item v-if="!editingGroup" label="组名称" prop="groupid">
          <el-input v-model="groupForm.groupid" placeholder="请输入组名称" />
        </el-form-item>
        <el-form-item label="注释">
          <el-input v-model="groupForm.comment" type="textarea" :rows="2" placeholder="注释" />
        </el-form-item>
        <el-form-item label="成员">
          <el-select v-model="groupForm.users" multiple filterable style="width: 100%">
            <el-option v-for="u in users" :key="u.userid" :label="u.userid" :value="u.userid" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="groupDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleGroupSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 角色创建/编辑对话框 -->
    <el-dialog v-model="roleDialogVisible" :title="editingRole ? '编辑角色' : '创建角色'" width="520px">
      <el-form ref="roleFormRef" :model="roleForm" :rules="roleRules" label-width="100px">
        <el-form-item v-if="!editingRole" label="角色名称" prop="roleid">
          <el-input v-model="roleForm.roleid" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="权限">
          <div class="privilege-grid">
            <el-checkbox-group v-model="roleForm.privs">
              <el-checkbox v-for="priv in allPrivileges" :key="priv" :value="priv" style="width: 50%">
                {{ priv }}
              </el-checkbox>
            </el-checkbox-group>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleRoleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- ACL 添加对话框 -->
    <el-dialog v-model="aclDialogVisible" title="添加权限" width="520px">
      <el-form ref="aclFormRef" :model="aclForm" :rules="aclRules" label-width="100px">
        <el-form-item label="路径" prop="path">
          <el-select v-model="aclForm.path" placeholder="选择路径" style="width: 100%" filterable allow-create>
            <el-option label="/" value="/" />
            <el-option v-for="node in nodes" :key="node.name" :label="`/nodes/${node.name}`" :value="`/nodes/${node.name}`" />
            <el-option v-for="vm in vms" :key="vm.vmid" :label="`/vms/${vm.vmid}`" :value="`/vms/${vm.vmid}`" />
            <el-option v-for="s in storages" :key="s.storage" :label="`/storage/${s.storage}`" :value="`/storage/${s.storage}`" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-radio-group v-model="aclForm.type">
            <el-radio value="user">用户</el-radio>
            <el-radio value="group">组</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="用户/组" prop="auth_id">
          <el-select v-model="aclForm.auth_id" placeholder="选择用户或组" style="width: 100%" filterable>
            <el-option v-if="aclForm.type === 'user'" v-for="u in users" :key="u.userid" :label="u.userid" :value="u.userid" />
            <el-option v-if="aclForm.type === 'group'" v-for="g in groups" :key="g.groupid" :label="g.groupid" :value="g.groupid" />
          </el-select>
        </el-form-item>
        <el-form-item label="角色" prop="roleid">
          <el-select v-model="aclForm.roleid" placeholder="选择角色" style="width: 100%">
            <el-option v-for="r in roles" :key="r.roleid" :label="r.roleid" :value="r.roleid" />
          </el-select>
        </el-form-item>
        <el-form-item label="向下继承">
          <el-switch v-model="aclForm.propagate" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="aclDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleACLSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 认证域配置对话框 -->
    <el-dialog v-model="domainDialogVisible" title="认证域配置" width="520px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="认证域">{{ selectedDomain?.realm }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ getRealmLabel(selectedDomain?.type || '') }}</el-descriptions-item>
        <el-descriptions-item label="注释">{{ selectedDomain?.comment || '--' }}</el-descriptions-item>
        <el-descriptions-item label="默认">{{ selectedDomain?.default ? '是' : '否' }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="domainDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
/**
 * AccessView - 访问管理页面
 * 
 * 管理用户、组、角色、权限（ACL）及认证域，
 * 所有操作通过真实 API 端点进行。
 */
import { ref, computed, onMounted } from 'vue'
import { Refresh, Plus, Check } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  listUsers, createUser, updateUser, deleteUser, updateUserPassword,
  listGroups, createGroup, updateGroup, deleteGroup,
  listRoles, createRole, updateRole, deleteRole,
  listACLs, setACL,
  listDomains, getDomain,
} from '@/api/access'
import { getClusterResources } from '@/api/cluster'
import type { User, Group, Role, ACL } from '@/api/types'

// ============================================================
// 通用状态
// ============================================================

const loading = ref(false)
const submitting = ref(false)
const activeTab = ref('users')

// ============================================================
// 用户管理
// ============================================================

const users = ref<User[]>([])
const userSearch = ref('')
const userRealmFilter = ref('')
const userDialogVisible = ref(false)
const editingUser = ref<User | null>(null)
const userFormRef = ref<FormInstance>()

const userForm = ref({
  username: '',
  realm: 'pve',
  password: '',
  email: '',
  firstname: '',
  lastname: '',
  enabled: true,
  groups: [] as string[],
})

const userRules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  realm: [{ required: true, message: '请选择认证域', trigger: 'change' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

const filteredUsers = computed(() => {
  let result = users.value
  if (userRealmFilter.value) {
    result = result.filter(u => u.realm.toLowerCase() === userRealmFilter.value.toLowerCase())
  }
  if (userSearch.value) {
    const q = userSearch.value.toLowerCase()
    result = result.filter(u => u.username.toLowerCase().includes(q) || u.userid.toLowerCase().includes(q))
  }
  return result
})

async function loadUsers() {
  try {
    users.value = await listUsers()
  } catch (error) {
    console.error('获取用户列表失败:', error)
  }
}

function handleCreateUser() {
  editingUser.value = null
  userForm.value = { username: '', realm: 'pve', password: '', email: '', firstname: '', lastname: '', enabled: true, groups: [] }
  userDialogVisible.value = true
}

function handleEditUser(user: User) {
  editingUser.value = user
  userForm.value = {
    username: user.username,
    realm: user.realm,
    password: '',
    email: user.email || '',
    firstname: user.firstname || '',
    lastname: user.lastname || '',
    enabled: !!user.enabled,
    groups: [],
  }
  userDialogVisible.value = true
}

async function handleUserSubmit() {
  const valid = await userFormRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    if (editingUser.value) {
      await updateUser(editingUser.value.userid, {
        email: userForm.value.email,
        firstname: userForm.value.firstname,
        lastname: userForm.value.lastname,
        enabled: userForm.value.enabled ? 1 : 0,
      })
      ElMessage.success('用户已更新')
    } else {
      await createUser({
        userid: `${userForm.value.username}@${userForm.value.realm}`,
        password: userForm.value.password,
        email: userForm.value.email,
        firstname: userForm.value.firstname,
        lastname: userForm.value.lastname,
        enabled: userForm.value.enabled ? 1 : 0,
        groups: userForm.value.groups,
      })
      ElMessage.success('用户已创建')
    }
    userDialogVisible.value = false
    await loadUsers()
  } catch (error) {
    console.error('操作失败:', error)
  } finally {
    submitting.value = false
  }
}

async function handleResetPassword(user: User) {
  const { value: password } = await ElMessageBox.prompt('请输入新密码', '重置密码', {
    inputType: 'password',
    inputPlaceholder: '请输入新密码',
  }).catch(() => ({ value: '' }))
  if (!password) return
  submitting.value = true
  try {
    await updateUserPassword(user.userid, password)
    ElMessage.success('密码已重置')
  } catch (error) {
    console.error('重置密码失败:', error)
  } finally {
    submitting.value = false
  }
}

async function handleToggleUser(user: User) {
  submitting.value = true
  try {
    await updateUser(user.userid, { enabled: user.enabled ? 0 : 1 })
    ElMessage.success(user.enabled ? '用户已禁用' : '用户已启用')
    await loadUsers()
  } catch (error) {
    console.error('操作失败:', error)
  } finally {
    submitting.value = false
  }
}

async function handleDeleteUser(user: User) {
  try {
    await ElMessageBox.confirm(`确定要删除用户"${user.userid}"吗？`, '确认删除', { type: 'warning' })
    await deleteUser(user.userid)
    ElMessage.success('用户已删除')
    await loadUsers()
  } catch {
    // 取消
  }
}

// ============================================================
// 组管理
// ============================================================

const groups = ref<Group[]>([])
const groupSearch = ref('')
const groupDialogVisible = ref(false)
const editingGroup = ref<Group | null>(null)
const groupFormRef = ref<FormInstance>()

const groupForm = ref({
  groupid: '',
  comment: '',
  users: [] as string[],
})

const groupRules: FormRules = {
  groupid: [{ required: true, message: '请输入组名称', trigger: 'blur' }],
}

const filteredGroups = computed(() => {
  if (!groupSearch.value) return groups.value
  const q = groupSearch.value.toLowerCase()
  return groups.value.filter(g => g.groupid.toLowerCase().includes(q))
})

async function loadGroups() {
  try {
    groups.value = await listGroups()
  } catch (error) {
    console.error('获取组列表失败:', error)
  }
}

function handleCreateGroup() {
  editingGroup.value = null
  groupForm.value = { groupid: '', comment: '', users: [] }
  groupDialogVisible.value = true
}

function handleEditGroup(group: Group) {
  editingGroup.value = group
  groupForm.value = {
    groupid: group.groupid,
    comment: group.comment || '',
    users: [...(group.users || [])],
  }
  groupDialogVisible.value = true
}

async function handleGroupSubmit() {
  const valid = await groupFormRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    if (editingGroup.value) {
      await updateGroup(editingGroup.value.groupid, {
        comment: groupForm.value.comment,
        users: groupForm.value.users.join(','),
      })
      ElMessage.success('组已更新')
    } else {
      await createGroup({
        groupid: groupForm.value.groupid,
        comment: groupForm.value.comment,
        users: groupForm.value.users,
      })
      ElMessage.success('组已创建')
    }
    groupDialogVisible.value = false
    await loadGroups()
  } catch (error) {
    console.error('操作失败:', error)
  } finally {
    submitting.value = false
  }
}

async function handleDeleteGroup(group: Group) {
  try {
    await ElMessageBox.confirm(`确定要删除组"${group.groupid}"吗？`, '确认删除', { type: 'warning' })
    await deleteGroup(group.groupid)
    ElMessage.success('组已删除')
    await loadGroups()
  } catch {
    // 取消
  }
}

// ============================================================
// 角色管理
// ============================================================

const roles = ref<Role[]>([])
const roleSearch = ref('')
const roleDialogVisible = ref(false)
const editingRole = ref<Role | null>(null)
const roleFormRef = ref<FormInstance>()

const roleForm = ref({
  roleid: '',
  privs: [] as string[],
})

const roleRules: FormRules = {
  roleid: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
}

const allPrivileges = [
  'VM.Audit', 'VM.PowerMgmt', 'VM.Config.CPU', 'VM.Config.Memory',
  'VM.Config.Disk', 'VM.Config.Network', 'VM.Config.CDROM',
  'VM.Config.HWType', 'VM.Config.Options', 'VM.Backup',
  'VM.Console', 'VM.Monitor', 'VM.Snapshot',
  'Datastore.Allocate', 'Datastore.AllocateSpace', 'Datastore.AllocateTemplate',
  'Datastore.Audit',
  'Pool.Allocate', 'Pool.Audit',
  'Sys.Audit', 'Sys.Modify', 'Sys.Console', 'Sys.Syslog',
  'Sys.PowerMgmt', 'Realm.Allocate', 'Realm.Audit',
]

const systemRoles = ['Administrator', 'PVEAdmin', 'PVEAuditor', 'PVEDatastoreAdmin', 'PVESysAdmin', 'PVEPoolAdmin', 'PVEVMAdmin', 'PVEVMUser', 'PVETemplateUser', 'PVEScheduler', 'NoAccess']

const filteredRoles = computed(() => {
  if (!roleSearch.value) return roles.value
  const q = roleSearch.value.toLowerCase()
  return roles.value.filter(r => r.roleid.toLowerCase().includes(q))
})

function isSystemRole(roleId: string): boolean {
  return systemRoles.includes(roleId)
}

async function loadRoles() {
  try {
    roles.value = await listRoles()
  } catch (error) {
    console.error('获取角色列表失败:', error)
  }
}

function handleCreateRole() {
  editingRole.value = null
  roleForm.value = { roleid: '', privs: [] }
  roleDialogVisible.value = true
}

function handleEditRole(role: Role) {
  editingRole.value = role
  roleForm.value = {
    roleid: role.roleid,
    privs: role.privs.split(',').map(p => p.trim()),
  }
  roleDialogVisible.value = true
}

async function handleRoleSubmit() {
  const valid = await roleFormRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    if (editingRole.value) {
      await updateRole(editingRole.value.roleid, roleForm.value.privs.join(','))
      ElMessage.success('角色已更新')
    } else {
      await createRole({
        roleid: roleForm.value.roleid,
        privs: roleForm.value.privs.join(','),
      })
      ElMessage.success('角色已创建')
    }
    roleDialogVisible.value = false
    await loadRoles()
  } catch (error) {
    console.error('操作失败:', error)
  } finally {
    submitting.value = false
  }
}

async function handleDeleteRole(role: Role) {
  if (isSystemRole(role.roleid)) {
    ElMessage.warning('系统内置角色不能删除')
    return
  }
  try {
    await ElMessageBox.confirm(`确定要删除角色"${role.roleid}"吗？`, '确认删除', { type: 'warning' })
    await deleteRole(role.roleid)
    ElMessage.success('角色已删除')
    await loadRoles()
  } catch {
    // 取消
  }
}

// ============================================================
// ACL 权限管理
// ============================================================

const acls = ref<ACL[]>([])
const aclSearch = ref('')
const aclDialogVisible = ref(false)
const aclFormRef = ref<FormInstance>()
const aclForm = ref({
  path: '/',
  type: 'user' as 'user' | 'group',
  auth_id: '',
  roleid: '',
  propagate: true,
})

const aclRules: FormRules = {
  path: [{ required: true, message: '请选择或输入路径', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  auth_id: [{ required: true, message: '请选择用户或组', trigger: 'change' }],
  roleid: [{ required: true, message: '请选择角色', trigger: 'change' }],
}

const filteredACLs = computed(() => {
  if (!aclSearch.value) return acls.value
  const q = aclSearch.value.toLowerCase()
  return acls.value.filter(a => a.path.toLowerCase().includes(q))
})

async function loadACLs() {
  try {
    acls.value = await listACLs()
  } catch (error) {
    console.error('获取ACL列表失败:', error)
  }
}

function handleAddACL() {
  aclForm.value = { path: '/', type: 'user', auth_id: '', roleid: '', propagate: true }
  aclDialogVisible.value = true
}

async function handleACLSubmit() {
  const valid = await aclFormRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    await setACL({
      path: aclForm.value.path,
      roles: aclForm.value.roleid,
      auth_id: aclForm.value.auth_id,
      users: aclForm.value.type === 'user' ? aclForm.value.auth_id : undefined,
      groups: aclForm.value.type === 'group' ? aclForm.value.auth_id : undefined,
      propagate: aclForm.value.propagate ? 1 : 0,
    })
    ElMessage.success('权限已添加')
    aclDialogVisible.value = false
    await loadACLs()
  } catch (error) {
    console.error('操作失败:', error)
  } finally {
    submitting.value = false
  }
}

async function handleDeleteACL(acl: ACL, index: number) {
  try {
    await ElMessageBox.confirm('确定要删除此权限吗？', '确认删除', { type: 'warning' })
    await setACL({
      path: acl.path,
      roles: acl.roleid,
      auth_id: acl.auth_id,
      users: acl.type === 'user' ? acl.auth_id : undefined,
      groups: acl.type === 'group' ? acl.auth_id : undefined,
      delete: 1,
    })
    ElMessage.success('权限已删除')
    await loadACLs()
  } catch {
    // 取消
  }
}

// ============================================================
// 认证域管理
// ============================================================

interface DomainInfo {
  realm: string
  type: string
  comment?: string
  default: number
}

const domains = ref<DomainInfo[]>([])
const domainDialogVisible = ref(false)
const selectedDomain = ref<DomainInfo | null>(null)

async function loadDomains() {
  try {
    domains.value = await listDomains()
  } catch (error) {
    console.error('获取认证域列表失败:', error)
  }
}

function handleViewDomain(domain: DomainInfo) {
  selectedDomain.value = domain
  domainDialogVisible.value = true
}

function getRealmType(type: string): '' | 'success' | 'warning' | 'info' {
  switch (type.toUpperCase()) {
    case 'PVE': return 'success'
    case 'PAM': return 'warning'
    case 'LDAP':
    case 'AD': return 'info'
    default: return ''
  }
}

function getRealmLabel(type: string): string {
  switch (type.toUpperCase()) {
    case 'PVE': return 'PVE 认证'
    case 'PAM': return 'PAM 认证'
    case 'LDAP': return 'LDAP'
    case 'AD': return 'Active Directory'
    default: return type
  }
}

// ============================================================
// 集群资源（用于 ACL 路径选择）
// ============================================================

const nodes = ref<{ name: string }[]>([])
const vms = ref<{ vmid: number; name: string }[]>([])
const storages = ref<{ storage: string }[]>([])

async function loadClusterResources() {
  try {
    const resources = await getClusterResources()
    nodes.value = resources.filter(r => r.type === 'node').map(r => ({ name: r.node || r.id }))
    vms.value = resources.filter(r => r.type === 'vm').map(r => ({ vmid: r.vmid || 0, name: r.name || '' }))
    storages.value = resources.filter(r => r.type === 'storage').map(r => ({ storage: r.storage || r.id }))
  } catch (error) {
    console.error('获取集群资源失败:', error)
  }
}

// ============================================================
// 加载与刷新
// ============================================================

async function loadData() {
  loading.value = true
  try {
    await Promise.all([
      loadUsers(),
      loadGroups(),
      loadRoles(),
      loadACLs(),
      loadDomains(),
      loadClusterResources(),
    ])
  } finally {
    loading.value = false
  }
}

async function refreshAll() {
  await loadData()
  ElMessage.success('刷新成功')
}

onMounted(() => {
  loadData()
})
</script>

<style scoped lang="scss">
@use '@/assets/styles/variables' as *;

.access-view {
  padding: $spacing-6;
  background: $color-bg-base;
  min-height: 100%;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: $spacing-4;

  .header-left {
    h2 {
      margin: 0;
      font-size: $font-size-2xl;
      font-weight: $font-weight-semibold;
      color: $color-text-primary;
    }

    .header-subtitle {
      font-size: $font-size-sm;
      color: $color-text-secondary;
      margin-top: $spacing-1;
    }
  }
}

.access-tabs {
  :deep(.el-tabs__header) {
    margin-bottom: $spacing-4;
  }
}

.tab-toolbar {
  display: flex;
  gap: $spacing-3;
  align-items: center;
  margin-bottom: $spacing-4;
}

.privilege-grid {
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid $color-border-light;
  border-radius: $radius-sm;
  padding: $spacing-3;

  .el-checkbox {
    margin: $spacing-1 0;
  }
}
</style>
