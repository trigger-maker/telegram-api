import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { messagesApi } from '@/api'
import { chatKeys } from './useChats'
import type {
  SendTextRequest,
  SendPhotoRequest,
  SendVideoRequest,
  SendAudioRequest,
  SendFileRequest,
  SendBulkRequest,
} from '@/api/messages.api'

export const MESSAGE_QUERY_KEY = 'messages'

// Helper to invalidate chat history after sending a message
const useInvalidateChatHistory = () => {
  const queryClient = useQueryClient()

  return (sessionId: string, chatId?: string) => {
    // Invalidate all histories of this session
    queryClient.invalidateQueries({
      queryKey: ['chats', 'history', sessionId],
    })
    // If we have the specific chatId, invalidate that too
    if (chatId) {
      const numericChatId = parseInt(chatId, 10)
      if (!isNaN(numericChatId)) {
        queryClient.invalidateQueries({
          queryKey: chatKeys.history(sessionId, numericChatId),
        })
      }
    }
  }
}

export const useSendTextMessage = () => {
  const invalidateHistory = useInvalidateChatHistory()

  return useMutation({
    mutationFn: ({ sessionId, data }: { sessionId: string; data: SendTextRequest }) =>
      messagesApi.sendText(sessionId, data),
    onSuccess: (_, variables) => {
      // Invalidate the history cache so it refreshes
      invalidateHistory(variables.sessionId, variables.data.to)
    },
  })
}

export const useSendPhotoMessage = () => {
  const invalidateHistory = useInvalidateChatHistory()

  return useMutation({
    mutationFn: ({ sessionId, data }: { sessionId: string; data: SendPhotoRequest }) =>
      messagesApi.sendPhoto(sessionId, data),
    onSuccess: (_, variables) => {
      invalidateHistory(variables.sessionId, variables.data.to)
    },
  })
}

export const useSendVideoMessage = () => {
  const invalidateHistory = useInvalidateChatHistory()

  return useMutation({
    mutationFn: ({ sessionId, data }: { sessionId: string; data: SendVideoRequest }) =>
      messagesApi.sendVideo(sessionId, data),
    onSuccess: (_, variables) => {
      invalidateHistory(variables.sessionId, variables.data.to)
    },
  })
}

export const useSendAudioMessage = () => {
  const invalidateHistory = useInvalidateChatHistory()

  return useMutation({
    mutationFn: ({ sessionId, data }: { sessionId: string; data: SendAudioRequest }) =>
      messagesApi.sendAudio(sessionId, data),
    onSuccess: (_, variables) => {
      invalidateHistory(variables.sessionId, variables.data.to)
    },
  })
}

export const useSendFileMessage = () => {
  const invalidateHistory = useInvalidateChatHistory()

  return useMutation({
    mutationFn: ({ sessionId, data }: { sessionId: string; data: SendFileRequest }) =>
      messagesApi.sendFile(sessionId, data),
    onSuccess: (_, variables) => {
      invalidateHistory(variables.sessionId, variables.data.to)
    },
  })
}

export const useSendBulkMessage = () => {
  return useMutation({
    mutationFn: ({ sessionId, data }: { sessionId: string; data: SendBulkRequest }) =>
      messagesApi.sendBulk(sessionId, data),
  })
}

export const useMessageStatus = (jobId: string) => {
  return useQuery({
    queryKey: [MESSAGE_QUERY_KEY, jobId],
    queryFn: () => messagesApi.getStatus(jobId),
    enabled: !!jobId,
    refetchInterval: 3000, // Polling every 3 seconds
  })
}
