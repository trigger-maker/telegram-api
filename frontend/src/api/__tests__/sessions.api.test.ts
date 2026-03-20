/**
 * Tests for sessions API functions
 */

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { sessionsApi } from '../sessions.api'
import { apiClient } from '../client'

// Mock the apiClient
vi.mock('../client', () => ({
  apiClient: {
    post: vi.fn(),
    get: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('sessionsApi.submitPassword', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should call apiClient.post with correct URL and data', async () => {
    const mockSession = {
      id: 'session-1',
      user_id: 'user-1',
      api_id: 12345,
      session_name: 'Test Session',
      auth_state: 'authenticated',
      is_active: true,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }

    vi.mocked(apiClient.post).mockResolvedValue(mockSession)

    const result = await sessionsApi.submitPassword('session-1', {
      password: 'test_password',
    })

    expect(apiClient.post).toHaveBeenCalledWith(
      '/sessions/session-1/submit-password',
      { password: 'test_password' }
    )
    expect(result).toEqual(mockSession)
  })

  it('should handle empty password', async () => {
    const mockError = new Error('Password is required')
    vi.mocked(apiClient.post).mockRejectedValue(mockError)

    await expect(
      sessionsApi.submitPassword('session-1', { password: '' })
    ).rejects.toThrow('Password is required')

    expect(apiClient.post).toHaveBeenCalledWith(
      '/sessions/session-1/submit-password',
      { password: '' }
    )
  })
})

describe('sessionsApi.regenerateQR', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should call apiClient.post with correct URL', async () => {
    const mockResponse = {
      session_id: 'session-1',
      qr_image_base64: 'new_base64_string',
      message: 'QR code regenerated successfully',
    }

    vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

    const result = await sessionsApi.regenerateQR('session-1')

    expect(apiClient.post).toHaveBeenCalledWith(
      '/sessions/session-1/qr/regenerate'
    )
    expect(result).toEqual(mockResponse)
  })

  it('should handle session not found', async () => {
    const mockError = new Error('Session not found')
    vi.mocked(apiClient.post).mockRejectedValue(mockError)

    await expect(sessionsApi.regenerateQR('invalid-session')).rejects.toThrow(
      'Session not found'
    )

    expect(apiClient.post).toHaveBeenCalledWith(
      '/sessions/invalid-session/qr/regenerate'
    )
  })
})

describe('sessionsApi.importTData', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should construct FormData with all fields', async () => {
    const mockResponse = {
      session: {
        session_id: 'session-1',
        is_active: true,
        telegram_user_id: 12345,
        username: 'testuser',
        auth_state: 'authenticated',
        auth_method: 'tdata' as const,
      },
    }

    vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

    const files: File[] = [new File(['content'], 'session.dat')]

    const result = await sessionsApi.importTData({
      api_id: 12345,
      api_hash: 'test_hash',
      session_name: 'Test Session',
      tdata: files,
    })

    expect(apiClient.post).toHaveBeenCalledWith(
      '/sessions/import-tdata',
      expect.any(FormData),
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    )

    // Verify FormData was constructed correctly
    const formDataCall = vi.mocked(apiClient.post).mock.calls[0]
    const formData = formDataCall[1] as FormData

    expect(formData.get('api_id')).toBe('12345')
    expect(formData.get('api_hash')).toBe('test_hash')
    expect(formData.get('session_name')).toBe('Test Session')

    expect(result).toEqual(mockResponse)
  })

  it('should construct FormData without optional session_name', async () => {
    const mockResponse = {
      session: {
        session_id: 'session-1',
        is_active: true,
        auth_state: 'authenticated',
        auth_method: 'tdata' as const,
      },
    }

    vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

    const files: File[] = [new File(['content'], 'session.dat')]

    await sessionsApi.importTData({
      api_id: 12345,
      api_hash: 'test_hash',
      tdata: files,
    })

    const formDataCall = vi.mocked(apiClient.post).mock.calls[0]
    const formData = formDataCall[1] as FormData

    expect(formData.get('session_name')).toBeNull()
  })

  it('should handle multiple TData files', async () => {
    const mockResponse = {
      session: {
        session_id: 'session-1',
        is_active: true,
        auth_state: 'authenticated',
        auth_method: 'tdata' as const,
      },
    }

    vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

    const files: File[] = [
      new File(['content1'], 'session.dat'),
      new File(['content2'], 'key.dat'),
      new File(['content3'], 'auth.dat'),
    ]

    await sessionsApi.importTData({
      api_id: 12345,
      api_hash: 'test_hash',
      tdata: files,
    })

    expect(apiClient.post).toHaveBeenCalledWith(
      '/sessions/import-tdata',
      expect.any(FormData),
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    )
  })

  it('should handle invalid API credentials', async () => {
    const mockError = new Error('Invalid API credentials')
    vi.mocked(apiClient.post).mockRejectedValue(mockError)

    const files: File[] = [new File(['content'], 'session.dat')]

    await expect(
      sessionsApi.importTData({
        api_id: 0,
        api_hash: 'invalid',
        tdata: files,
      })
    ).rejects.toThrow('Invalid API credentials')
  })
})

describe('sessionsApi existing functions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should create a session', async () => {
    const mockResponse = {
      session: {
        id: 'session-1',
        user_id: 'user-1',
        api_id: 12345,
        session_name: 'Test Session',
        auth_state: 'code_sent',
        is_active: false,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      phone_code_hash: 'test_hash',
    }

    vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

    const result = await sessionsApi.create({
      phone: '+1234567890',
      api_id: 12345,
      api_hash: 'test_hash',
      session_name: 'Test Session',
      auth_method: 'sms',
    })

    expect(apiClient.post).toHaveBeenCalledWith('/sessions', {
      phone: '+1234567890',
      api_id: 12345,
      api_hash: 'test_hash',
      session_name: 'Test Session',
      auth_method: 'sms',
    })
    expect(result).toEqual(mockResponse)
  })

  it('should verify code', async () => {
    const mockSession = {
      id: 'session-1',
      user_id: 'user-1',
      api_id: 12345,
      session_name: 'Test Session',
      auth_state: 'authenticated',
      is_active: true,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }

    vi.mocked(apiClient.post).mockResolvedValue(mockSession)

    const result = await sessionsApi.verifyCode('session-1', {
      code: '123456',
    })

    expect(apiClient.post).toHaveBeenCalledWith(
      '/sessions/session-1/verify',
      { code: '123456' }
    )
    expect(result).toEqual(mockSession)
  })

  it('should list sessions', async () => {
    const mockSessions = [
      {
        id: 'session-1',
        user_id: 'user-1',
        api_id: 12345,
        session_name: 'Test Session',
        auth_state: 'authenticated',
        is_active: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
    ]

    vi.mocked(apiClient.get).mockResolvedValue(mockSessions)

    const result = await sessionsApi.list()

    expect(apiClient.get).toHaveBeenCalledWith('/sessions')
    expect(result).toEqual(mockSessions)
  })

  it('should get session status', async () => {
    const mockStatus = {
      session: {
        id: 'session-1',
        user_id: 'user-1',
        api_id: 12345,
        session_name: 'Test Session',
        auth_state: 'authenticated',
        is_active: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      status: 'authenticated' as const,
    }

    vi.mocked(apiClient.get).mockResolvedValue(mockStatus)

    const result = await sessionsApi.get('session-1')

    expect(apiClient.get).toHaveBeenCalledWith('/sessions/session-1')
    expect(result).toEqual(mockStatus)
  })

  it('should delete session', async () => {
    const mockResponse = { deleted: true }
    vi.mocked(apiClient.delete).mockResolvedValue(mockResponse)

    const result = await sessionsApi.delete('session-1')

    expect(apiClient.delete).toHaveBeenCalledWith('/sessions/session-1')
    expect(result).toEqual(mockResponse)
  })
})
