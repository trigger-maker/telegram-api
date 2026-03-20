/* global React */
/* eslint-disable max-lines-per-function */
import { useState } from 'react'
import { Button, Input, Alert, FileUpload } from '@/components/common'
import { useSendVideoMessage } from '@/hooks'
import { useToast } from '@/contexts'
import { ApiException } from '@/types'
import { Send } from 'lucide-react'

interface SendVideoFormProps {
  sessionId: string
}

export const SendVideoForm = ({ sessionId }: SendVideoFormProps) => {
  const toast = useToast()
  const [to, setTo] = useState('')
  const [videoUrl, setVideoUrl] = useState('')
  const [caption, setCaption] = useState('')
  const [error, setError] = useState('')

  const sendMessage = useSendVideoMessage()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!to.trim()) {
      setError('Recipient is required')
      return
    }

    if (!videoUrl.trim()) {
      setError('Video is required')
      return
    }

    try {
      const response = await sendMessage.mutateAsync({
        sessionId,
        data: { to: to.trim(), video_url: videoUrl.trim(), caption: caption.trim() || undefined },
      })

      toast.success('Video sent', `Job ID: ${response.job_id}`)
      setTo('')
      setVideoUrl('')
      setCaption('')
    } catch (err) {
      if (err instanceof ApiException) {
        setError(err.message)
      } else {
        setError('Error sending video')
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
        type="video"
        label="Video"
        value={videoUrl}
        onChange={setVideoUrl}
        placeholder="https://your-server.com/video.mp4"
        disabled={sendMessage.isPending}
      />

      <Input
        label="Caption (Optional)"
        type="text"
        placeholder="Video description..."
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
        Send Video
      </Button>
    </form>
  )
}
