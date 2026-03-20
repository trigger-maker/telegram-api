export interface TelegramSession {
  id: string
  user_id: string
  phone_number?: string
  api_id: number
  session_name: string
  auth_state: string
  telegram_user_id?: number
  telegram_username?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export type AuthMethod = 'sms' | 'qr' | 'tdata'

export interface CreateSessionRequest {
  phone?: string
  api_id: number
  api_hash: string
  session_name: string
  auth_method?: AuthMethod
}

export interface CreateSessionResponse {
  session: TelegramSession
  phone_code_hash?: string
  qr_image_base64?: string
  hint?: string // 2FA password hint
  message?: string
  next_step?: string
}

export interface SubmitPasswordRequest {
  password: string
}

export interface ImportTDataRequest {
  api_id: number
  api_hash: string
  session_name?: string
  tdata: File[]
}

export interface ImportTDataResponse {
  session: {
    session_id: string
    is_active: boolean
    telegram_user_id?: number
    username?: string
    auth_state: string
    auth_method: 'tdata'
  }
}

export interface RegenerateQRResponse {
  session_id: string
  qr_image_base64: string
  message: string
}

export interface VerifyCodeRequest {
  code: string
}

export interface SessionStatus {
  session: TelegramSession
  status: 'waiting' | 'failed' | 'authenticated'
  message?: string
}
