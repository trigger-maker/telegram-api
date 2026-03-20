import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  Loader2,
  AlertCircle,
  Webhook,
  Play,
  Square,
  Trash2,
  Plus,
  Clock,
  Link as LinkIcon,
  Shield,
  RefreshCw,
  Zap,
  Radio,
  Settings,
  Eye,
  EyeOff,
} from 'lucide-react'
import { Layout } from '@/components/layout'
import { Button, Alert, Card, Input, Modal } from '@/components/common'
import { useSession, useWebhook } from '@/hooks'
import { useToast } from '@/contexts'
import { WEBHOOK_EVENTS } from '@/config/constants'
import { WebhookCreateRequest } from '@/api/webhooks.api'
import { SessionStatus } from '@/types'

const useWebhookHandlers = (webhook: ReturnType<typeof useWebhook>, toast: ReturnType<typeof useToast>) => {
  const handleCreateWebhook = async (formData: WebhookCreateRequest, autoStart: boolean) => {
    if (!formData.url) {
      toast.error('Error', 'URL is required')
      return
    }

    try {
      await webhook.create(formData, autoStart)
      toast.success(
        'Webhook created',
        autoStart
          ? 'The webhook has been created and is listening for events'
          : 'The webhook has been configured successfully'
      )
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Could not create webhook'
      toast.error('Error', message)
    }
  }

  const handleDeleteWebhook = async () => {
    try {
      await webhook.delete()
      toast.success('Webhook deleted', 'The webhook has been deleted')
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Could not delete webhook'
      toast.error('Error', message)
    }
  }

  const handleStartWebhook = async () => {
    try {
      await webhook.start()
      toast.success('Webhook started', 'The webhook is listening for events')
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Could not start webhook'
      toast.error('Error', message)
    }
  }

  const handleStopWebhook = async () => {
    try {
      await webhook.stop()
      toast.success('Webhook stopped', 'Webhook has stopped listening')
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Could not stop webhook'
      toast.error('Error', message)
    }
  }

  return { handleCreateWebhook, handleDeleteWebhook, handleStartWebhook, handleStopWebhook }
}

const ConnectionStatusIcon = ({ webhook }: { webhook: ReturnType<typeof useWebhook> }) => (
  <div
    className={`relative flex items-center justify-center w-10 h-10 rounded-full ${
      webhook.isListening
        ? 'bg-green-100 dark:bg-green-900/50'
        : 'bg-gray-100 dark:bg-gray-800'
    }`}
  >
    <Radio
      className={`w-5 h-5 ${
        webhook.isListening
          ? 'text-green-600 dark:text-green-400'
          : 'text-gray-400'
      }`}
    />
    {webhook.isListening && (
      <span className="absolute -top-0.5 -right-0.5 flex h-3 w-3">
        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
        <span className="relative inline-flex rounded-full h-3 w-3 bg-green-500"></span>
      </span>
    )}
  </div>
)

const ConnectionStatusText = ({ webhook }: { webhook: ReturnType<typeof useWebhook> }) => (
  <div>
    <p className="font-semibold text-gray-900 dark:text-white">
      {webhook.isListening ? 'Listening for events' : 'Webhook stopped'}
    </p>
    <p className="text-sm text-gray-600 dark:text-gray-400">
      {webhook.isListening
        ? `Connected since ${webhook.poolSession?.started_at ? new Date(webhook.poolSession.started_at).toLocaleTimeString('en-US') : 'now'}`
        : 'Start the webhook to receive events'}
    </p>
  </div>
)

const ConnectionStatusActions = ({
  webhook,
  handleStartWebhook,
  handleStopWebhook,
}: {
  webhook: ReturnType<typeof useWebhook>
  handleStartWebhook: () => void
  handleStopWebhook: () => void
}) => (
  <div className="flex items-center gap-2">
    {webhook.isListening ? (
      <Button
        variant="secondary"
        onClick={handleStopWebhook}
        isLoading={webhook.isStopping}
      >
        <Square className="w-4 h-4 mr-2" />
        Stop
      </Button>
    ) : (
      <Button
        variant="primary"
        onClick={handleStartWebhook}
        isLoading={webhook.isStarting}
      >
        <Play className="w-4 h-4 mr-2" />
        Start
      </Button>
    )}
  </div>
)

const ConnectionStatusBanner = ({
  webhook,
  handleStartWebhook,
  handleStopWebhook,
}: {
  webhook: ReturnType<typeof useWebhook>
  handleStartWebhook: () => void
  handleStopWebhook: () => void
}) => (
  <div
    className={`p-4 rounded-xl border-2 transition-all ${
      webhook.isListening
        ? 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800'
        : 'bg-gray-50 dark:bg-gray-800/50 border-gray-200 dark:border-gray-700'
    }`}
  >
    <div className="flex items-center justify-between">
      <div className="flex items-center gap-3">
        <ConnectionStatusIcon webhook={webhook} />
        <ConnectionStatusText webhook={webhook} />
      </div>

      <ConnectionStatusActions
        webhook={webhook}
        handleStartWebhook={handleStartWebhook}
        handleStopWebhook={handleStopWebhook}
      />
    </div>
  </div>
)

const WebhookConfigHeader = ({
  webhook,
  setShowDeleteConfirm,
}: {
  webhook: ReturnType<typeof useWebhook>
  setShowDeleteConfirm: (value: boolean) => void
}) => (
  <div className="flex items-start justify-between mb-6">
    <div className="flex items-center gap-4">
      <div
        className={`p-3 rounded-xl ${
          webhook.isListening
            ? 'bg-green-100 dark:bg-green-900/30'
            : 'bg-gray-100 dark:bg-gray-800'
        }`}
      >
        <Webhook
          className={`w-6 h-6 ${
            webhook.isListening
              ? 'text-green-600 dark:text-green-400'
              : 'text-gray-500'
          }`}
        />
      </div>
      <div>
        <h3 className="font-semibold text-gray-900 dark:text-white">
          Webhook Configuration
        </h3>
        <p className="text-sm text-gray-500 dark:text-gray-400">
          Created on {new Date(webhook.config!.created_at).toLocaleDateString('en-US')}
        </p>
      </div>
    </div>

    <Button
      variant="danger"
      onClick={() => setShowDeleteConfirm(true)}
      isLoading={webhook.isDeleting}
    >
      <Trash2 className="w-4 h-4" />
    </Button>
  </div>
)

const WebhookConfigUrl = ({ config }: { config: NonNullable<ReturnType<typeof useWebhook>['config']> }) => (
  <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl">
    <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 mb-1">
      <LinkIcon className="w-4 h-4" />
      Destination URL
    </div>
    <p className="font-mono text-sm text-gray-900 dark:text-white break-all">
      {config.url}
    </p>
  </div>
)

const WebhookConfigSecret = ({
  config,
  showSecret,
  setShowSecret,
}: {
  config: NonNullable<ReturnType<typeof useWebhook>['config']>
  showSecret: boolean
  setShowSecret: (value: boolean) => void
}) => (
  <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl">
    <div className="flex items-center justify-between mb-1">
      <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
        <Shield className="w-4 h-4" />
        Secret
      </div>
      {config.secret && (
        <button
          onClick={() => setShowSecret(!showSecret)}
          className="text-gray-400 hover:text-gray-600"
        >
          {showSecret ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
        </button>
      )}
    </div>
    <p className="font-mono text-sm text-gray-900 dark:text-white">
      {config.secret
        ? showSecret
          ? config.secret
          : '••••••••••••••••••'
        : 'Not configured'}
    </p>
  </div>
)

const WebhookConfigTimeout = ({ config }: { config: NonNullable<ReturnType<typeof useWebhook>['config']> }) => (
  <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl">
    <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 mb-1">
      <Clock className="w-4 h-4" />
      Timeout
    </div>
    <p className="text-sm text-gray-900 dark:text-white">
      {config.timeout_ms}ms ({(config.timeout_ms / 1000).toFixed(1)}s)
    </p>
  </div>
)

const WebhookConfigRetries = ({ config }: { config: NonNullable<ReturnType<typeof useWebhook>['config']> }) => (
  <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl">
    <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 mb-1">
      <RefreshCw className="w-4 h-4" />
      Retries
    </div>
    <p className="text-sm text-gray-900 dark:text-white">
      {config.max_retries} attempts
    </p>
  </div>
)

const WebhookConfigEvents = ({ config }: { config: NonNullable<ReturnType<typeof useWebhook>['config']> }) => (
  <div className="mt-6">
    <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
      Subscribed events ({config.events?.length || 0})
    </h4>
    {config.events && config.events.length > 0 ? (
      <div className="flex flex-wrap gap-2">
        {config.events.map((eventId) => {
          const event = WEBHOOK_EVENTS.find((e) => e.id === eventId)
          return (
            <span
              key={eventId}
              className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-primary-100 dark:bg-primary-900/30 text-primary-700 dark:text-primary-400 rounded-lg text-sm font-medium"
            >
              <Zap className="w-3 h-3" />
              {event?.label || eventId}
            </span>
          )
        })}
      </div>
    ) : (
      <p className="text-sm text-gray-500 dark:text-gray-400">
        No events configured - all event types will be received
      </p>
    )}
  </div>
)

const WebhookConfigError = ({ config }: { config: NonNullable<ReturnType<typeof useWebhook>['config']> }) => {
  if (!config.last_error) return null

  return (
    <div className="mt-6 p-4 bg-red-50 dark:bg-red-900/20 rounded-xl border border-red-200 dark:border-red-800">
      <div className="flex items-center gap-2 text-red-700 dark:text-red-400 text-sm font-medium mb-1">
        <AlertCircle className="w-4 h-4" />
        Last error
      </div>
      <p className="text-sm text-red-600 dark:text-red-300">{config.last_error}</p>
      {config.last_error_at && (
        <p className="text-xs text-red-500 dark:text-red-400 mt-1">
          {new Date(config.last_error_at).toLocaleString('en-US')}
        </p>
      )}
    </div>
  )
}

const WebhookConfigCard = ({
  webhook,
  showSecret,
  setShowSecret,
  setShowDeleteConfirm,
}: {
  webhook: ReturnType<typeof useWebhook>
  showSecret: boolean
  setShowSecret: (value: boolean) => void
  setShowDeleteConfirm: (value: boolean) => void
}) => {
  const config = webhook.config
  if (!config) return null

  return (
    <Card className="p-6">
      <WebhookConfigHeader webhook={webhook} setShowDeleteConfirm={setShowDeleteConfirm} />

      <div className="grid gap-4 md:grid-cols-2">
        <WebhookConfigUrl config={config} />
        <WebhookConfigSecret config={config} showSecret={showSecret} setShowSecret={setShowSecret} />
        <WebhookConfigTimeout config={config} />
        <WebhookConfigRetries config={config} />
      </div>

      <WebhookConfigEvents config={config} />

      <WebhookConfigError config={config} />
    </Card>
  )
}

const EmptyWebhookCard = ({ onShowCreateModal }: { onShowCreateModal: () => void }) => (
  <Card className="p-12 text-center">
    <div className="inline-flex items-center justify-center w-16 h-16 bg-gray-100 dark:bg-gray-800 rounded-full mb-4">
      <Webhook className="w-8 h-8 text-gray-400" />
    </div>
    <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
      No webhook configured
    </h3>
    <p className="text-gray-600 dark:text-gray-400 mb-6 max-w-sm mx-auto">
      Configure a webhook to receive Telegram events in real time
    </p>
    <Button variant="primary" onClick={onShowCreateModal}>
      <Plus className="w-4 h-4 mr-2" />
      Configure Webhook
    </Button>
  </Card>
)

const PoolSessionItem = ({
  poolSession,
  sessionId,
}: {
  poolSession: { session_id: string; session_name: string; telegram_id: number; is_connected: boolean }
  sessionId: string
}) => (
  <div
    className={`flex items-center justify-between p-3 rounded-lg ${
      poolSession.session_id === sessionId
        ? 'bg-primary-50 dark:bg-primary-900/20 border border-primary-200 dark:border-primary-800'
        : 'bg-gray-50 dark:bg-gray-800/50'
    }`}
  >
    <div className="flex items-center gap-3">
      <div
        className={`w-2 h-2 rounded-full ${
          poolSession.is_connected ? 'bg-green-500' : 'bg-gray-400'
        }`}
      />
      <div>
        <p className="font-medium text-gray-900 dark:text-white text-sm">
          {poolSession.session_name}
          {poolSession.session_id === sessionId && (
            <span className="ml-2 text-xs text-primary-600 dark:text-primary-400">
              (This session)
            </span>
          )}
        </p>
        <p className="text-xs text-gray-500 dark:text-gray-400">
          ID: {poolSession.telegram_id}
        </p>
      </div>
    </div>
    <span
      className={`text-xs px-2 py-1 rounded-full ${
        poolSession.is_connected
          ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400'
          : 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400'
      }`}
    >
      {poolSession.is_connected ? 'Connected' : 'Disconnected'}
    </span>
  </div>
)

const PoolStatusCard = ({ webhook, sessionId }: { webhook: ReturnType<typeof useWebhook>; sessionId: string }) => {
  if (!webhook.poolStatus || webhook.poolStatus.active_count === 0) return null

  return (
    <Card className="p-6">
      <div className="flex items-center gap-3 mb-4">
        <Settings className="w-5 h-5 text-gray-500" />
        <h3 className="font-semibold text-gray-900 dark:text-white">
          Active sessions pool ({webhook.poolStatus.active_count})
        </h3>
      </div>
      <div className="space-y-2">
        {webhook.poolStatus.sessions.map((poolSession) => (
          <PoolSessionItem key={poolSession.session_id} poolSession={poolSession} sessionId={sessionId} />
        ))}
      </div>
    </Card>
  )
}

const EventInfoItem = ({ event }: { event: { id: string; label: string; description: string } }) => (
  <div className="p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
    <div className="flex items-center gap-2 mb-1">
      <Zap className="w-4 h-4 text-primary-500" />
      <p className="font-medium text-gray-900 dark:text-white text-sm">{event.label}</p>
    </div>
    <p className="text-xs text-gray-500 dark:text-gray-400">{event.description}</p>
    <p className="text-xs font-mono text-gray-400 dark:text-gray-500 mt-1">
      {event.id}
    </p>
  </div>
)

const EventsInfoCard = () => (
  <Card className="p-6">
    <h3 className="font-semibold text-gray-900 dark:text-white mb-4">Available events</h3>
    <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
      {WEBHOOK_EVENTS.map((event) => (
        <EventInfoItem key={event.id} event={event} />
      ))}
    </div>
  </Card>
)

const EventSelectorItem = ({
  event,
  formData,
  toggleEvent,
}: {
  event: { id: string; label: string }
  formData: WebhookCreateRequest
  toggleEvent: (eventId: string) => void
}) => (
  <label
    className={`
      flex items-center gap-3 p-3 rounded-lg border-2 cursor-pointer transition-all
      ${
        formData.events?.includes(event.id)
          ? 'border-primary-600 bg-primary-50 dark:bg-primary-900/20'
          : 'border-gray-200 dark:border-gray-700 hover:border-primary-300'
      }
    `}
  >
    <input
      type="checkbox"
      checked={formData.events?.includes(event.id)}
      onChange={() => toggleEvent(event.id)}
      className="w-4 h-4 text-primary-600 rounded focus:ring-primary-500"
    />
    <div className="min-w-0">
      <p className="font-medium text-gray-900 dark:text-white text-sm">
        {event.label}
      </p>
      <p className="text-xs text-gray-500 dark:text-gray-400 truncate">
        {event.id}
      </p>
    </div>
  </label>
)

const EventSelectorHeader = ({
  formData,
  selectAllEvents,
  clearAllEvents,
}: {
  formData: WebhookCreateRequest
  selectAllEvents: () => void
  clearAllEvents: () => void
}) => (
  <div className="flex items-center justify-between mb-3">
    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
      Events to listen ({formData.events?.length || 0} selected)
    </label>
    <div className="flex gap-2">
      <button
        type="button"
        onClick={selectAllEvents}
        className="text-xs text-primary-600 hover:text-primary-700 dark:text-primary-400"
      >
        Select all
      </button>
      <span className="text-gray-300 dark:text-gray-600">|</span>
      <button
        type="button"
        onClick={clearAllEvents}
        className="text-xs text-gray-500 hover:text-gray-700 dark:text-gray-400"
      >
        Clear
      </button>
    </div>
  </div>
)

const EventSelectorList = ({
  formData,
  toggleEvent,
}: {
  formData: WebhookCreateRequest
  toggleEvent: (eventId: string) => void
}) => (
  <div className="grid gap-2 md:grid-cols-2 max-h-64 overflow-y-auto">
    {WEBHOOK_EVENTS.map((event) => (
      <EventSelectorItem
        key={event.id}
        event={event}
        formData={formData}
        toggleEvent={toggleEvent}
      />
    ))}
  </div>
)

const AutoStartToggle = ({
  autoStart,
  setAutoStart,
}: {
  autoStart: boolean
  setAutoStart: (value: boolean) => void
}) => (
  <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl">
    <label className="flex items-center gap-3 cursor-pointer">
      <input
        type="checkbox"
        checked={autoStart}
        onChange={(e) => setAutoStart(e.target.checked)}
        className="w-5 h-5 text-primary-600 rounded focus:ring-primary-500"
      />
      <div>
        <p className="font-medium text-gray-900 dark:text-white">
          Start automatically
        </p>
        <p className="text-sm text-gray-500 dark:text-gray-400">
          The webhook will start listening to events immediately after creation
        </p>
      </div>
    </label>
  </div>
)

const CreateWebhookFormUrl = ({
  formData,
  setFormData,
}: {
  formData: WebhookCreateRequest
  setFormData: (value: WebhookCreateRequest) => void
}) => (
  <>
    <Input
      label="Webhook URL *"
      type="url"
      placeholder="https://your-server.com/webhook"
      value={formData.url}
      onChange={(e) => setFormData({ ...formData, url: e.target.value })}
    />

    <Input
      label="Secret (optional)"
      type="text"
      placeholder="Secret key to sign requests"
      value={formData.secret}
      onChange={(e) => setFormData({ ...formData, secret: e.target.value })}
      helperText="Will be used to sign requests with HMAC-SHA256"
    />
  </>
)

const CreateWebhookFormSettings = ({
  formData,
  setFormData,
}: {
  formData: WebhookCreateRequest
  setFormData: (value: WebhookCreateRequest) => void
}) => (
  <div className="grid grid-cols-2 gap-4">
    <Input
      label="Timeout (ms)"
      type="number"
      value={formData.timeout_ms}
      onChange={(e) =>
        setFormData({ ...formData, timeout_ms: parseInt(e.target.value) || 5000 })
      }
      min={1000}
      max={30000}
    />
    <Input
      label="Max Retries"
      type="number"
      value={formData.max_retries}
      onChange={(e) =>
        setFormData({ ...formData, max_retries: parseInt(e.target.value) || 3 })
      }
      min={0}
      max={10}
    />
  </div>
)

const CreateWebhookFormEvents = ({
  formData,
  toggleEvent,
  selectAllEvents,
  clearAllEvents,
}: {
  formData: WebhookCreateRequest
  toggleEvent: (eventId: string) => void
  selectAllEvents: () => void
  clearAllEvents: () => void
}) => (
  <div>
    <EventSelectorHeader
      formData={formData}
      selectAllEvents={selectAllEvents}
      clearAllEvents={clearAllEvents}
    />
    <EventSelectorList formData={formData} toggleEvent={toggleEvent} />
    {(!formData.events || formData.events.length === 0) && (
      <p className="text-xs text-amber-600 dark:text-amber-400 mt-2">
        If you don't select events, all event types will be received
      </p>
    )}
  </div>
)

const CreateWebhookForm = ({
  formData,
  setFormData,
  toggleEvent,
  selectAllEvents,
  clearAllEvents,
  autoStart,
  setAutoStart,
}: {
  formData: WebhookCreateRequest
  setFormData: (value: WebhookCreateRequest) => void
  toggleEvent: (eventId: string) => void
  selectAllEvents: () => void
  clearAllEvents: () => void
  autoStart: boolean
  setAutoStart: (value: boolean) => void
}) => (
  <div className="p-6 space-y-6">
    <CreateWebhookFormUrl formData={formData} setFormData={setFormData} />
    <CreateWebhookFormSettings formData={formData} setFormData={setFormData} />
    <CreateWebhookFormEvents
      formData={formData}
      toggleEvent={toggleEvent}
      selectAllEvents={selectAllEvents}
      clearAllEvents={clearAllEvents}
    />
    <AutoStartToggle autoStart={autoStart} setAutoStart={setAutoStart} />
  </div>
)

const CreateWebhookModalButtons = ({
  onClose,
  handleCreateWebhook,
  webhook,
  autoStart,
}: {
  onClose: () => void
  handleCreateWebhook: () => void
  webhook: ReturnType<typeof useWebhook>
  autoStart: boolean
}) => (
  <div className="flex gap-3 pt-4 border-t border-gray-200 dark:border-gray-700 px-6 pb-6">
    <Button
      variant="secondary"
      onClick={onClose}
      fullWidth
    >
      Cancel
    </Button>
    <Button
      variant="primary"
      onClick={handleCreateWebhook}
      isLoading={webhook.isCreating || webhook.isStarting}
      fullWidth
    >
      <Webhook className="w-4 h-4 mr-2" />
      {autoStart ? 'Create and Start' : 'Create Webhook'}
    </Button>
  </div>
)

const useCreateWebhookEvents = (
  _formData: WebhookCreateRequest,
  setFormData: (value: WebhookCreateRequest | ((prev: WebhookCreateRequest) => WebhookCreateRequest)) => void
) => {
  const toggleEvent = (eventId: string) => {
    setFormData((prev: WebhookCreateRequest) => ({
      ...prev,
      events: prev.events?.includes(eventId)
        ? prev.events.filter((e: string) => e !== eventId)
        : [...(prev.events || []), eventId],
    }))
  }

  const selectAllEvents = () => {
    setFormData((prev: WebhookCreateRequest) => ({
      ...prev,
      events: WEBHOOK_EVENTS.map((e) => e.id),
    }))
  }

  const clearAllEvents = () => {
    setFormData((prev: WebhookCreateRequest) => ({
      ...prev,
      events: [],
    }))
  }

  return { toggleEvent, selectAllEvents, clearAllEvents }
}

const CreateWebhookModalContent = ({
  formData,
  setFormData,
  autoStart,
  setAutoStart,
  handleCreateWebhook,
  webhook,
  onClose,
}: {
  formData: WebhookCreateRequest
  setFormData: (value: WebhookCreateRequest | ((prev: WebhookCreateRequest) => WebhookCreateRequest)) => void
  autoStart: boolean
  setAutoStart: (value: boolean) => void
  handleCreateWebhook: () => void
  webhook: ReturnType<typeof useWebhook>
  onClose: () => void
}) => {
  const { toggleEvent, selectAllEvents, clearAllEvents } = useCreateWebhookEvents(formData, setFormData)

  return (
    <>
      <CreateWebhookForm
        formData={formData}
        setFormData={setFormData}
        toggleEvent={toggleEvent}
        selectAllEvents={selectAllEvents}
        clearAllEvents={clearAllEvents}
        autoStart={autoStart}
        setAutoStart={setAutoStart}
      />
      <CreateWebhookModalButtons
        onClose={onClose}
        handleCreateWebhook={handleCreateWebhook}
        webhook={webhook}
        autoStart={autoStart}
      />
    </>
  )
}

const CreateWebhookModal = ({
  isOpen,
  onClose,
  formData,
  setFormData,
  autoStart,
  setAutoStart,
  handleCreateWebhook,
  webhook,
}: {
  isOpen: boolean
  onClose: () => void
  formData: WebhookCreateRequest
  setFormData: (value: WebhookCreateRequest | ((prev: WebhookCreateRequest) => WebhookCreateRequest)) => void
  autoStart: boolean
  setAutoStart: (value: boolean) => void
  handleCreateWebhook: () => void
  webhook: ReturnType<typeof useWebhook>
}) => (
  <Modal
    isOpen={isOpen}
    onClose={onClose}
    title="Configure Webhook"
    size="lg"
  >
    <CreateWebhookModalContent
      formData={formData}
      setFormData={setFormData}
      autoStart={autoStart}
      setAutoStart={setAutoStart}
      handleCreateWebhook={handleCreateWebhook}
      webhook={webhook}
      onClose={onClose}
    />
  </Modal>
)

const DeleteConfirmModalContent = ({
  handleDeleteWebhook,
  webhook,
}: {
  handleDeleteWebhook: () => void
  webhook: ReturnType<typeof useWebhook>
}) => (
  <div className="p-6">
    <div className="flex items-center gap-4 mb-4">
      <div className="p-3 bg-red-100 dark:bg-red-900/30 rounded-full">
        <Trash2 className="w-6 h-6 text-red-600 dark:text-red-400" />
      </div>
      <div>
        <p className="font-medium text-gray-900 dark:text-white">
          Are you sure you want to delete the webhook?
        </p>
        <p className="text-sm text-gray-500 dark:text-gray-400">
          It will stop listening to events and the configuration will be deleted
        </p>
      </div>
    </div>

    <div className="flex gap-3">
      <Button
        variant="secondary"
        onClick={() => handleDeleteWebhook()}
        fullWidth
      >
        Cancel
      </Button>
      <Button
        variant="danger"
        onClick={() => handleDeleteWebhook()}
        isLoading={webhook.isDeleting}
        fullWidth
      >
        Delete
      </Button>
    </div>
  </div>
)

const DeleteConfirmModal = ({
  isOpen,
  onClose,
  handleDeleteWebhook,
  webhook,
}: {
  isOpen: boolean
  onClose: () => void
  handleDeleteWebhook: () => void
  webhook: ReturnType<typeof useWebhook>
}) => (
  <Modal
    isOpen={isOpen}
    onClose={onClose}
    title="Delete Webhook"
    size="sm"
  >
    <DeleteConfirmModalContent
      handleDeleteWebhook={handleDeleteWebhook}
      webhook={webhook}
    />
  </Modal>
)

const WebhooksPageHeader = ({
  session,
  webhook,
  config,
  navigate,
}: {
  session: SessionStatus
  webhook: ReturnType<typeof useWebhook>
  config: ReturnType<typeof useWebhook>['config']
  navigate: (path: string) => void
}) => (
  <div className="flex items-center justify-between">
    <div className="flex items-center gap-4">
      <Button variant="ghost" onClick={() => navigate('/dashboard')}>
        <ArrowLeft className="w-4 h-4" />
      </Button>
      <div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Webhooks</h1>
        <p className="text-gray-600 dark:text-gray-400 mt-1">
          {session.session.session_name} - Receive events in real time
        </p>
      </div>
    </div>

    <div className="flex items-center gap-2">
      <Button variant="ghost" onClick={() => webhook.refetch()} disabled={webhook.isActing}>
        <RefreshCw className={`w-4 h-4 ${webhook.isLoading ? 'animate-spin' : ''}`} />
      </Button>
      {!config && (
        <Button variant="primary" onClick={() => navigate('/dashboard')}>
          <Plus className="w-4 h-4 mr-2" />
          Configure Webhook
        </Button>
      )}
    </div>
  </div>
)

const WebhooksPageContent = ({
  webhook,
  config,
  sessionId,
  showSecret,
  setShowSecret,
  setShowDeleteConfirm,
  setShowCreateModal,
}: {
  webhook: ReturnType<typeof useWebhook>
  config: ReturnType<typeof useWebhook>['config']
  sessionId: string
  showSecret: boolean
  setShowSecret: (value: boolean) => void
  setShowDeleteConfirm: (value: boolean) => void
  setShowCreateModal: (value: boolean) => void
}) => (
  <>
    {config && (
      <ConnectionStatusBanner
        webhook={webhook}
        handleStartWebhook={webhook.start}
        handleStopWebhook={webhook.stop}
      />
    )}

    {config ? (
      <WebhookConfigCard
        webhook={webhook}
        showSecret={showSecret}
        setShowSecret={setShowSecret}
        setShowDeleteConfirm={setShowDeleteConfirm}
      />
    ) : (
      <EmptyWebhookCard onShowCreateModal={() => setShowCreateModal(true)} />
    )}

    <PoolStatusCard webhook={webhook} sessionId={sessionId} />

    <EventsInfoCard />
  </>
)

const WebhooksPageLoading = () => (
  <Layout>
    <div className="flex items-center justify-center py-12">
      <Loader2 className="w-8 h-8 animate-spin text-primary-600" />
    </div>
  </Layout>
)

const WebhooksPageInvalidSession = () => (
  <Layout>
    <Alert variant="error">Invalid session ID</Alert>
  </Layout>
)

const WebhooksPageNotFound = () => (
  <Layout>
    <Alert variant="error">Session not found</Alert>
  </Layout>
)

const WebhooksPageMainContent = ({
  session,
  webhook,
  config,
  sessionId,
  showSecret,
  setShowSecret,
  setShowDeleteConfirm,
  setShowCreateModal,
  navigate,
}: {
  session: SessionStatus
  webhook: ReturnType<typeof useWebhook>
  config: ReturnType<typeof useWebhook>['config']
  sessionId: string
  showSecret: boolean
  setShowSecret: (value: boolean) => void
  setShowDeleteConfirm: (value: boolean) => void
  setShowCreateModal: (value: boolean) => void
  navigate: (path: string) => void
}) => (
  <>
    <WebhooksPageHeader
      session={session}
      webhook={webhook}
      config={config}
      navigate={navigate}
    />

    <WebhooksPageContent
      webhook={webhook}
      config={config}
      sessionId={sessionId}
      showSecret={showSecret}
      setShowSecret={setShowSecret}
      setShowDeleteConfirm={setShowDeleteConfirm}
      setShowCreateModal={setShowCreateModal}
    />
  </>
)

const WebhooksPageMainModals = ({
  showCreateModal,
  setShowCreateModal,
  resetForm,
  formData,
  setFormData,
  autoStart,
  setAutoStart,
  handleCreateWebhook,
  webhook,
  showDeleteConfirm,
  setShowDeleteConfirm,
  handlers,
}: {
  showCreateModal: boolean
  setShowCreateModal: (value: boolean) => void
  resetForm: () => void
  formData: WebhookCreateRequest
  setFormData: (value: WebhookCreateRequest | ((prev: WebhookCreateRequest) => WebhookCreateRequest)) => void
  autoStart: boolean
  setAutoStart: (value: boolean) => void
  handleCreateWebhook: () => void
  webhook: ReturnType<typeof useWebhook>
  showDeleteConfirm: boolean
  setShowDeleteConfirm: (value: boolean) => void
  handlers: ReturnType<typeof useWebhookHandlers>
}) => (
  <>
    <CreateWebhookModal
      isOpen={showCreateModal}
      onClose={() => {
        setShowCreateModal(false)
        resetForm()
      }}
      formData={formData}
      setFormData={setFormData}
      autoStart={autoStart}
      setAutoStart={setAutoStart}
      handleCreateWebhook={handleCreateWebhook}
      webhook={webhook}
    />

    <DeleteConfirmModal
      isOpen={showDeleteConfirm}
      onClose={() => setShowDeleteConfirm(false)}
      handleDeleteWebhook={handlers.handleDeleteWebhook}
      webhook={webhook}
    />
  </>
)

interface WebhooksPageMainProps {
  session: SessionStatus
  webhook: ReturnType<typeof useWebhook>
  config: ReturnType<typeof useWebhook>['config']
  sessionId: string
  showSecret: boolean
  setShowSecret: (value: boolean) => void
  setShowDeleteConfirm: (value: boolean) => void
  setShowCreateModal: (value: boolean) => void
  showCreateModal: boolean
  resetForm: () => void
  formData: WebhookCreateRequest
  setFormData: (value: WebhookCreateRequest | ((prev: WebhookCreateRequest) => WebhookCreateRequest)) => void
  autoStart: boolean
  setAutoStart: (value: boolean) => void
  handleCreateWebhook: () => void
  showDeleteConfirm: boolean
  handlers: ReturnType<typeof useWebhookHandlers>
  navigate: (path: string) => void
}

const WebhooksPageMain = ({
  session,
  webhook,
  config,
  sessionId,
  showSecret,
  setShowSecret,
  setShowDeleteConfirm,
  setShowCreateModal,
  showCreateModal,
  resetForm,
  formData,
  setFormData,
  autoStart,
  setAutoStart,
  handleCreateWebhook,
  showDeleteConfirm,
  handlers,
  navigate,
}: WebhooksPageMainProps) => (
  <Layout>
    <div className="max-w-4xl mx-auto space-y-6">
      <WebhooksPageMainContent
        session={session}
        webhook={webhook}
        config={config}
        sessionId={sessionId}
        showSecret={showSecret}
        setShowSecret={setShowSecret}
        setShowDeleteConfirm={setShowDeleteConfirm}
        setShowCreateModal={setShowCreateModal}
        navigate={navigate}
      />

      <WebhooksPageMainModals
        showCreateModal={showCreateModal}
        setShowCreateModal={setShowCreateModal}
        resetForm={resetForm}
        formData={formData}
        setFormData={setFormData}
        autoStart={autoStart}
        setAutoStart={setAutoStart}
        handleCreateWebhook={handleCreateWebhook}
        webhook={webhook}
        showDeleteConfirm={showDeleteConfirm}
        setShowDeleteConfirm={setShowDeleteConfirm}
        handlers={handlers}
      />
    </div>
  </Layout>
)

const useWebhooksPageState = () => {
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [showSecret, setShowSecret] = useState(false)
  const [autoStart, setAutoStart] = useState(true)

  const [formData, setFormData] = useState<WebhookCreateRequest>({
    url: '',
    events: ['message.new'],
    secret: '',
    timeout_ms: 5000,
    max_retries: 3,
  })

  return {
    showCreateModal,
    setShowCreateModal,
    showDeleteConfirm,
    setShowDeleteConfirm,
    showSecret,
    setShowSecret,
    autoStart,
    setAutoStart,
    formData,
    setFormData,
  }
}

const useWebhooksPageHandlers = (webhook: ReturnType<typeof useWebhook>, toast: ReturnType<typeof useToast>) => {
  const handlers = useWebhookHandlers(webhook, toast)

  const resetForm = (setFormData: (value: WebhookCreateRequest) => void, setAutoStart: (value: boolean) => void) => {
    setFormData({
      url: '',
      events: ['message.new'],
      secret: '',
      timeout_ms: 5000,
      max_retries: 3,
    })
    setAutoStart(true)
  }

  const handleCreateWebhook = (
    formData: WebhookCreateRequest,
    autoStart: boolean,
    setShowCreateModal: (value: boolean) => void,
    resetForm: () => void
  ) => {
    handlers.handleCreateWebhook(formData, autoStart)
    setShowCreateModal(false)
    resetForm()
  }

  return { handlers, resetForm, handleCreateWebhook }
}

const onCreateWebhook = (
  state: ReturnType<typeof useWebhooksPageState>,
  handleCreateWebhook: ReturnType<typeof useWebhooksPageHandlers>['handleCreateWebhook'],
  resetForm: ReturnType<typeof useWebhooksPageHandlers>['resetForm']
) => () => {
  handleCreateWebhook(
    state.formData,
    state.autoStart,
    state.setShowCreateModal,
    () => resetForm(state.setFormData, state.setAutoStart)
  )
}

export const WebhooksPage = () => {
  const { sessionId } = useParams<{ sessionId: string }>()
  const navigate = useNavigate()
  const toast = useToast()

  const { data: session, isLoading: sessionLoading } = useSession(sessionId!)
  const webhook = useWebhook(sessionId!)

  const state = useWebhooksPageState()
  const { handlers, resetForm, handleCreateWebhook } = useWebhooksPageHandlers(webhook, toast)

  if (!sessionId) {
    return <WebhooksPageInvalidSession />
  }

  if (sessionLoading || webhook.isLoading) {
    return <WebhooksPageLoading />
  }

  if (!session) {
    return <WebhooksPageNotFound />
  }

  const config = webhook.config

  return (
    <WebhooksPageMain
      session={session}
      webhook={webhook}
      config={config}
      sessionId={sessionId}
      showSecret={state.showSecret}
      setShowSecret={state.setShowSecret}
      setShowDeleteConfirm={state.setShowDeleteConfirm}
      setShowCreateModal={state.setShowCreateModal}
      showCreateModal={state.showCreateModal}
      resetForm={() => resetForm(state.setFormData, state.setAutoStart)}
      formData={state.formData}
      setFormData={state.setFormData}
      autoStart={state.autoStart}
      setAutoStart={state.setAutoStart}
      handleCreateWebhook={onCreateWebhook(state, handleCreateWebhook, resetForm)}
      showDeleteConfirm={state.showDeleteConfirm}
      handlers={handlers}
      navigate={navigate}
    />
  )
}
