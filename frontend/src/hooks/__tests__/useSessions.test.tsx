/**
 * Tests for sessions hooks
 */

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  useSubmitPassword,
  useRegenerateQR,
  useImportTData,
  SESSIONS_QUERY_KEY,
} from '../useSessions'
import { sessionsApi } from '@/api/sessions.api'

// Mock the sessionsApi
vi.mock('@/api/sessions.api', () => ({
  sessionsApi: {
    submitPassword: vi.fn(),
    regenerateQR: vi.fn(),
    importTData: vi.fn(),
  },
}))

describe('useSubmitPassword', () => {
  let queryClient: QueryClient

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        mutations: {
          retry: false,
        },
      },
    })
    vi.clearAllMocks()
  })

  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )

  it('should submit password successfully', async () => {
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

    vi.mocked(sessionsApi.submitPassword).mockResolvedValue(mockSession)

    const { result } = renderHook(() => useSubmitPassword(), { wrapper })

    await result.current.mutateAsync({
      sessionId: 'session-1',
      password: 'test_password',
    })

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })

    expect(sessionsApi.submitPassword).toHaveBeenCalledWith('session-1', {
      password: 'test_password',
    })
  })

  it('should invalidate sessions query on success', async () => {
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

    vi.mocked(sessionsApi.submitPassword).mockResolvedValue(mockSession)

    const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries')

    const { result } = renderHook(() => useSubmitPassword(), { wrapper })

    await result.current.mutateAsync({
      sessionId: 'session-1',
      password: 'test_password',
    })

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })

    expect(invalidateQueriesSpy).toHaveBeenCalledWith({
      queryKey: [SESSIONS_QUERY_KEY],
    })
  })

  it('should handle submission error', async () => {
    const mockError = new Error('Invalid password')
    vi.mocked(sessionsApi.submitPassword).mockRejectedValue(mockError)

    const { result } = renderHook(() => useSubmitPassword(), { wrapper })

    await expect(
      result.current.mutateAsync({
        sessionId: 'session-1',
        password: 'wrong_password',
      })
    ).rejects.toThrow('Invalid password')

    await waitFor(() => {
      expect(result.current.isError).toBe(true)
    })
  })

  it('should handle empty password', async () => {
    const mockError = new Error('Password is required')
    vi.mocked(sessionsApi.submitPassword).mockRejectedValue(mockError)

    const { result } = renderHook(() => useSubmitPassword(), { wrapper })

    await expect(
      result.current.mutateAsync({
        sessionId: 'session-1',
        password: '',
      })
    ).rejects.toThrow('Password is required')
  })
})

describe('useRegenerateQR', () => {
  let queryClient: QueryClient

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        mutations: {
          retry: false,
        },
      },
    })
    vi.clearAllMocks()
  })

  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )

  it('should regenerate QR code successfully', async () => {
    const mockResponse = {
      session_id: 'session-1',
      qr_image_base64: 'new_base64_string',
      message: 'QR code regenerated successfully',
    }

    vi.mocked(sessionsApi.regenerateQR).mockResolvedValue(mockResponse)

    const { result } = renderHook(() => useRegenerateQR(), { wrapper })

    await result.current.mutateAsync('session-1')

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })

    expect(sessionsApi.regenerateQR).toHaveBeenCalledWith('session-1')
  })

  it('should handle regeneration error', async () => {
    const mockError = new Error('Session not found')
    vi.mocked(sessionsApi.regenerateQR).mockRejectedValue(mockError)

    const { result } = renderHook(() => useRegenerateQR(), { wrapper })

    await expect(
      result.current.mutateAsync('invalid-session')
    ).rejects.toThrow('Session not found')

    await waitFor(() => {
      expect(result.current.isError).toBe(true)
    })
  })

  it('should handle max attempts exceeded error', async () => {
    const mockError = new Error('Maximum regeneration attempts exceeded')
    vi.mocked(sessionsApi.regenerateQR).mockRejectedValue(mockError)

    const { result } = renderHook(() => useRegenerateQR(), { wrapper })

    await expect(
      result.current.mutateAsync('session-1')
    ).rejects.toThrow('Maximum regeneration attempts exceeded')

    await waitFor(() => {
      expect(result.current.isError).toBe(true)
    })
  })
})

describe('useImportTData', () => {
  let queryClient: QueryClient

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        mutations: {
          retry: false,
        },
      },
    })
    vi.clearAllMocks()
  })

  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )

  it('should import TData successfully', async () => {
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

    vi.mocked(sessionsApi.importTData).mockResolvedValue(mockResponse)

    const { result } = renderHook(() => useImportTData(), { wrapper })

    const files: File[] = [new File(['content'], 'session.dat')]

    await result.current.mutateAsync({
      api_id: 12345,
      api_hash: 'test_hash',
      session_name: 'Test Session',
      tdata: files,
    })

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })

    expect(sessionsApi.importTData).toHaveBeenCalledWith({
      api_id: 12345,
      api_hash: 'test_hash',
      session_name: 'Test Session',
      tdata: files,
    })
  })

  it('should invalidate sessions query on success', async () => {
    const mockResponse = {
      session: {
        session_id: 'session-1',
        is_active: true,
        auth_state: 'authenticated',
        auth_method: 'tdata' as const,
      },
    }

    vi.mocked(sessionsApi.importTData).mockResolvedValue(mockResponse)

    const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries')

    const { result } = renderHook(() => useImportTData(), { wrapper })

    const files: File[] = [new File(['content'], 'session.dat')]

    await result.current.mutateAsync({
      api_id: 12345,
      api_hash: 'test_hash',
      tdata: files,
    })

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })

    expect(invalidateQueriesSpy).toHaveBeenCalledWith({
      queryKey: [SESSIONS_QUERY_KEY],
    })
  })

  it('should handle import error', async () => {
    const mockError = new Error('Invalid TData files')
    vi.mocked(sessionsApi.importTData).mockRejectedValue(mockError)

    const { result } = renderHook(() => useImportTData(), { wrapper })

    const files: File[] = [new File(['content'], 'invalid.dat')]

    await expect(
      result.current.mutateAsync({
        api_id: 12345,
        api_hash: 'test_hash',
        tdata: files,
      })
    ).rejects.toThrow('Invalid TData files')

    await waitFor(() => {
      expect(result.current.isError).toBe(true)
    })
  })

  it('should handle invalid API credentials', async () => {
    const mockError = new Error('Invalid API credentials')
    vi.mocked(sessionsApi.importTData).mockRejectedValue(mockError)

    const { result } = renderHook(() => useImportTData(), { wrapper })

    const files: File[] = [new File(['content'], 'session.dat')]

    await expect(
      result.current.mutateAsync({
        api_id: 0,
        api_hash: 'invalid',
        tdata: files,
      })
    ).rejects.toThrow('Invalid API credentials')
  })

  it('should handle import without session_name', async () => {
    const mockResponse = {
      session: {
        session_id: 'session-1',
        is_active: true,
        auth_state: 'authenticated',
        auth_method: 'tdata' as const,
      },
    }

    vi.mocked(sessionsApi.importTData).mockResolvedValue(mockResponse)

    const { result } = renderHook(() => useImportTData(), { wrapper })

    const files: File[] = [new File(['content'], 'session.dat')]

    await result.current.mutateAsync({
      api_id: 12345,
      api_hash: 'test_hash',
      tdata: files,
    })

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })

    expect(sessionsApi.importTData).toHaveBeenCalledWith({
      api_id: 12345,
      api_hash: 'test_hash',
      tdata: files,
    })
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

    vi.mocked(sessionsApi.importTData).mockResolvedValue(mockResponse)

    const { result } = renderHook(() => useImportTData(), { wrapper })

    const files: File[] = [
      new File(['content1'], 'session.dat'),
      new File(['content2'], 'key.dat'),
      new File(['content3'], 'auth.dat'),
    ]

    await result.current.mutateAsync({
      api_id: 12345,
      api_hash: 'test_hash',
      tdata: files,
    })

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })

    expect(sessionsApi.importTData).toHaveBeenCalledWith({
      api_id: 12345,
      api_hash: 'test_hash',
      tdata: files,
    })
  })
})
