/* global React */
/* eslint-disable max-lines-per-function */
import { useState } from 'react'
import { Modal, Button, Input, Alert } from '@/components/common'
import { useVerifyCode } from '@/hooks'
import { ApiException } from '@/types'
import { MessageSquare } from 'lucide-react'

interface VerifySMSModalProps {
  isOpen: boolean
  onClose: () => void
  sessionId: string
  phoneNumber: string
  onSuccess: (hint?: string) => void
}

export const VerifySMSModal = ({
  isOpen,
  onClose,
  sessionId,
  phoneNumber,
  onSuccess,
}: VerifySMSModalProps) => {
  const [code, setCode] = useState('')
  const [error, setError] = useState('')

  const verifyCode = useVerifyCode()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!code || code.length < 5) {
      setError('Code must be at least 5 digits')
      return
    }

    try {
      const response = await verifyCode.mutateAsync({ sessionId, code })
      setCode('')
      // Pass hint if session requires 2FA password
      const hint = response.hint
      onSuccess(hint)
      onClose()
    } catch (err) {
      if (err instanceof ApiException) {
        setError(err.message)
      } else {
        setError('Error verifying code')
      }
    }
  }

  const handleClose = () => {
    setCode('')
    setError('')
    onClose()
  }

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Verify SMS Code"
      size="sm"
    >
      <div className="p-6">
        <div className="flex items-center justify-center mb-6">
          <div className="w-16 h-16 rounded-full bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center">
            <MessageSquare className="w-8 h-8 text-primary-600 dark:text-primary-400" />
          </div>
        </div>

        <p className="text-center text-gray-600 dark:text-gray-400 mb-6">
          We have sent a verification code to{' '}
          <span className="font-semibold text-gray-900 dark:text-white">{phoneNumber}</span>
        </p>

        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <Alert variant="error">
              {error}
            </Alert>
          )}

          <Input
            label="Verification Code"
            type="text"
            placeholder="12345"
            value={code}
            onChange={(e) => setCode(e.target.value.replace(/\D/g, ''))}
            maxLength={6}
            disabled={verifyCode.isPending}
            autoFocus
          />

          <div className="flex gap-3">
            <Button
              type="button"
              variant="secondary"
              onClick={handleClose}
              disabled={verifyCode.isPending}
              fullWidth
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              isLoading={verifyCode.isPending}
              fullWidth
            >
              Verify
            </Button>
          </div>
        </form>

        <p className="text-xs text-center text-gray-500 dark:text-gray-400 mt-4">
          Didn't receive the code? Wait a moment and check your Telegram messages.
        </p>
      </div>
    </Modal>
  )
}
