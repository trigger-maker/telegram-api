/* global React, File */
/* eslint-disable max-lines-per-function */
import { useState } from 'react'
import { Modal, Button, Input, Alert, FileUpload } from '@/components/common'
import { useImportTData } from '@/hooks'
import { FileText, X } from 'lucide-react'

interface ImportTDataModalProps {
  isOpen: boolean
  onClose: () => void
  onSuccess: (sessionId: string) => void
}

export const ImportTDataModal = ({
  isOpen,
  onClose,
  onSuccess,
}: ImportTDataModalProps) => {
  const [formData, setFormData] = useState({
    api_id: '',
    api_hash: '',
    session_name: '',
  })
  const [tdataFiles, setTdataFiles] = useState<File[]>([])
  const [error, setError] = useState('')
  const [validationErrors, setValidationErrors] = useState({
    api_id: '',
    api_hash: '',
    files: '',
  })

  const importTData = useImportTData()

  const handleRemoveFile = (index: number) => {
    setTdataFiles((prev) => prev.filter((_, i) => i !== index))
  }

  const validateForm = () => {
    const errors = {
      api_id: '',
      api_hash: '',
      files: '',
    }
    let isValid = true

    if (!formData.api_id.trim()) {
      errors.api_id = 'API ID is required'
      isValid = false
    }

    if (!formData.api_hash.trim()) {
      errors.api_hash = 'API Hash is required'
      isValid = false
    }

    if (tdataFiles.length === 0) {
      errors.files = 'At least one TData file is required'
      isValid = false
    }

    setValidationErrors(errors)
    return isValid
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!validateForm()) {
      return
    }

    try {
      const response = await importTData.mutateAsync({
        api_id: parseInt(formData.api_id),
        api_hash: formData.api_hash,
        session_name: formData.session_name || undefined,
        tdata: tdataFiles,
      })
      onSuccess(response.session.session_id)
      handleClose()
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message)
      } else {
        setError('Failed to import TData session')
      }
    }
  }

  const handleClose = () => {
    setFormData({
      api_id: '',
      api_hash: '',
      session_name: '',
    })
    setTdataFiles([])
    setError('')
    setValidationErrors({
      api_id: '',
      api_hash: '',
      files: '',
    })
    onClose()
  }

  if (!isOpen) return null

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Import TData Session"
      size="lg"
    >
      <div className="p-6">
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && <Alert variant="error">{error}</Alert>}

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <Input
                label="API ID"
                type="number"
                placeholder="12345678"
                value={formData.api_id}
                onChange={(e) =>
                  setFormData({ ...formData, api_id: e.target.value })
                }
                disabled={importTData.isPending}
                error={validationErrors.api_id}
              />
            </div>

            <div>
              <Input
                label="API Hash"
                type="text"
                placeholder="abc123def456..."
                value={formData.api_hash}
                onChange={(e) =>
                  setFormData({ ...formData, api_hash: e.target.value })
                }
                disabled={importTData.isPending}
                error={validationErrors.api_hash}
              />
            </div>
          </div>

          <Input
            label="Session Name (optional)"
            type="text"
            placeholder="My TData Session"
            value={formData.session_name}
            onChange={(e) =>
              setFormData({ ...formData, session_name: e.target.value })
            }
            disabled={importTData.isPending}
          />

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              TData Files
            </label>
            <FileUpload
              type="file"
              value=""
              onChange={() => {}}
              accept=".dat,.session,.key"
              disabled={importTData.isPending}
            />

            {validationErrors.files && (
              <p className="text-sm text-red-600 dark:text-red-400 mt-2">
                {validationErrors.files}
              </p>
            )}

            {tdataFiles.length > 0 && (
              <div className="mt-4 space-y-2">
                <p className="text-sm font-medium text-gray-700 dark:text-gray-300">
                  Selected files ({tdataFiles.length}):
                </p>
                <div className="space-y-2 max-h-48 overflow-y-auto">
                  {tdataFiles.map((file, index) => (
                    <div
                      key={`${file.name}-${index}`}
                      className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg"
                    >
                      <div className="flex items-center gap-3">
                        <FileText className="w-5 h-5 text-gray-400" />
                        <div>
                          <p className="text-sm font-medium text-gray-900 dark:text-white">
                            {file.name}
                          </p>
                          <p className="text-xs text-gray-500 dark:text-gray-400">
                            {(file.size / 1024).toFixed(2)} KB
                          </p>
                        </div>
                      </div>
                      <Button
                        type="button"
                        variant="secondary"
                        onClick={() => handleRemoveFile(index)}
                        disabled={importTData.isPending}
                      >
                        <X className="w-4 h-4" />
                      </Button>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>

          <Alert variant="info">
            <p className="text-sm">
              <strong>Note:</strong> Importing TData from Telegram Desktop allows you
              to use existing sessions without re-authenticating. Make sure to export
              your TData from Telegram Desktop first.
            </p>
          </Alert>

          <div className="flex gap-3 pt-4">
            <Button
              type="button"
              variant="secondary"
              onClick={handleClose}
              disabled={importTData.isPending}
              fullWidth
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              isLoading={importTData.isPending}
              fullWidth
            >
              Import
            </Button>
          </div>
        </form>
      </div>
    </Modal>
  )
}
