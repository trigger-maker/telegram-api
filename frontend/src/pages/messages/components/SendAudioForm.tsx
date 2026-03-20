/* global React */
/* eslint-disable max-lines-per-function */
import { useState } from 'react'
import { Button, Input, Alert, FileUpload } from '@/components/common'
import { useSendAudioMessage } from '@/hooks'
import { useToast } from '@/contexts'
import { ApiException } from '@/types'
import { Send } from 'lucide-react'

interface SendAudioFormProps {
  sessionId: string
}

export const SendAudioForm = ({ sessionId }: SendAudioFormProps) => {
  const toast = useToast()
  const [to, setTo] = useState('')
  const [audioUrl, setAudioUrl] = useState('')
  const [caption, setCaption] = useState('')
  const [error, setError] = useState('')

  const sendMessage = useSendAudioMessage()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!to.trim()) {
      setError('Recipient is required')
      return
    }

    if (!audioUrl.trim()) {
      setError('Audio is required')
      return
    }

    try {
      const response = await sendMessage.mutateAsync({
        sessionId,
        data: { to: to.trim(), audio_url: audioUrl.trim(), caption: caption.trim() || undefined },
      })

      toast.success('Audio sent', `Job ID: ${response.job_id}`)
      setTo('')
      setAudioUrl('')
      setCaption('')
    } catch (err) {
      if (err instanceof ApiException) {
        setError(err.message)
      } else {
        setError('Error sending audio')
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
        type="audio"
        label="Audio"
        value={audioUrl}
        onChange={setAudioUrl}
        placeholder="https://your-server.com/audio.mp3"
        disabled={sendMessage.isPending}
      />

      <Input
        label="Caption (Optional)"
        type="text"
        placeholder="Audio description..."
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
        Send Audio
      </Button>
    </form>
  )
}
