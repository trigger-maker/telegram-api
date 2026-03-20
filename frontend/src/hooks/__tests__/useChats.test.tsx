/**
 * Tests for chats hooks
 */

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useContacts, useInfiniteContacts, contactKeys } from '../useChats'
import { getContacts, GetContactsParams } from '@/api/chats.api'

// Mock the chats API
vi.mock('@/api/chats.api', () => ({
  getContacts: vi.fn(),
  GetContactsParams: {},
}))

describe('useContacts', () => {
  let queryClient: QueryClient

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
          staleTime: 0,
        },
      },
    })
    vi.clearAllMocks()
  })

  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )

  it('should fetch contacts successfully', async () => {
    const mockResponse = {
      contacts: [
        {
          id: 1,
          first_name: 'John',
          last_name: 'Doe',
          username: 'johndoe',
          phone: '+1234567890',
          is_mutual: true,
          is_blocked: false,
        },
      ],
      total_count: 1,
      has_more: false,
      from_cache: false,
    }

    vi.mocked(getContacts).mockResolvedValue(mockResponse)

    const { result } = renderHook(
      () => useContacts('session-1', { limit: 50, offset: 0 }),
      { wrapper }
    )

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })

    expect(result.current.data).toEqual(mockResponse)
    expect(getContacts).toHaveBeenCalledWith('session-1', {
      limit: 50,
      offset: 0,
    })
  })

  it('should fetch contacts with refresh parameter', async () => {
    const mockResponse = {
      contacts: [],
      total_count: 0,
      has_more: false,
      from_cache: false,
    }

    vi.mocked(getContacts).mockResolvedValue(mockResponse)

    const { result } = renderHook(
      () => useContacts('session-1', { limit: 50, offset: 0, refresh: true }),
      { wrapper }
    )

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })

    expect(getContacts).toHaveBeenCalledWith('session-1', {
      limit: 50,
      offset: 0,
      refresh: true,
    })
  })

  it('should fetch contacts without parameters', async () => {
    const mockResponse = {
      contacts: [],
      total_count: 0,
      has_more: false,
      from_cache: false,
    }

    vi.mocked(getContacts).mockResolvedValue(mockResponse)

    const { result } = renderHook(() => useContacts('session-1'), { wrapper })

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })

    expect(getContacts).toHaveBeenCalledWith('session-1', undefined)
  })

  it('should handle fetch error', async () => {
    const mockError = new Error('Session not found')
    vi.mocked(getContacts).mockRejectedValue(mockError)

    const { result } = renderHook(() => useContacts('invalid-session'), {
      wrapper,
    })

    await waitFor(() => {
      expect(result.current.isError).toBe(true)
    })

    expect(result.current.error).toEqual(mockError)
  })

  it('should be disabled when sessionId is empty', () => {
    const { result } = renderHook(() => useContacts(''), { wrapper })

    expect(result.current.fetchStatus).toBe('idle')
    expect(getContacts).not.toHaveBeenCalled()
  })

  it('should have correct query key', async () => {
    const mockResponse = {
      contacts: [],
      total_count: 0,
      has_more: false,
      from_cache: false,
    }

    vi.mocked(getContacts).mockResolvedValue(mockResponse)

    const params: GetContactsParams = { limit: 50, offset: 0, refresh: true }
    const expectedKey = contactKeys.list('session-1', params)

    const { result } = renderHook(() => useContacts('session-1', params), {
      wrapper,
    })

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })

    expect(result.current.queryKey).toEqual(expectedKey)
  })
})

describe('useInfiniteContacts', () => {
  let queryClient: QueryClient

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
          staleTime: 0,
        },
      },
    })
    vi.clearAllMocks()
  })

  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )

  it('should fetch infinite contacts successfully', async () => {
    const mockResponse = {
      contacts: [
        {
          id: 1,
          first_name: 'John',
          last_name: 'Doe',
          username: 'johndoe',
          phone: '+1234567890',
          is_mutual: true,
          is_blocked: false,
        },
      ],
      total_count: 100,
      has_more: true,
      from_cache: false,
    }

    vi.mocked(getContacts).mockResolvedValue(mockResponse)

    const { result } = renderHook(
      () => useInfiniteContacts('session-1', undefined, 50),
      { wrapper }
    )

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })

    expect(result.current.data?.pages).toHaveLength(1)
    expect(result.current.data?.pages[0]).toEqual(mockResponse)
    expect(getContacts).toHaveBeenCalledWith('session-1', {
      limit: 50,
      offset: 0,
    })
  })

  it('should fetch next page', async () => {
    const mockPage1 = {
      contacts: Array.from({ length: 50 }, (_, i) => ({
        id: i + 1,
        first_name: `User${i}`,
        is_mutual: false,
        is_blocked: false,
      })),
      total_count: 100,
      has_more: true,
      from_cache: false,
    }

    const mockPage2 = {
      contacts: Array.from({ length: 50 }, (_, i) => ({
        id: i + 51,
        first_name: `User${i + 50}`,
        is_mutual: false,
        is_blocked: false,
      })),
      total_count: 100,
      has_more: false,
      from_cache: false,
    }

    vi.mocked(getContacts)
      .mockResolvedValueOnce(mockPage1)
      .mockResolvedValueOnce(mockPage2)

    const { result } = renderHook(
      () => useInfiniteContacts('session-1', 50),
      { wrapper }
    )

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })

    await result.current.fetchNextPage()

    await waitFor(() => {
      expect(result.current.data?.pages).toHaveLength(2)
    })

    expect(getContacts).toHaveBeenCalledTimes(2)
    expect(getContacts).toHaveBeenLastCalledWith('session-1', {
      limit: 50,
      offset: 50,
    })
  })

  it('should stop fetching when has_more is false', async () => {
    const mockResponse = {
      contacts: [],
      total_count: 0,
      has_more: false,
      from_cache: false,
    }

    vi.mocked(getContacts).mockResolvedValue(mockResponse)

    const { result } = renderHook(
      () => useInfiniteContacts('session-1', 50),
      { wrapper }
    )

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })

    expect(result.current.hasNextPage).toBe(false)
  })

  it('should have correct query key', async () => {
    const mockResponse = {
      contacts: [],
      total_count: 0,
      has_more: false,
      from_cache: false,
    }

    vi.mocked(getContacts).mockResolvedValue(mockResponse)

    const expectedKey = contactKeys.infinite('session-1')

    const { result } = renderHook(
      () => useInfiniteContacts('session-1', 50),
      { wrapper }
    )

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })

    expect(result.current.queryKey).toEqual(expectedKey)
  })

  it('should be disabled when sessionId is empty', () => {
    const { result } = renderHook(() => useInfiniteContacts('', 50), {
      wrapper,
    })

    expect(result.current.fetchStatus).toBe('idle')
    expect(getContacts).not.toHaveBeenCalled()
  })

  it('should handle fetch error', async () => {
    const mockError = new Error('Session not found')
    vi.mocked(getContacts).mockRejectedValue(mockError)

    const { result } = renderHook(() => useInfiniteContacts('invalid-session', 50), {
      wrapper,
    })

    await waitFor(() => {
      expect(result.current.isError).toBe(true)
    })

    expect(result.current.error).toEqual(mockError)
  })
})
