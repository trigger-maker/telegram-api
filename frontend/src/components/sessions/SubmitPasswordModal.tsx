/* global React */
/* eslint-disable max-lines-per-function */
import { useState } from 'react'
import { Modal, Button, Input, Alert } from '@/components/common'
import { useSubmitPassword } from '@/hooks'
import { ApiException } from '@/types'
import { Lock } from 'lucide-react'

interface SubmitPasswordModalProps {
  isOpen: boolean
  onClose: () => void
  sessionId: string
  hint?: string // Password hint from backend
  onSuccess: () => void
}

export const SubmitPasswordModal = ({
  isOpen,
  onClose,
  sessionId,
  hint,
  onSuccess,
}: SubmitPasswordModalProps) => {
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')

  const submitPassword = useSubmitPassword()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!password) {
      setError('Password is required')
      return
    }

    try {
      await submitPassword.mutateAsync({ sessionId, password })
      setPassword('')
      onSuccess()
      onClose()
    } catch (err) {
      if (err instanceof ApiException) {
        setError(err.message)
      } else if (err instanceof Error) {
        setError(err.message)
      } else {
        setError('Error submitting password')
      }
    }
  }

  const handleClose = () => {
    setPassword('')
    setError('')
    onClose()
  }

  if (!isOpen) return null

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Two-Step Verification"
      size="sm"
    >
      <div className="p-6">
        <div className="flex items-center justify-center mb-6">
          <div className="w-16 h-16 rounded-full bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center">
            <Lock className="w-8 h-8 text-primary-600 dark:text-primary-400" />
          </div>
        </div>

        <p className="text-center text-gray-600 dark:text-gray-400 mb-6">
          Your account has two-step verification enabled. Please enter your
          password to continue.
        </p>

        {hint && (
          <div className="mb-4 p-3 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg border border-yellow-200 dark:border-yellow-800">
            <p className="text-sm text-yellow-800 dark:text-yellow-200">
              <span className="font-semibold">Password hint:</span> {hint}
            </p>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          {error && <Alert variant="error">{error}</Alert>}

          <Input
            label="Password"
            type="password"
            placeholder="Enter your password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            disabled={submitPassword.isPending}
            autoFocus
          />

          <div className="flex gap-3">
            <Button
              type="button"
              variant="secondary"
              onClick={handleClose}
              disabled={submitPassword.isPending}
              fullWidth
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              isLoading={submitPassword.isPending}
              fullWidth
            >
              Submit
            </Button>
          </div>
        </form>

        <p className="text-xs text-center text-gray-500 dark:text-gray-400 mt-4">
          If you forgot your password, you may need to reset it in Telegram.
        </p>
      </div>
    </Modal>
  )
}
