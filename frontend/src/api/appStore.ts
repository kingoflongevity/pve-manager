import { get, post, del } from './request'

export interface AppTemplate {
  id: number
  name: string
  category: string
  description: string
  icon: string
  version: string
  author: string
  min_cpu: number
  min_memory_mb: number
  min_disk_gb: number
  type: string
  os_template: string
  packages: string
  variables: string
  setup_steps: string
  is_built_in: boolean
  created_at: string
  updated_at: string
}

export interface AppDeployment {
  id: number
  template_id: number
  name: string
  node: string
  type: string
  vmid: number
  status: string
  progress: number
  step_info: string
  config: string
  error_msg: string
  user_id: string
  started_at: string
  completed_at: string
  created_at: string
}

export function getAppTemplates(category?: string): Promise<AppTemplate[]> {
  return get<AppTemplate[]>('/pve/apps', category ? { category } : undefined)
}

export function getAppTemplateDetail(id: number): Promise<AppTemplate> {
  return get<AppTemplate>(`/pve/apps/${id}`)
}

export function getAppCategories(): Promise<string[]> {
  return get<string[]>('/pve/apps/categories')
}

export function deployApp(data: {
  template_id: number
  name: string
  node: string
  config?: Record<string, string>
}): Promise<{ deployment_id: number; message: string }> {
  return post<{ deployment_id: number; message: string }>('/pve/apps/deploy', data)
}

export function getAppDeployments(): Promise<AppDeployment[]> {
  return get<AppDeployment[]>('/pve/apps/deployments')
}

export function getAppDeploymentDetail(id: number): Promise<AppDeployment> {
  return get<AppDeployment>(`/pve/apps/deployments/${id}`)
}

export function deleteAppDeployment(id: number): Promise<void> {
  return del<void>(`/pve/apps/deployments/${id}`)
}

export function importAppTemplate(yaml: string): Promise<AppTemplate> {
  return post<AppTemplate>('/pve/apps/import', { yaml })
}

export function syncAppTemplates(remoteUrl: string): Promise<{ synced_count: number }> {
  return post<{ synced_count: number }>('/pve/apps/sync', { remote_url: remoteUrl })
}
