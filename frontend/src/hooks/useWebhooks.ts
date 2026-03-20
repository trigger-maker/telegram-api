/* eslint-disable complexity */
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  getWebhookConfig,
  createWebhook,
  deleteWebhook,
  startWebhook,
  stopWebhook,
  getPoolStatus,
  WebhookCreateRequest,
} from '@/api/webhooks.api'

// Query keys
const webhookKeys = {
  all: ['webhooks'] as const,
  config: (sessionId: string) => [...webhookKeys.all, 'config', sessionId] as const,
  pool: () => [...webhookKeys.all, 'pool'] as const,
}

/**
 * Hook to get the webhook configuration of a session
 */
export const useWebhookConfig = (sessionId: string, options?: { enabled?: boolean }) => {
  return useQuery({
    queryKey: webhookKeys.config(sessionId),
    queryFn: () => getWebhookConfig(sessionId),
    enabled: options?.enabled !== false && !!sessionId,
    retry: false,
    staleTime: 10000, // 10 seconds
  })
}

/**
 * Hook to get the state of the active sessions pool
 */
export const usePoolStatus = (options?: { enabled?: boolean; refetchInterval?: number }) => {
  return useQuery({
    queryKey: webhookKeys.pool(),
    queryFn: getPoolStatus,
    enabled: options?.enabled !== false,
    refetchInterval: options?.refetchInterval ?? 5000, // Refresh every 5s by default
    staleTime: 3000,
  })
}

/**
 * Hook to create a webhook
 */
export const useCreateWebhook = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ sessionId, data }: { sessionId: string; data: WebhookCreateRequest }) =>
      createWebhook(sessionId, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: webhookKeys.config(variables.sessionId) })
    },
  })
}

/**
 * Hook to delete a webhook
 */
export const useDeleteWebhook = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (sessionId: string) => deleteWebhook(sessionId),
    onSuccess: (_, sessionId) => {
      queryClient.invalidateQueries({ queryKey: webhookKeys.config(sessionId) })
      queryClient.invalidateQueries({ queryKey: webhookKeys.pool() })
    },
  })
}

/**
 * Hook to start webhook listening
 */
export const useStartWebhook = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (sessionId: string) => startWebhook(sessionId),
    onSuccess: (_, sessionId) => {
      queryClient.invalidateQueries({ queryKey: webhookKeys.config(sessionId) })
      queryClient.invalidateQueries({ queryKey: webhookKeys.pool() })
    },
  })
}

/**
 * Hook to stop webhook listening
 */
export const useStopWebhook = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (sessionId: string) => stopWebhook(sessionId),
    onSuccess: (_, sessionId) => {
      queryClient.invalidateQueries({ queryKey: webhookKeys.config(sessionId) })
      queryClient.invalidateQueries({ queryKey: webhookKeys.pool() })
    },
  })
}

/**
 * Combined hook to manage webhooks with all operations
 */
export const useWebhook = (sessionId: string) => {
  const config = useWebhookConfig(sessionId)
  const pool = usePoolStatus({ enabled: !!config.data })
  const createMutation = useCreateWebhook()
  const deleteMutation = useDeleteWebhook()
  const startMutation = useStartWebhook()
  const stopMutation = useStopWebhook()

  // Check if this session is actively listening in the pool
  const isListening = pool.data?.sessions?.some(
    (s) => s.session_id === sessionId && s.is_connected
  ) ?? false

  const poolSession = pool.data?.sessions?.find((s) => s.session_id === sessionId)

  return {
    // State
    config: config.data,
    isLoading: config.isLoading,
    isError: config.isError,
    error: config.error,
    isListening,
    poolSession,
    poolStatus: pool.data,

    // Actions
    create: async (data: WebhookCreateRequest, autoStart = false) => {
      const result = await createMutation.mutateAsync({ sessionId, data })
      if (autoStart) {
        await startMutation.mutateAsync(sessionId)
      }
      return result
    },
    delete: () => deleteMutation.mutateAsync(sessionId),
    start: () => startMutation.mutateAsync(sessionId),
    stop: () => stopMutation.mutateAsync(sessionId),
    refetch: () => {
      config.refetch()
      pool.refetch()
    },

    // Mutation states
    isCreating: createMutation.isPending,
    isDeleting: deleteMutation.isPending,
    isStarting: startMutation.isPending,
    isStopping: stopMutation.isPending,
    isActing: createMutation.isPending || deleteMutation.isPending ||
              startMutation.isPending || stopMutation.isPending,
  }
}
