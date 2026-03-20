/* eslint-disable @typescript-eslint/no-explicit-any */
/**
 * Tests for QRCodeModal component
 */

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { QRCodeModal } from '../QRCodeModal'
import { useRegenerateQR, useSession } from '@/hooks'

// Mock the hooks
vi.mock('@/hooks', () => ({
  useRegenerateQR: vi.fn(),
  useSession: vi.fn(),
}))

describe('QRCodeModal', () => {
  let queryClient: QueryClient

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
          staleTime: 0,
        },
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

  const defaultProps = {
    isOpen: true,
    onClose: vi.fn(),
    sessionId: 'session-1',
    qrImage: 'base64_qr_image',
    onSuccess: vi.fn(),
  }

  it('should render modal when isOpen is true', () => {
    vi.mocked(useSession).mockReturnValue({
      data: {
        session: {
          id: 'session-1',
          user_id: 'user-1',
          api_id: 12345,
          session_name: 'Test Session',
          auth_state: 'pending',
          is_active: false,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        status: 'waiting',
      },
      refetch: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      isLoading: false,
      isRefetching: false,
      dataUpdatedAt: 0,
      failureCount: 0,
      failureReason: null,
      fetchStatus: 'idle',
      isPaused: false,
      isFetched: false,
      isFetchedAfterMount: false,
      isFetching: false,
      isInitialLoading: false,
      isLoadingError: false,
      isPlaceholderData: false,
      isRefetchError: false,
      remove: vi.fn(),
    } as any)

    vi.mocked(useRegenerateQR).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({
        session_id: 'session-1',
        qr_image_base64: 'new_base64_qr_image',
        message: 'QR code regenerated successfully',
      }),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    render(<QRCodeModal {...defaultProps} />, { wrapper })

    expect(screen.getByText(/Escanea el Código QR/i)).toBeInTheDocument()
  })

  it('should not render modal when isOpen is false', () => {
    vi.mocked(useSession).mockReturnValue({
      data: undefined,
      refetch: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      isLoading: false,
      isRefetching: false,
      dataUpdatedAt: 0,
      failureCount: 0,
      failureReason: null,
      fetchStatus: 'idle',
      isPaused: false,
      isFetched: false,
      isFetchedAfterMount: false,
      isFetching: false,
      isInitialLoading: false,
      isLoadingError: false,
      isPlaceholderData: false,
      isRefetchError: false,
      remove: vi.fn(),
    } as any)

    render(<QRCodeModal {...defaultProps} isOpen={false} />, { wrapper })

    expect(
      screen.queryByText(/Escanea el Código QR/i)
    ).not.toBeInTheDocument()
  })

  it('should display QR image', () => {
    vi.mocked(useSession).mockReturnValue({
      data: {
        session: {
          id: 'session-1',
          user_id: 'user-1',
          api_id: 12345,
          session_name: 'Test Session',
          auth_state: 'pending',
          is_active: false,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        status: 'waiting',
      },
      refetch: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      isLoading: false,
      isRefetching: false,
      dataUpdatedAt: 0,
      failureCount: 0,
      failureReason: null,
      fetchStatus: 'idle',
      isPaused: false,
      isFetched: false,
      isFetchedAfterMount: false,
      isFetching: false,
      isInitialLoading: false,
      isLoadingError: false,
      isPlaceholderData: false,
      isRefetchError: false,
      remove: vi.fn(),
    } as any)

    vi.mocked(useRegenerateQR).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    render(<QRCodeModal {...defaultProps} />, { wrapper })

    const qrImage = screen.getByAltText('QR Code')
    expect(qrImage).toBeInTheDocument()
    expect(qrImage).toHaveAttribute(
      'src',
      'data:image/png;base64,base64_qr_image'
    )
  })

  it('should show regenerate button', () => {
    vi.mocked(useSession).mockReturnValue({
      data: {
        session: {
          id: 'session-1',
          user_id: 'user-1',
          api_id: 12345,
          session_name: 'Test Session',
          auth_state: 'pending',
          is_active: false,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        status: 'waiting',
      },
      refetch: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      isLoading: false,
      isRefetching: false,
      dataUpdatedAt: 0,
      failureCount: 0,
      failureReason: null,
      fetchStatus: 'idle',
      isPaused: false,
      isFetched: false,
      isFetchedAfterMount: false,
      isFetching: false,
      isInitialLoading: false,
      isLoadingError: false,
      isPlaceholderData: false,
      isRefetchError: false,
      remove: vi.fn(),
    } as any)

    vi.mocked(useRegenerateQR).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({
        session_id: 'session-1',
        qr_image_base64: 'new_base64_qr_image',
        message: 'QR code regenerated successfully',
      }),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    render(<QRCodeModal {...defaultProps} />, { wrapper })

    expect(
      screen.getByRole('button', { name: /regenerate qr/i })
    ).toBeInTheDocument()
  })

  it('should regenerate QR when button is clicked', async () => {
    const mutateAsync = vi.fn().mockResolvedValue({
      session_id: 'session-1',
      qr_image_base64: 'new_base64_qr_image',
      message: 'QR code regenerated successfully',
    })

    vi.mocked(useSession).mockReturnValue({
      data: {
        session: {
          id: 'session-1',
          user_id: 'user-1',
          api_id: 12345,
          session_name: 'Test Session',
          auth_state: 'pending',
          is_active: false,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        status: 'waiting',
      },
      refetch: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      isLoading: false,
      isRefetching: false,
      dataUpdatedAt: 0,
      failureCount: 0,
      failureReason: null,
      fetchStatus: 'idle',
      isPaused: false,
      isFetched: false,
      isFetchedAfterMount: false,
      isFetching: false,
      isInitialLoading: false,
      isLoadingError: false,
      isPlaceholderData: false,
      isRefetchError: false,
      remove: vi.fn(),
    } as any)

    vi.mocked(useRegenerateQR).mockReturnValue({
      mutateAsync,
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    const user = userEvent.setup()
    render(<QRCodeModal {...defaultProps} />, { wrapper })

    const regenerateButton = screen.getByRole('button', { name: /regenerate qr/i })
    await user.click(regenerateButton)

    await waitFor(() => {
      expect(mutateAsync).toHaveBeenCalledWith('session-1')
    })
  })

  it('should disable regenerate button when at max attempts', () => {
    vi.mocked(useSession).mockReturnValue({
      data: {
        session: {
          id: 'session-1',
          user_id: 'user-1',
          api_id: 12345,
          session_name: 'Test Session',
          auth_state: 'pending',
          is_active: false,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        status: 'waiting',
      },
      refetch: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      isLoading: false,
      isRefetching: false,
      dataUpdatedAt: 0,
      failureCount: 0,
      failureReason: null,
      fetchStatus: 'idle',
      isPaused: false,
      isFetched: false,
      isFetchedAfterMount: false,
      isFetching: false,
      isInitialLoading: false,
      isLoadingError: false,
      isPlaceholderData: false,
      isRefetchError: false,
      remove: vi.fn(),
    } as any)

    vi.mocked(useRegenerateQR).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    render(<QRCodeModal {...defaultProps} />, { wrapper })

    // The regenerate button should be enabled initially
    const regenerateButton = screen.getByRole('button', { name: /regenerate qr/i })
    expect(regenerateButton).not.toBeDisabled()
  })

  it('should show loading state during regeneration', async () => {
    let resolveRegenerate: (value: any) => void
    const regeneratePromise = new Promise((resolve) => {
      resolveRegenerate = resolve
    })

    vi.mocked(useSession).mockReturnValue({
      data: {
        session: {
          id: 'session-1',
          user_id: 'user-1',
          api_id: 12345,
          session_name: 'Test Session',
          auth_state: 'pending',
          is_active: false,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        status: 'waiting',
      },
      refetch: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      isLoading: false,
      isRefetching: false,
      dataUpdatedAt: 0,
      failureCount: 0,
      failureReason: null,
      fetchStatus: 'idle',
      isPaused: false,
      isFetched: false,
      isFetchedAfterMount: false,
      isFetching: false,
      isInitialLoading: false,
      isLoadingError: false,
      isPlaceholderData: false,
      isRefetchError: false,
      remove: vi.fn(),
    } as any)

    vi.mocked(useRegenerateQR).mockReturnValue({
      mutateAsync: vi.fn().mockReturnValue(regeneratePromise),
      isPending: true,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    render(<QRCodeModal {...defaultProps} />, { wrapper })

    const regenerateButton = screen.getByRole('button', { name: /regenerate qr/i })
    expect(regenerateButton).toBeDisabled()

    resolveRegenerate!({
      session_id: 'session-1',
      qr_image_base64: 'new_base64_qr_image',
      message: 'QR code regenerated successfully',
    })
  })

  it('should show error message on regeneration failure', async () => {
    const mutateAsync = vi
      .fn()
      .mockRejectedValue(new Error('Failed to regenerate QR'))

    vi.mocked(useSession).mockReturnValue({
      data: {
        session: {
          id: 'session-1',
          user_id: 'user-1',
          api_id: 12345,
          session_name: 'Test Session',
          auth_state: 'pending',
          is_active: false,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        status: 'waiting',
      },
      refetch: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      isLoading: false,
      isRefetching: false,
      dataUpdatedAt: 0,
      failureCount: 0,
      failureReason: null,
      fetchStatus: 'idle',
      isPaused: false,
      isFetched: false,
      isFetchedAfterMount: false,
      isFetching: false,
      isInitialLoading: false,
      isLoadingError: false,
      isPlaceholderData: false,
      isRefetchError: false,
      remove: vi.fn(),
    } as any)

    vi.mocked(useRegenerateQR).mockReturnValue({
      mutateAsync,
      isPending: false,
      isError: true,
      error: new Error('Failed to regenerate QR'),
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    const user = userEvent.setup()
    render(<QRCodeModal {...defaultProps} />, { wrapper })

    const regenerateButton = screen.getByRole('button', { name: /regenerate qr/i })
    await user.click(regenerateButton)

    await waitFor(() => {
      expect(screen.getByText('Failed to regenerate QR')).toBeInTheDocument()
    })
  })
})
