/* global FormData */
import type { InternalAxiosRequestConfig } from 'axios'
import { apiClient } from './client'
import {
  TelegramSession,
  CreateSessionRequest,
  CreateSessionResponse,
  VerifyCodeRequest,
  SubmitPasswordRequest,
  ImportTDataRequest,
  ImportTDataResponse,
  RegenerateQRResponse,
  SessionStatus,
} from '@/types'

export const sessionsApi = {
  create: async (data: CreateSessionRequest): Promise<CreateSessionResponse> => {
    return apiClient.post<CreateSessionResponse>('/sessions', data)
  },

  verifyCode: async (sessionId: string, data: VerifyCodeRequest): Promise<CreateSessionResponse> => {
    return apiClient.post<CreateSessionResponse>(`/sessions/${sessionId}/verify`, data)
  },

  list: async (): Promise<TelegramSession[]> => {
    return apiClient.get<TelegramSession[]>('/sessions')
  },

  get: async (sessionId: string): Promise<SessionStatus> => {
    return apiClient.get<SessionStatus>(`/sessions/${sessionId}`)
  },

  delete: async (sessionId: string): Promise<{ deleted: boolean }> => {
    return apiClient.delete<{ deleted: boolean }>(`/sessions/${sessionId}`)
  },

  submitPassword: async (
    sessionId: string,
    data: SubmitPasswordRequest
  ): Promise<TelegramSession> => {
    return apiClient.post<TelegramSession>(
      `/sessions/${sessionId}/submit-password`,
      data
    )
  },

  regenerateQR: async (sessionId: string): Promise<RegenerateQRResponse> => {
    return apiClient.post<RegenerateQRResponse>(
      `/sessions/${sessionId}/qr/regenerate`
    )
  },

  importTData: async (
    data: ImportTDataRequest
  ): Promise<ImportTDataResponse> => {
    const formData = new FormData()
    formData.append('api_id', data.api_id.toString())
    formData.append('api_hash', data.api_hash)
    if (data.session_name) {
      formData.append('session_name', data.session_name)
    }
    data.tdata.forEach((file) => {
      formData.append('tdata', file)
    })

    return apiClient.postWithConfig<ImportTDataResponse>(
      '/sessions/import-tdata',
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      } as InternalAxiosRequestConfig
    )
  },
}
