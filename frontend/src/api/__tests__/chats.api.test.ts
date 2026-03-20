/**
 * Tests for chats API functions
 */

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { getContacts, GetContactsParams } from '../chats.api'
import { apiClient } from '../client'

// Mock the apiClient
vi.mock('../client', () => ({
  apiClient: {
    get: vi.fn(),
  },
}))

describe('getContacts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should call apiClient.get with correct URL and params', async () => {
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

    vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

    const result = await getContacts('session-1', {
      limit: 50,
      offset: 0,
    })

    expect(apiClient.get).toHaveBeenCalledWith(
      '/sessions/session-1/contacts?limit=50&offset=0'
    )
    expect(result).toEqual(mockResponse)
  })

  it('should call apiClient.get with refresh parameter', async () => {
    const mockResponse = {
      contacts: [],
      total_count: 0,
      has_more: false,
      from_cache: false,
    }

    vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

    await getContacts('session-1', {
      limit: 50,
      offset: 0,
      refresh: true,
    })

    expect(apiClient.get).toHaveBeenCalledWith(
      '/sessions/session-1/contacts?limit=50&offset=0&refresh=true'
    )
  })

  it('should call apiClient.get with refresh=false', async () => {
    const mockResponse = {
      contacts: [],
      total_count: 0,
      has_more: false,
      from_cache: false,
    }

    vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

    await getContacts('session-1', {
      limit: 50,
      offset: 0,
      refresh: false,
    })

    expect(apiClient.get).toHaveBeenCalledWith(
      '/sessions/session-1/contacts?limit=50&offset=0&refresh=false'
    )
  })

  it('should call apiClient.get without optional parameters', async () => {
    const mockResponse = {
      contacts: [],
      total_count: 0,
      has_more: false,
      from_cache: false,
    }

    vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

    const result = await getContacts('session-1')

    expect(apiClient.get).toHaveBeenCalledWith('/sessions/session-1/contacts')
    expect(result).toEqual(mockResponse)
  })

  it('should NOT include search parameter in URL', async () => {
    const mockResponse = {
      contacts: [],
      total_count: 0,
      has_more: false,
      from_cache: false,
    }

    vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

    // @ts-expect-error - Testing that search is not included
    await getContacts('session-1', {
      limit: 50,
      offset: 0,
      search: 'test',
    } as GetContactsParams)

    const callUrl = vi.mocked(apiClient.get).mock.calls[0][0]

    // Verify search is NOT in the URL
    expect(callUrl).not.toContain('search')
    expect(callUrl).toContain('limit=50')
    expect(callUrl).toContain('offset=0')
  })

  it('should handle large limit values', async () => {
    const mockResponse = {
      contacts: [],
      total_count: 0,
      has_more: false,
      from_cache: false,
    }

    vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

    await getContacts('session-1', {
      limit: 1000,
      offset: 0,
    })

    expect(apiClient.get).toHaveBeenCalledWith(
      '/sessions/session-1/contacts?limit=1000&offset=0'
    )
  })

  it('should handle offset parameter', async () => {
    const mockResponse = {
      contacts: [],
      total_count: 0,
      has_more: false,
      from_cache: false,
    }

    vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

    await getContacts('session-1', {
      limit: 50,
      offset: 100,
    })

    expect(apiClient.get).toHaveBeenCalledWith(
      '/sessions/session-1/contacts?limit=50&offset=100'
    )
  })

  it('should handle session not found error', async () => {
    const mockError = new Error('Session not found')
    vi.mocked(apiClient.get).mockRejectedValue(mockError)

    await expect(getContacts('invalid-session')).rejects.toThrow(
      'Session not found'
    )
  })

  it('should return from_cache field from response', async () => {
    const mockResponse = {
      contacts: [],
      total_count: 0,
      has_more: false,
      from_cache: true,
    }

    vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

    const result = await getContacts('session-1')

    expect(result.from_cache).toBe(true)
  })
})

describe('GetContactsParams interface', () => {
  it('should accept params with limit and offset', () => {
    const params: GetContactsParams = {
      limit: 50,
      offset: 0,
    }
    expect(params.limit).toBe(50)
    expect(params.offset).toBe(0)
  })

  it('should accept params with refresh', () => {
    const params: GetContactsParams = {
      limit: 50,
      offset: 0,
      refresh: true,
    }
    expect(params.refresh).toBe(true)
  })

  it('should accept empty params object', () => {
    const params: GetContactsParams = {}
    expect(params.limit).toBeUndefined()
    expect(params.offset).toBeUndefined()
    expect(params.refresh).toBeUndefined()
  })
})
