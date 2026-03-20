/* global React */
/* eslint-disable max-lines-per-function */
import { useState } from 'react'
import { Button, Input, Alert, FileUpload } from '@/components/common'
import { useSendPhotoMessage } from '@/hooks'
import { useToast } from '@/contexts'
import { ApiException } from '@/types'
import { Send } from 'lucide-react'

interface SendPhotoFormProps {
  sessionId: string
}

export const SendPhotoForm = ({ sessionId }: SendPhotoFormProps) => {
  const toast = useToast()
  const [to, setTo] = useState('')
  const [photoUrl, setPhotoUrl] = useState('')
  const [caption, setCaption] = useState('')
  const [error, setError] = useState('')

  const sendMessage = useSendPhotoMessage()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!to.trim()) {
      setError('Recipient is required')
      return
    }

    if (!photoUrl.trim()) {
      setError('Image is required')
      return
    }

    try {
      const response = await sendMessage.mutateAsync({
        sessionId,
        data: { to: to.trim(), photo_url: photoUrl.trim(), caption: caption.trim() || undefined },
      })

      toast.success('Photo sent', `Job ID: ${response.job_id}`)
      setTo('')
      setPhotoUrl('')
      setCaption('')
    } catch (err) {
      if (err instanceof ApiException) {
        setError(err.message)
      } else {
        setError('Error sending photo')
      }
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {error && <Alert variant="error">{error}</Alert>}

      <Input
        label="Recipient"
        type="text"
        placeholder="@username, +573001234567 or chat ID"
        value={to}
        onChange={(e) => setTo(e.target.value)}
        disabled={sendMessage.isPending}
      />

      <FileUpload
        type="image"
        label="Image"
        value={photoUrl}
        onChange={setPhotoUrl}
        placeholder="https://your-server.com/image.jpg"
        disabled={sendMessage.isPending}
      />

      <Input
        label="Caption (Optional)"
        type="text"
        placeholder="Photo description..."
        value={caption}
        onChange={(e) => setCaption(e.target.value)}
        disabled={sendMessage.isPending}
      />

      <Button
        type="submit"
        variant="primary"
        isLoading={sendMessage.isPending}
        fullWidth
        className="h-12"
      >
        <Send className="w-4 h-4 mr-2" />
        Send Photo
      </Button>
    </form>
  )
}
