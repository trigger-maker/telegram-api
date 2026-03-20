import { useState } from 'react'
import { Smartphone, QrCode, HelpCircle, Database } from 'lucide-react'
import { Modal, Button, Input, Tabs, Alert } from '@/components/common'
import { TelegramGuide } from './TelegramGuide'
import { useCreateSession } from '@/hooks'
import { ApiException, CreateSessionRequest, CreateSessionResponse } from '@/types'

interface CreateSessionModalProps {
  isOpen: boolean
  onClose: () => void
  onSuccess: (sessionId: string, data: CreateSessionResponse) => void
  onImportTData?: () => void
}

/* eslint-disable max-lines-per-function, complexity */
export const CreateSessionModal = ({ isOpen, onClose, onSuccess, onImportTData }: CreateSessionModalProps) => {
  const [activeTab, setActiveTab] = useState('form')
  const [authMethod, setAuthMethod] = useState<'sms' | 'qr' | 'tdata'>('sms')
  const [formData, setFormData] = useState({
    session_name: '',
    phone: '',
    api_id: '',
    api_hash: '',
  })
  const [error, setError] = useState('')

  const createSession = useCreateSession()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    // Validations
    if (!formData.session_name.trim()) {
      setError('Session name is required')
      return
    }

    if (!formData.api_id || !formData.api_hash) {
      setError('API ID and API Hash are required')
      return
    }

    if (authMethod === 'sms' && !formData.phone.trim()) {
      setError('Phone number is required for SMS')
      return
    }

    try {
      const payload: CreateSessionRequest = {
        session_name: formData.session_name,
        api_id: parseInt(formData.api_id),
        api_hash: formData.api_hash,
        auth_method: authMethod,
      }

      if (authMethod === 'sms') {
        payload.phone = formData.phone
      }

      const response = await createSession.mutateAsync(payload)
      onSuccess(response.session.id, response)
    } catch (err) {
      if (err instanceof ApiException) {
        setError(err.message)
      } else {
        setError('Error creating session')
      }
    }
  }

  const handleReset = () => {
    setFormData({
      session_name: '',
      phone: '',
      api_id: '',
      api_hash: '',
    })
    setError('')
    setAuthMethod('sms')
    setActiveTab('form')
  }

  const handleClose = () => {
    handleReset()
    onClose()
  }

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="New Telegram Session"
      size="lg"
    >
      <div className="p-4 sm:p-6">
        <div className="overflow-x-auto -mx-4 sm:mx-0 px-4 sm:px-0">
          <Tabs
            tabs={[
              { id: 'form', label: 'Create Session', icon: <Smartphone className="w-4 h-4" /> },
              { id: 'guide', label: 'Credentials', icon: <HelpCircle className="w-4 h-4" /> },
            ]}
            activeTab={activeTab}
            onChange={setActiveTab}
          />
        </div>

        <div className="mt-4 sm:mt-6">
          {activeTab === 'guide' ? (
            <TelegramGuide />
          ) : (
            <form onSubmit={handleSubmit} className="space-y-4 sm:space-y-6">
              {error && (
                <Alert variant="error">
                  {error}
                </Alert>
              )}

              <Input
                label="Session Name"
                type="text"
                placeholder="My Telegram Session"
                value={formData.session_name}
                onChange={(e) => setFormData({ ...formData, session_name: e.target.value })}
                disabled={createSession.isPending}
              />

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Input
                  label="API ID"
                  type="number"
                  placeholder="12345678"
                  value={formData.api_id}
                  onChange={(e) => setFormData({ ...formData, api_id: e.target.value })}
                  disabled={createSession.isPending}
                />

                <Input
                  label="API Hash"
                  type="text"
                  placeholder="abc123def456..."
                  value={formData.api_hash}
                  onChange={(e) => setFormData({ ...formData, api_hash: e.target.value })}
                  disabled={createSession.isPending}
                />
              </div>

              <div className="space-y-3">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                  Authentication Method
                </label>

                <div className="grid grid-cols-3 gap-3">
                  <button
                    type="button"
                    onClick={() => setAuthMethod('sms')}
                    disabled={createSession.isPending}
                    className={`p-4 rounded-lg border-2 transition-all ${
                      authMethod === 'sms'
                        ? 'border-primary-600 bg-primary-50 dark:bg-primary-900/20'
                        : 'border-gray-200 dark:border-gray-700 hover:border-primary-300 dark:hover:border-primary-700'
                    }`}
                  >
                    <Smartphone
                      className={`w-6 h-6 mx-auto mb-2 ${
                        authMethod === 'sms' ? 'text-primary-600 dark:text-primary-400' : 'text-gray-400'
                      }`}
                    />
                    <p className="text-sm font-medium text-gray-900 dark:text-white">SMS</p>
                    <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                      Code via SMS
                    </p>
                  </button>

                  <button
                    type="button"
                    onClick={() => setAuthMethod('qr')}
                    disabled={createSession.isPending}
                    className={`p-4 rounded-lg border-2 transition-all ${
                      authMethod === 'qr'
                        ? 'border-primary-600 bg-primary-50 dark:bg-primary-900/20'
                        : 'border-gray-200 dark:border-gray-700 hover:border-primary-300 dark:hover:border-primary-700'
                    }`}
                  >
                    <QrCode
                      className={`w-6 h-6 mx-auto mb-2 ${
                        authMethod === 'qr' ? 'text-primary-600 dark:text-primary-400' : 'text-gray-400'
                      }`}
                    />
                    <p className="text-sm font-medium text-gray-900 dark:text-white">QR Code</p>
                    <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                      Scan with Telegram
                    </p>
                  </button>

                  <button
                    type="button"
                    onClick={() => {
                      setAuthMethod('tdata')
                      if (onImportTData) {
                        onImportTData()
                      }
                    }}
                    disabled={createSession.isPending}
                    className={`p-4 rounded-lg border-2 transition-all ${
                      authMethod === 'tdata'
                        ? 'border-primary-600 bg-primary-50 dark:bg-primary-900/20'
                        : 'border-gray-200 dark:border-gray-700 hover:border-primary-300 dark:hover:border-primary-700'
                    }`}
                  >
                    <Database
                      className={`w-6 h-6 mx-auto mb-2 ${
                        authMethod === 'tdata' ? 'text-primary-600 dark:text-primary-400' : 'text-gray-400'
                      }`}
                    />
                    <p className="text-sm font-medium text-gray-900 dark:text-white">TData</p>
                    <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                      Import from Desktop
                    </p>
                  </button>
                </div>
              </div>

              {authMethod === 'sms' && (
                <Input
                  label="Phone Number"
                  type="tel"
                  placeholder="+573001234567"
                  value={formData.phone}
                  onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
                  disabled={createSession.isPending}
                />
              )}

              {authMethod === 'qr' && (
                <Alert variant="info">
                  <p className="text-sm">
                    A QR code will be generated that you can scan from your Telegram app.
                    The QR will be automatically regenerated if it expires (maximum 3 attempts).
                  </p>
                </Alert>
              )}

              {authMethod === 'tdata' && (
                <Alert variant="info">
                  <p className="text-sm">
                    Import an existing session from Telegram Desktop. You need to export
                    the TData files from your Telegram Desktop installation first.
                  </p>
                </Alert>
              )}

              <div className="flex flex-col-reverse sm:flex-row gap-3 pt-4">
                <Button
                  type="button"
                  variant="secondary"
                  onClick={handleClose}
                  disabled={createSession.isPending}
                  fullWidth
                >
                  Cancel
                </Button>
                <Button
                  type="submit"
                  variant="primary"
                  isLoading={createSession.isPending}
                  fullWidth
                >
                  {authMethod === 'sms' ? 'Send Code' : authMethod === 'qr' ? 'Generate QR' : 'Import TData'}
                </Button>
              </div>
            </form>
          )}
        </div>
      </div>
    </Modal>
  )
}
