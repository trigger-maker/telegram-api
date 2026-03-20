/* eslint-disable @typescript-eslint/no-explicit-any */
/**
 * Tests for SubmitPasswordModal component
 */

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { SubmitPasswordModal } from '../SubmitPasswordModal'
import { useSubmitPassword } from '@/hooks'

// Mock the hook
vi.mock('@/hooks', () => ({
  useSubmitPassword: vi.fn(),
}))

describe('SubmitPasswordModal', () => {
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
    sessionId: 'session-1',
    onSuccess: vi.fn(),
  }

  it('should render modal when isOpen is true', () => {
    vi.mocked(useSubmitPassword).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    render(<SubmitPasswordModal {...defaultProps} />, { wrapper })

    expect(screen.getByText('Two-Step Verification')).toBeInTheDocument()
    expect(screen.getByLabelText('Password')).toBeInTheDocument()
  })

  it('should not render modal when isOpen is false', () => {
    vi.mocked(useSubmitPassword).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    render(<SubmitPasswordModal {...defaultProps} isOpen={false} />, { wrapper })

    expect(screen.queryByText('Two-Step Verification')).not.toBeInTheDocument()
  })

  it('should display hint if provided', () => {
    vi.mocked(useSubmitPassword).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    render(<SubmitPasswordModal {...defaultProps} hint="My password" />, {
      wrapper,
    })

    expect(screen.getByText('My password')).toBeInTheDocument()
  })

  it('should not display hint if not provided', () => {
    vi.mocked(useSubmitPassword).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    render(<SubmitPasswordModal {...defaultProps} />, { wrapper })

    expect(screen.queryByText('Password hint:')).not.toBeInTheDocument()
  })

  it('should submit password successfully', async () => {
    const mutateAsync = vi.fn().mockResolvedValue({})
    vi.mocked(useSubmitPassword).mockReturnValue({
      mutateAsync,
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    const user = userEvent.setup()
    render(<SubmitPasswordModal {...defaultProps} />, { wrapper })

    const passwordInput = screen.getByLabelText('Password')
    await user.type(passwordInput, 'test_password')

    const submitButton = screen.getByRole('button', { name: /submit/i })
    await user.click(submitButton)

    await waitFor(() => {
      expect(mutateAsync).toHaveBeenCalledWith({
        sessionId: 'session-1',
        password: 'test_password',
      })
    })

    expect(defaultProps.onSuccess).toHaveBeenCalled()
    expect(defaultProps.onClose).toHaveBeenCalled()
  })

  it('should show validation error for empty password', async () => {
    vi.mocked(useSubmitPassword).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    const user = userEvent.setup()
    render(<SubmitPasswordModal {...defaultProps} />, { wrapper })

    const submitButton = screen.getByRole('button', { name: /submit/i })
    await user.click(submitButton)

    expect(
      screen.getByText('Password is required')
    ).toBeInTheDocument()
  })

  it('should show error message on submission failure', async () => {
    const mutateAsync = vi.fn().mockRejectedValue(new Error('Invalid password'))
    vi.mocked(useSubmitPassword).mockReturnValue({
      mutateAsync,
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    const user = userEvent.setup()
    render(<SubmitPasswordModal {...defaultProps} />, { wrapper })

    const passwordInput = screen.getByLabelText('Password')
    await user.type(passwordInput, 'wrong_password')

    const submitButton = screen.getByRole('button', { name: /submit/i })
    await user.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Invalid password')).toBeInTheDocument()
    })
  })

  it('should show loading state during submission', async () => {
    let resolveSubmit: (value: any) => void
    const mutatePromise = new Promise((resolve) => {
      resolveSubmit = resolve
    })

    vi.mocked(useSubmitPassword).mockReturnValue({
      mutateAsync: vi.fn().mockReturnValue(mutatePromise),
      isPending: true,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    render(<SubmitPasswordModal {...defaultProps} />, { wrapper })

    const passwordInput = screen.getByLabelText('Password')
    fireEvent.change(passwordInput, { target: { value: 'test_password' } })

    const submitButton = screen.getByRole('button', { name: /submit/i })
    expect(submitButton).toBeDisabled()

    resolveSubmit!({})
  })

  it('should close modal on cancel button click', async () => {
    vi.mocked(useSubmitPassword).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    const user = userEvent.setup()
    render(<SubmitPasswordModal {...defaultProps} />, { wrapper })

    const cancelButton = screen.getByRole('button', { name: /cancel/i })
    await user.click(cancelButton)

    expect(defaultProps.onClose).toHaveBeenCalled()
  })

  it('should reset form on close', async () => {
    vi.mocked(useSubmitPassword).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    const user = userEvent.setup()
    render(<SubmitPasswordModal {...defaultProps} />, { wrapper })

    const passwordInput = screen.getByLabelText('Password')
    await user.type(passwordInput, 'test_password')
    expect(passwordInput).toHaveValue('test_password')

    const cancelButton = screen.getByRole('button', { name: /cancel/i })
    await user.click(cancelButton)

    expect(defaultProps.onClose).toHaveBeenCalled()
  })

  it('should mask password input', () => {
    vi.mocked(useSubmitPassword).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    render(<SubmitPasswordModal {...defaultProps} />, { wrapper })

    const passwordInput = screen.getByLabelText('Password') as HTMLInputElement
    expect(passwordInput.type).toBe('password')
  })

  it('should focus password input on open', () => {
    vi.mocked(useSubmitPassword).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
      isError: false,
      error: null,
      isSuccess: false,
      reset: vi.fn(),
    } as any)

    render(<SubmitPasswordModal {...defaultProps} />, { wrapper })

    const passwordInput = screen.getByLabelText('Password')
    expect(passwordInput).toHaveFocus()
  })
})
