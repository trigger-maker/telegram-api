import { useQuery, useInfiniteQuery } from '@tanstack/react-query'
import {
  getChats,
  getChatInfo,
  getChatHistory,
  getContacts,
  GetChatsParams,
  GetHistoryParams,
  GetContactsParams,
  ChatsResponse,
  Chat,
  HistoryResponse,
  ContactsResponse,
} from '@/api/chats.api'

// =============== QUERY KEYS ===============

export const chatKeys = {
  all: ['chats'] as const,
  lists: () => [...chatKeys.all, 'list'] as const,
  list: (sessionId: string, params?: GetChatsParams) =>
    [...chatKeys.lists(), sessionId, params] as const,
  details: () => [...chatKeys.all, 'detail'] as const,
  detail: (sessionId: string, chatId: number) =>
    [...chatKeys.details(), sessionId, chatId] as const,
  history: (sessionId: string, chatId: number, params?: GetHistoryParams) =>
    [...chatKeys.all, 'history', sessionId, chatId, params] as const,
}

export const contactKeys = {
  all: ['contacts'] as const,
  lists: () => [...contactKeys.all, 'list'] as const,
  list: (sessionId: string, params?: GetContactsParams) =>
    [...contactKeys.lists(), sessionId, params] as const,
  infinite: (sessionId: string) =>
    [...contactKeys.all, 'infinite', sessionId] as const,
}

// =============== HOOKS ===============

/**
 * Hook to get the list of chats of a session
 */
export const useChats = (sessionId: string, params?: GetChatsParams) => {
  return useQuery<ChatsResponse>({
    queryKey: chatKeys.list(sessionId, params),
    queryFn: () => getChats(sessionId, params),
    enabled: !!sessionId,
    staleTime: 1000 * 30, // 30 seconds
  })
}

/**
 * Hook to get information of a specific chat
 */
export const useChatInfo = (sessionId: string, chatId: number) => {
  return useQuery<Chat>({
    queryKey: chatKeys.detail(sessionId, chatId),
    queryFn: () => getChatInfo(sessionId, chatId),
    enabled: !!sessionId && !!chatId,
    staleTime: 1000 * 60, // 1 minute
  })
}

/**
 * Hook to get the message history of a chat
 * With automatic polling for real-time synchronization
 */
export const useChatHistory = (
  sessionId: string,
  chatId: number,
  params?: GetHistoryParams & { enablePolling?: boolean; pollingInterval?: number }
) => {
  const { enablePolling = true, pollingInterval = 4000, ...queryParams } = params || {}

  return useQuery<HistoryResponse>({
    queryKey: chatKeys.history(sessionId, chatId, queryParams),
    queryFn: () => getChatHistory(sessionId, chatId, queryParams),
    enabled: !!sessionId && !!chatId,
    staleTime: 1000 * 3, // 3 seconds
    refetchInterval: enablePolling ? pollingInterval : false, // Polling every 4 seconds by default
    refetchIntervalInBackground: false, // Don't refresh if tab is not active
  })
}

/**
 * Hook to get the list of contacts with pagination
 */
export const useContacts = (sessionId: string, params?: GetContactsParams) => {
  return useQuery<ContactsResponse>({
    queryKey: contactKeys.list(sessionId, params),
    queryFn: () => getContacts(sessionId, params),
    enabled: !!sessionId,
    staleTime: 1000 * 60 * 5, // 5 minutes
  })
}

/**
 * Hook to get contacts with infinite scroll
 */
export const useInfiniteContacts = (
  sessionId: string,
  limit: number = 50
) => {
  return useInfiniteQuery({
    queryKey: contactKeys.infinite(sessionId),
    queryFn: ({ pageParam = 0 }) =>
      getContacts(sessionId, { limit, offset: pageParam }),
    initialPageParam: 0,
    getNextPageParam: (lastPage, allPages) => {
      if (!lastPage.has_more) return undefined
      return allPages.length * limit
    },
    enabled: !!sessionId,
    staleTime: 1000 * 60 * 5, // 5 minutes
  })
}
