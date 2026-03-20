import { apiClient } from './client'

// =============== TYPES ===============

export interface WebhookConfig {
  id: string
  session_id: string
  url: string
  events: string[]
  secret?: string
  timeout_ms: number
  max_retries: number
  is_active: boolean
  last_error?: string
  last_error_at?: string
  created_at: string
  updated_at: string
}

export interface WebhookCreateRequest {
  url: string
  events?: string[]
  secret?: string
  timeout_ms?: number
  max_retries?: number
}

export interface WebhookResponse {
  id: string
  session_id: string
  url: string
  events: string[]
  is_active: boolean
}

// =============== API FUNCTIONS ===============

/**
 * Gets the current webhook configuration
 */
export const getWebhookConfig = async (sessionId: string): Promise<WebhookConfig> => {
  return apiClient.get<WebhookConfig>(`/sessions/${sessionId}/webhook`)
}

/**
 * Configures a webhook for the session
 */
export const createWebhook = async (
  sessionId: string,
  data: WebhookCreateRequest
): Promise<WebhookResponse> => {
  return apiClient.post<WebhookResponse>(`/sessions/${sessionId}/webhook`, data)
}

/**
 * Deletes the webhook configuration
 */
export const deleteWebhook = async (sessionId: string): Promise<void> => {
  return apiClient.delete(`/sessions/${sessionId}/webhook`)
}

/**
 * Starts listening to webhook events
 */
export const startWebhook = async (sessionId: string): Promise<void> => {
  return apiClient.post(`/sessions/${sessionId}/webhook/start`)
}

/**
 * Stops listening to webhook events
 */
export const stopWebhook = async (sessionId: string): Promise<void> => {
  return apiClient.post(`/sessions/${sessionId}/webhook/stop`)
}

// =============== POOL STATUS ===============

export interface PoolSessionInfo {
  session_id: string
  session_name: string
  telegram_id: number
  started_at: string
  is_connected: boolean
}

export interface PoolStatus {
  active_count: number
  sessions: PoolSessionInfo[]
}

/**
 * Gets the state of the active sessions pool
 */
export const getPoolStatus = async (): Promise<PoolStatus> => {
  return apiClient.get<PoolStatus>('/pool/status')
}
