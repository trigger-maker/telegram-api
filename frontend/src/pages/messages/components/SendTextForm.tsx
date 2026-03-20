import { useState } from 'react'
import { Button, Input, Alert } from '@/components/common'
import { useSendTextMessage } from '@/hooks'
import { useToast } from '@/contexts'
import { ApiException } from '@/types'
import { Send } from 'lucide-react'

interface SendTextFormProps {
  sessionId: string
}

// Tailwind classes for textarea
const TEXTAREA_CLASSES = 'w-full px-4 py-3 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all resize-none'

/* eslint-disable max-lines-per-function */
export const SendTextForm = ({ sessionId }: SendTextFormProps) => {
  const toast = useToast()
  const [to, setTo] = useState('')
  const [text, setText] = useState('')
  const [error, setError] = useState('')

  const sendMessage = useSendTextMessage()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!to.trim()) {
      setError('Recipient is required')
      return
    }

    if (!text.trim()) {
      setError('Message is required')
      return
    }

    try {
      const response = await sendMessage.mutateAsync({
        sessionId,
        data: { to: to.trim(), text: text.trim() },
      })

      toast.success('Message sent', `Job ID: ${response.job_id}`)
      setTo('')
      setText('')
    } catch (err) {
      if (err instanceof ApiException) {
        setError(err.message)
      } else {
        setError('Error sending message')
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

      <div>
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
          Message
        </label>
        <textarea
          className={TEXTAREA_CLASSES}
          rows={6}
          placeholder="Type your message here..."
          value={text}
          onChange={(e) => setText(e.target.value)}
          disabled={sendMessage.isPending}
        />
        <p className="text-xs text-gray-500 dark:text-gray-400 mt-2">
          {text.length} characters
        </p>
      </div>

      <Button
        type="submit"
        variant="primary"
        isLoading={sendMessage.isPending}
        fullWidth
        className="h-12"
      >
        <Send className="w-4 h-4 mr-2" />
        Send Message
      </Button>
    </form>
  )
}
