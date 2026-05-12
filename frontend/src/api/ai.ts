import { get, post, put, del } from './request'

export interface AIModelConfig {
  id: number
  name: string
  provider: string
  base_url: string
  model: string
  max_tokens: number
  temperature: number
  timeout: number
  is_enabled: boolean
  is_default: boolean
  sort_order: number
  created_at: string
  updated_at: string
}

export interface AIConversation {
  id: number
  title: string
  scene: string
  model_config_id: number
  user_id: string
  created_at: string
  updated_at: string
  messages?: AIMessage[]
}

export interface AIMessage {
  id: number
  conversation_id: number
  role: 'user' | 'assistant' | 'system' | 'tool'
  content: string
  tool_calls?: string
  tool_call_id?: string
  created_at: string
}

export interface AIReport {
  id: number
  title: string
  type: string
  content: string
  model_config_id: number
  user_id: string
  schedule_id: number
  created_at: string
}

/**
 * AI 模型配置管理
 */
export async function getAIModels(): Promise<AIModelConfig[]> {
  return get<AIModelConfig[]>('/pve/ai/models')
}

export async function createAIModel(data: Partial<AIModelConfig>): Promise<AIModelConfig> {
  return post<AIModelConfig>('/pve/ai/models', data)
}

export async function updateAIModel(id: number, data: Partial<AIModelConfig>): Promise<void> {
  return put<void>(`/pve/ai/models/${id}`, data)
}

export async function deleteAIModel(id: number): Promise<void> {
  return del<void>(`/pve/ai/models/${id}`)
}

export async function setDefaultModel(id: number): Promise<void> {
  return post<void>(`/pve/ai/models/${id}/default`)
}

export async function testModelConnection(id: number): Promise<{ message: string }> {
  return post<{ message: string }>(`/pve/ai/models/${id}/test`)
}

/**
 * AI 对话管理
 */
export async function getConversations(limit = 50): Promise<AIConversation[]> {
  return get<AIConversation[]>('/pve/ai/conversations', { params: { limit } })
}

export async function getConversation(id: number): Promise<AIConversation> {
  return get<AIConversation>(`/pve/ai/conversations/${id}`)
}

export async function createConversation(data: {
  title?: string
  scene: string
  model_config_id?: number
  message: string
}): Promise<AIConversation> {
  return post<AIConversation>('/pve/ai/conversations', data)
}

/** 在已有对话中发送消息 */
export async function sendConversationMessage(
  conversationId: number,
  message: string
): Promise<AIConversation> {
  return post<AIConversation>(`/pve/ai/conversations/${conversationId}/message`, { message })
}

export async function deleteConversation(id: number): Promise<void> {
  return del<void>(`/pve/ai/conversations/${id}`)
}

/**
 * AI 报告管理
 */
export async function getReports(type?: string, limit = 50): Promise<AIReport[]> {
  return get<AIReport[]>('/pve/ai/reports', { params: { type, limit } })
}

export async function getReport(id: number): Promise<AIReport> {
  return get<AIReport>(`/pve/ai/reports/${id}`)
}

export async function generateReport(data: {
  title: string
  type: string
  model_config_id?: number
  content?: string
}): Promise<AIReport> {
  return post<AIReport>('/pve/ai/reports/generate', data)
}

export async function deleteReport(id: number): Promise<void> {
  return del<void>(`/pve/ai/reports/${id}`)
}

/**
 * AI 诊断和建议
 */
export async function diagnoseSystem(data: {
  node: string
  message?: string
}): Promise<any> {
  return post<any>('/pve/ai/diagnose', data)
}

export async function getSuggestion(data: {
  resource_type: string
  resource_id: string
  message?: string
}): Promise<any> {
  return post<any>('/pve/ai/suggest', data)
}
