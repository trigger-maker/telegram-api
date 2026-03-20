/* global setInterval, setTimeout, clearInterval */
/* eslint-disable max-lines-per-function, complexity */
import { useEffect, useState } from 'react'
import { Modal, Alert, Button } from '@/components/common'
import { useSession, useRegenerateQR } from '@/hooks'
import { QrCode, Loader2, CheckCircle, XCircle, RefreshCw } from 'lucide-react'

interface QRCodeModalProps {
  isOpen: boolean
  onClose: () => void
  sessionId: string
  qrImage: string
  onSuccess: () => void
}

const MAX_REGENERATE_ATTEMPTS = 3

export const QRCodeModal = ({
  isOpen,
  onClose,
  sessionId,
  qrImage: initialQrImage,
  onSuccess,
}: QRCodeModalProps) => {
  const [qrImage, setQrImage] = useState(initialQrImage)
  const [attempt, setAttempt] = useState(1)
  const [regenerateError, setRegenerateError] = useState('')
  const [status, setStatus] = useState<'waiting' | 'success' | 'failed'>('waiting')

  // Polling to check if QR was scanned
  const { refetch } = useSession(sessionId)
  const regenerateQR = useRegenerateQR()

  useEffect(() => {
    if (!isOpen) return

    const interval = setInterval(async () => {
      const result = await refetch()

      if (result.data?.session.is_active) {
        setStatus('success')
        setTimeout(() => {
          onSuccess()
          onClose()
        }, 2000)
        clearInterval(interval)
      }
    }, 3000) // Polling every 3 seconds

    return () => clearInterval(interval)
  }, [isOpen, sessionId, refetch, onSuccess, onClose])

  const handleRegenerateQR = async () => {
    if (attempt >= MAX_REGENERATE_ATTEMPTS) {
      setRegenerateError('Maximum regeneration attempts exceeded')
      return
    }

    setRegenerateError('')

    try {
      const response = await regenerateQR.mutateAsync(sessionId)
      setQrImage(response.qr_image_base64)
      setAttempt((prev) => prev + 1)
    } catch (err) {
      if (err instanceof Error) {
        setRegenerateError(err.message)
      } else {
        setRegenerateError('Failed to regenerate QR code')
      }
    }
  }

  const canRegenerate = attempt < MAX_REGENERATE_ATTEMPTS && !regenerateQR.isPending

  return (
    <Modal
      isOpen={isOpen}
      onClose={status === 'waiting' ? onClose : () => {}}
      title="Scan the QR Code"
      size="md"
      showClose={status === 'waiting'}
    >
      <div className="p-6">
        {status === 'waiting' && (
          <>
            <div className="flex items-center justify-center mb-6">
              <div className="relative">
                <div className="w-64 h-64 bg-white p-4 rounded-xl shadow-lg">
                  <img
                    src={`data:image/png;base64,${qrImage}`}
                    alt="QR Code"
                    className="w-full h-full"
                  />
                </div>
                <div className="absolute -bottom-2 -right-2 bg-primary-600 text-white rounded-full p-2">
                  <QrCode className="w-6 h-6" />
                </div>
              </div>
            </div>

            <Alert variant="info">
              <div className="space-y-2">
                <p className="font-semibold text-sm">How to scan:</p>
                <ol className="text-sm space-y-1 ml-4 list-decimal">
                  <li>Open Telegram on your phone</li>
                  <li>Go to Settings → Devices → Link Desktop Device</li>
                  <li>Scan this QR code</li>
                </ol>
              </div>
            </Alert>

            <div className="flex items-center justify-center gap-2 mt-6">
              <Loader2 className="w-5 h-5 animate-spin text-primary-600" />
              <span className="text-sm text-gray-600 dark:text-gray-400">
                Waiting for scan... (Attempt {attempt}/3)
              </span>
            </div>

            {regenerateError && (
              <Alert variant="error" className="mt-4">
                {regenerateError}
              </Alert>
            )}

            <div className="flex justify-center mt-4">
              <Button
                type="button"
                variant="secondary"
                onClick={handleRegenerateQR}
                disabled={!canRegenerate}
                isLoading={regenerateQR.isPending}
              >
                <RefreshCw className="w-4 h-4 mr-2" />
                Regenerate QR
              </Button>
            </div>

            <p className="text-xs text-center text-gray-500 dark:text-gray-400 mt-4">
              {canRegenerate
                ? 'Click to regenerate QR code if it expires'
                : 'Maximum regeneration attempts reached'}
            </p>
          </>
        )}

        {status === 'success' && (
          <div className="text-center py-8">
            <div className="flex items-center justify-center mb-4">
              <div className="w-16 h-16 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center">
                <CheckCircle className="w-10 h-10 text-green-600 dark:text-green-400" />
              </div>
            </div>
            <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-2">
              Session Created!
            </h3>
            <p className="text-gray-600 dark:text-gray-400">
              Your Telegram session is active and ready to use
            </p>
          </div>
        )}

        {status === 'failed' && (
          <div className="text-center py-8">
            <div className="flex items-center justify-center mb-4">
              <div className="w-16 h-16 rounded-full bg-red-100 dark:bg-red-900/30 flex items-center justify-center">
                <XCircle className="w-10 h-10 text-red-600 dark:text-red-400" />
              </div>
            </div>
            <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-2">
              Error Creating Session
            </h3>
            <p className="text-gray-600 dark:text-gray-400 mb-6">
              Maximum attempts reached. Please try again.
            </p>
            <Button variant="primary" onClick={onClose}>
              Close
            </Button>
          </div>
        )}
      </div>
    </Modal>
  )
}
