/* eslint-disable @typescript-eslint/no-explicit-any */
/**
 * Tests for ImportTDataModal component
 */

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ImportTDataModal } from '../ImportTDataModal'
import { useImportTData } from '@/hooks'

// Mock the hook
vi.mock('@/hooks', () => ({
  useImportTData: vi.fn(),
}))

describe('ImportTDataModal', () => {
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

  const defaultProps = {
    isOpen: true,
    onClose: vi.fn(),
    onSuccess: vi.fn(),
  }

  it('should render modal when isOpen is true', () => {
    vi.mocked(useImportTData).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({
        session: {
          session_id: 'session-1',
          is_active: true,
          auth_state: 'authenticated',
          auth_method: 'tdata' as const,
        },
      }),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    render(<ImportTDataModal {...defaultProps} />, { wrapper })

    expect(screen.getByText('Import TData Session')).toBeInTheDocument()
  })

  it('should not render modal when isOpen is false', () => {
    vi.mocked(useImportTData).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    render(<ImportTDataModal {...defaultProps} isOpen={false} />, { wrapper })

    expect(
      screen.queryByText('Import TData Session')
    ).not.toBeInTheDocument()
  })

  it('should have API ID input field', () => {
    vi.mocked(useImportTData).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    render(<ImportTDataModal {...defaultProps} />, { wrapper })

    expect(screen.getByLabelText('API ID')).toBeInTheDocument()
  })

  it('should have API Hash input field', () => {
    vi.mocked(useImportTData).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    render(<ImportTDataModal {...defaultProps} />, { wrapper })

    expect(screen.getByLabelText('API Hash')).toBeInTheDocument()
  })

  it('should have Session Name input field', () => {
    vi.mocked(useImportTData).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    render(<ImportTDataModal {...defaultProps} />, { wrapper })

    expect(screen.getByLabelText('Session Name (optional)')).toBeInTheDocument()
  })

  it('should have file upload area', () => {
    vi.mocked(useImportTData).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    render(<ImportTDataModal {...defaultProps} />, { wrapper })

    expect(screen.getByText(/drag and drop/i)).toBeInTheDocument()
  })

  it('should submit form with valid data', async () => {
    const mutateAsync = vi.fn().mockResolvedValue({
      session: {
        session_id: 'session-1',
        is_active: true,
        auth_state: 'authenticated',
        auth_method: 'tdata' as const,
      },
    })

    vi.mocked(useImportTData).mockReturnValue({
      mutateAsync,
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    const user = userEvent.setup()
    render(<ImportTDataModal {...defaultProps} />, { wrapper })

    const apiIdInput = screen.getByLabelText('API ID')
    await user.type(apiIdInput, '12345')

    const apiHashInput = screen.getByLabelText('API Hash')
    await user.type(apiHashInput, 'test_hash')

    const sessionNameInput = screen.getByLabelText('Session Name (optional)')
    await user.type(sessionNameInput, 'Test Session')

    // Create a mock file
    const file = new File(['content'], 'session.dat', { type: 'application/octet-stream' })
    const fileInput = screen.getByLabelText(/upload tdata files/i)
    await user.upload(fileInput, [file])

    const submitButton = screen.getByRole('button', { name: /import/i })
    await user.click(submitButton)

    await waitFor(() => {
      expect(mutateAsync).toHaveBeenCalled()
    })

    expect(defaultProps.onSuccess).toHaveBeenCalledWith('session-1')
  })

  it('should show validation error for missing API ID', async () => {
    vi.mocked(useImportTData).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    const user = userEvent.setup()
    render(<ImportTDataModal {...defaultProps} />, { wrapper })

    const apiHashInput = screen.getByLabelText('API Hash')
    await user.type(apiHashInput, 'test_hash')

    const submitButton = screen.getByRole('button', { name: /import/i })
    await user.click(submitButton)

    expect(screen.getByText('API ID is required')).toBeInTheDocument()
  })

  it('should show validation error for missing API Hash', async () => {
    vi.mocked(useImportTData).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    const user = userEvent.setup()
    render(<ImportTDataModal {...defaultProps} />, { wrapper })

    const apiIdInput = screen.getByLabelText('API ID')
    await user.type(apiIdInput, '12345')

    const submitButton = screen.getByRole('button', { name: /import/i })
    await user.click(submitButton)

    expect(screen.getByText('API Hash is required')).toBeInTheDocument()
  })

  it('should show validation error for no files uploaded', async () => {
    vi.mocked(useImportTData).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    const user = userEvent.setup()
    render(<ImportTDataModal {...defaultProps} />, { wrapper })

    const apiIdInput = screen.getByLabelText('API ID')
    await user.type(apiIdInput, '12345')

    const apiHashInput = screen.getByLabelText('API Hash')
    await user.type(apiHashInput, 'test_hash')

    const submitButton = screen.getByRole('button', { name: /import/i })
    await user.click(submitButton)

    expect(screen.getByText('At least one TData file is required')).toBeInTheDocument()
  })

  it('should show error message on import failure', async () => {
    const mutateAsync = vi
      .fn()
      .mockRejectedValue(new Error('Invalid TData files'))

    vi.mocked(useImportTData).mockReturnValue({
      mutateAsync,
      isPending: false,
      isError: true,
      error: new Error('Invalid TData files'),
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    const user = userEvent.setup()
    render(<ImportTDataModal {...defaultProps} />, { wrapper })

    const apiIdInput = screen.getByLabelText('API ID')
    await user.type(apiIdInput, '12345')

    const apiHashInput = screen.getByLabelText('API Hash')
    await user.type(apiHashInput, 'test_hash')

    const file = new File(['content'], 'session.dat')
    const fileInput = screen.getByLabelText(/upload tdata files/i)
    await user.upload(fileInput, [file])

    const submitButton = screen.getByRole('button', { name: /import/i })
    await user.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Invalid TData files')).toBeInTheDocument()
    })
  })

  it('should show loading state during import', async () => {
    let resolveImport: (value: any) => void
    const importPromise = new Promise((resolve) => {
      resolveImport = resolve
    })

    vi.mocked(useImportTData).mockReturnValue({
      mutateAsync: vi.fn().mockReturnValue(importPromise),
      isPending: true,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    render(<ImportTDataModal {...defaultProps} />, { wrapper })

    const submitButton = screen.getByRole('button', { name: /import/i })
    expect(submitButton).toBeDisabled()

    resolveImport!({
      session: {
        session_id: 'session-1',
        is_active: true,
        auth_state: 'authenticated',
        auth_method: 'tdata' as const,
      },
    })
  })

  it('should close modal on cancel button click', async () => {
    vi.mocked(useImportTData).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    const user = userEvent.setup()
    render(<ImportTDataModal {...defaultProps} />, { wrapper })

    const cancelButton = screen.getByRole('button', { name: /cancel/i })
    await user.click(cancelButton)

    expect(defaultProps.onClose).toHaveBeenCalled()
  })

  it('should display uploaded files', async () => {
    vi.mocked(useImportTData).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    const user = userEvent.setup()
    render(<ImportTDataModal {...defaultProps} />, { wrapper })

    const file = new File(['content'], 'session.dat')
    const fileInput = screen.getByLabelText(/upload tdata files/i)
    await user.upload(fileInput, [file])

    expect(screen.getByText('session.dat')).toBeInTheDocument()
  })

  it('should allow removing uploaded files', async () => {
    vi.mocked(useImportTData).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    const user = userEvent.setup()
    render(<ImportTDataModal {...defaultProps} />, { wrapper })

    const file = new File(['content'], 'session.dat')
    const fileInput = screen.getByLabelText(/upload tdata files/i)
    await user.upload(fileInput, [file])

    expect(screen.getByText('session.dat')).toBeInTheDocument()

    const removeButton = screen.getByRole('button', { name: /remove/i })
    await user.click(removeButton)

    expect(screen.queryByText('session.dat')).not.toBeInTheDocument()
  })

  it('should submit without session name', async () => {
    const mutateAsync = vi.fn().mockResolvedValue({
      session: {
        session_id: 'session-1',
        is_active: true,
        auth_state: 'authenticated',
        auth_method: 'tdata' as const,
      },
    })

    vi.mocked(useImportTData).mockReturnValue({
      mutateAsync,
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    const user = userEvent.setup()
    render(<ImportTDataModal {...defaultProps} />, { wrapper })

    const apiIdInput = screen.getByLabelText('API ID')
    await user.type(apiIdInput, '12345')

    const apiHashInput = screen.getByLabelText('API Hash')
    await user.type(apiHashInput, 'test_hash')

    const file = new File(['content'], 'session.dat')
    const fileInput = screen.getByLabelText(/upload tdata files/i)
    await user.upload(fileInput, [file])

    const submitButton = screen.getByRole('button', { name: /import/i })
    await user.click(submitButton)

    await waitFor(() => {
      expect(mutateAsync).toHaveBeenCalled()
    })

    expect(defaultProps.onSuccess).toHaveBeenCalledWith('session-1')
  })
})
