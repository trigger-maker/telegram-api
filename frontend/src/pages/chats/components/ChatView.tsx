/* eslint-disable max-lines-per-function */
import { useEffect, useRef, useState, useCallback } from 'react'
import {
  Loader2,
  AlertCircle,
  Image as ImageIcon,
  Video,
  Music,
  FileText,
  CheckCheck,
  Check,
  RefreshCw,
} from 'lucide-react'
import { useChatHistory, useChatInfo } from '@/hooks'
import { Alert, Card, Button } from '@/components/common'
import { ChatMessage } from '@/api/chats.api'
import { MessageInput } from './MessageInput'

interface ChatViewProps {
  sessionId: string
  chatId: number
}

// Tailwind class constants
const messageContainerClasses = 'max-w-[85%] sm:max-w-[70%] rounded-2xl px-3 sm:px-4 py-2 shadow-sm'
const footerInfoClasses =
  'border-t border-gray-200 dark:border-gray-700 px-3 py-1.5 ' +
  'text-xs text-gray-500 dark:text-gray-500 text-center bg-gray-50 dark:bg-gray-900/50'
const outgoingMessageClasses = 'bg-primary-600 text-white rounded-br-md'
const incomingMessageClasses = 'bg-white dark:bg-gray-800 text-gray-900 dark:text-white rounded-bl-md'
const mediaContainerClasses = 'flex items-center gap-2 mb-2 text-sm'
const outgoingMediaTextClasses = 'text-white/90'
const incomingMediaTextClasses = 'text-gray-600 dark:text-gray-400'
const timeStatusContainerClasses = 'flex items-center justify-end gap-1 mt-1 text-[10px] sm:text-xs'
const outgoingTimeTextClasses = 'text-white/70'
const incomingTimeTextClasses = 'text-gray-500 dark:text-gray-500'

const getMediaIcon = (mediaType?: string) => {
  if (!mediaType) return null
  const iconClass = 'w-4 h-4'
  const type = mediaType.toLowerCase()
  
  const iconMap: Record<string, React.ReactNode> = {
    photo: <ImageIcon className={iconClass} />,
    video: <Video className={iconClass} />,
    audio: <Music className={iconClass} />,
    document: <FileText className={iconClass} />,
    file: <FileText className={iconClass} />,
  }
  
  return iconMap[type] || null
}

const getMediaLabel = (mediaType?: string): string => {
  if (!mediaType) return ''
  const type = mediaType.toLowerCase()
  
  const labelMap: Record<string, string> = {
    photo: 'Photo',
    video: 'Video',
    audio: 'Audio',
    document: 'Document',
    file: 'Document',
  }
  
  return labelMap[type] || mediaType
}

const formatMessageTime = (dateStr: string): string => {
  const date = new Date(dateStr)
  return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })
}

const formatMessageDate = (dateStr: string): string => {
  const date = new Date(dateStr)
  const today = new Date()
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)

  const dateStrObj = date.toDateString()
  const todayStr = today.toDateString()
  const yesterdayStr = yesterday.toDateString()

  if (dateStrObj === todayStr) return 'Today'
  if (dateStrObj === yesterdayStr) return 'Yesterday'
  return date.toLocaleDateString('en-US', { day: '2-digit', month: 'long', year: 'numeric' })
}

const groupMessagesByDate = (messages: ChatMessage[]) => {
  const sortedMessages = [...messages].reverse()
  const groups: { [key: string]: ChatMessage[] } = {}
  const dateOrder: string[] = []

  sortedMessages.forEach((msg) => {
    const dateKey = new Date(msg.date).toDateString()
    if (!groups[dateKey]) {
      groups[dateKey] = []
      dateOrder.push(dateKey)
    }
    groups[dateKey].push(msg)
  })

  return dateOrder.map((dateKey) => ({
    date: groups[dateKey][0].date,
    messages: groups[dateKey],
  }))
}

// Helper component for message sender
const MessageSender = ({ message }: { message: ChatMessage }) => {
  if (message.is_outgoing || !message.from_name) return null
  return (
    <div className="text-xs font-semibold mb-1 text-primary-600 dark:text-primary-400">
      {message.from_name}
    </div>
  )
}

// Helper component for message forward info
const MessageForwardInfo = ({ message }: { message: ChatMessage }) => {
  if (!message.forward_from) return null
  return (
    <div className="text-xs mb-1 opacity-75 italic">
      Forwarded from: {message.forward_from}
    </div>
  )
}

// Helper component for message media
const MessageMedia = ({ message, mediaTextClasses }: { message: ChatMessage; mediaTextClasses: string }) => {
  if (!message.media_type) return null
  return (
    <div className={`${mediaContainerClasses} ${mediaTextClasses}`}>
      {getMediaIcon(message.media_type)}
      <span>{getMediaLabel(message.media_type)}</span>
    </div>
  )
}

// Helper component for message text
const MessageText = ({ text }: { text: string }) => (
  <div className="whitespace-pre-wrap break-words text-sm sm:text-base">
    {text}
  </div>
)

// Helper component for message time and status
const MessageTimeStatus = ({ message, timeTextClasses }: { message: ChatMessage; timeTextClasses: string }) => (
  <div className={`${timeStatusContainerClasses} ${timeTextClasses}`}>
    <span>{formatMessageTime(message.date)}</span>
    {message.is_outgoing && (
      <>
        {message.is_read ? <CheckCheck className="w-3 h-3" /> : <Check className="w-3 h-3" />}
      </>
    )}
  </div>
)

// Helper component for message bubble
const MessageBubble = ({ message }: { message: ChatMessage }) => {
  const isOutgoing = message.is_outgoing
  const containerClasses = `${messageContainerClasses} ${
    isOutgoing ? outgoingMessageClasses : incomingMessageClasses
  }`
  const mediaTextClasses = isOutgoing ? outgoingMediaTextClasses : incomingMediaTextClasses
  const timeTextClasses = isOutgoing ? outgoingTimeTextClasses : incomingTimeTextClasses

  return (
    <div className={containerClasses}>
      <MessageSender message={message} />
      <MessageForwardInfo message={message} />
      <MessageMedia message={message} mediaTextClasses={mediaTextClasses} />
      {message.text && <MessageText text={message.text} />}
      <MessageTimeStatus message={message} timeTextClasses={timeTextClasses} />
    </div>
  )
}

// Helper component for date separator
const DateSeparator = ({ date }: { date: string }) => (
  <div className="flex items-center justify-center my-3 sm:my-4">
    <div className="px-3 py-1 bg-gray-200 dark:bg-gray-700 rounded-full text-xs text-gray-600 dark:text-gray-400">
      {formatMessageDate(date)}
    </div>
  </div>
)

// Helper component for message group
const MessageGroup = ({ group }: { group: { date: string; messages: ChatMessage[] } }) => (
  <div>
    <DateSeparator date={group.date} />
    {group.messages.map((message) => (
      <div
        key={message.id}
        className={`flex mb-2 sm:mb-3 ${message.is_outgoing ? 'justify-end' : 'justify-start'}`}
      >
        <MessageBubble message={message} />
      </div>
    ))}
  </div>
)

// Helper component for chat header
const ChatHeader = ({
  chatTitle,
  username,
  isFetching,
  onRefresh,
}: {
  chatTitle: string
  username?: string
  isFetching: boolean
  onRefresh: () => void
}) => (
  <div className="border-b border-gray-200 dark:border-gray-700 p-3 sm:p-4 shrink-0 flex items-center justify-between">
    <div className="min-w-0 flex items-center gap-2">
      <div>
        <h3 className="font-semibold text-gray-900 dark:text-white truncate">
          {chatTitle}
        </h3>
        {username && <p className="text-sm text-gray-500 dark:text-gray-500">@{username}</p>}
      </div>
      <div className="flex items-center gap-1.5 px-2 py-0.5 bg-green-100 dark:bg-green-900/30 rounded-full">
        <span className="relative flex h-2 w-2">
          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
          <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
        </span>
        <span className="text-xs text-green-700 dark:text-green-400 font-medium">Sync</span>
      </div>
    </div>
    <Button variant="ghost" onClick={onRefresh} disabled={isFetching} className="shrink-0">
      <RefreshCw className={`w-4 h-4 ${isFetching ? 'animate-spin' : ''}`} />
    </Button>
  </div>
)

const ChatViewLoading = () => (
  <Card className="flex flex-col items-center justify-center py-12 gap-3">
    <Loader2 className="w-8 h-8 animate-spin text-primary-600" />
    <p className="text-sm text-gray-500 dark:text-gray-400">Loading messages...</p>
  </Card>
)

const ChatViewError = () => (
  <Alert variant="error">
    <div className="flex items-center gap-2">
      <AlertCircle className="w-5 h-5" />
      <span>Error loading chat history</span>
    </div>
  </Alert>
)

const ChatViewEmpty = () => (
  <div className="flex items-center justify-center h-full">
    <div className="text-center text-gray-500 dark:text-gray-400">
      <AlertCircle className="w-12 h-12 mx-auto mb-3 opacity-50" />
      <p>No messages in this chat</p>
      <p className="text-sm mt-1">Send the first message</p>
    </div>
  </div>
)

const ChatViewFooterInfo = ({
  messagesCount,
  hasMore,
}: {
  messagesCount: number
  hasMore?: boolean
}) => {
  if (messagesCount === 0) return null

  return (
    <div className={footerInfoClasses}>
      {messagesCount} message{messagesCount !== 1 ? 's' : ''}
      {hasMore && ' • There are more previous messages'}
    </div>
  )
}

const ChatViewContent = ({
  messagesContainerRef,
  onScroll,
  messages,
  messageGroups,
  messagesEndRef,
}: {
  messagesContainerRef: React.RefObject<HTMLDivElement | null>
  onScroll: () => void
  messages: ChatMessage[]
  messageGroups: ReturnType<typeof groupMessagesByDate>
  messagesEndRef: React.RefObject<HTMLDivElement | null>
}) => (
  <div
    ref={messagesContainerRef}
    onScroll={onScroll}
    className="flex-1 overflow-y-auto p-3 sm:p-4 space-y-4 bg-gray-50 dark:bg-gray-900/50"
  >
    {messages.length === 0 ? (
      <ChatViewEmpty />
    ) : (
      <>
        {messageGroups.map((group, idx) => (
          <MessageGroup key={idx} group={group} />
        ))}
        <div ref={messagesEndRef} />
      </>
    )}
  </div>
)

const getFullName = (firstName?: string, lastName?: string): string => {
  const first = firstName || ''
  const last = lastName || ''
  const fullName = `${first} ${last}`.trim()
  return fullName || 'Chat'
}

const getChatTitle = (chatInfo?: { title?: string; first_name?: string; last_name?: string }): string => {
  if (chatInfo?.title) return chatInfo.title
  return getFullName(chatInfo?.first_name, chatInfo?.last_name)
}

export const ChatView = ({ sessionId, chatId }: ChatViewProps) => {
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const [lastMessageCount, setLastMessageCount] = useState(0)
  const [isUserScrolledUp, setIsUserScrolledUp] = useState(false)
  const messagesContainerRef = useRef<HTMLDivElement>(null)

  const { data: chatInfo } = useChatInfo(sessionId, chatId)
  const { data: historyData, isLoading, error, refetch, isFetching } = useChatHistory(sessionId, chatId, {
    limit: 50,
    enablePolling: true,
    pollingInterval: 4000,
  })

  const messages = historyData?.messages ?? []
  const messageGroups = messages.length > 0 ? groupMessagesByDate(messages) : []

  const scrollToBottom = (behavior: ScrollBehavior = 'smooth') => {
    messagesEndRef.current?.scrollIntoView({ behavior })
  }

  const handleScroll = () => {
    if (!messagesContainerRef.current) return
    const { scrollTop, scrollHeight, clientHeight } = messagesContainerRef.current
    const isAtBottom = scrollHeight - scrollTop - clientHeight < 100
    setIsUserScrolledUp(!isAtBottom)
  }

  const shouldAutoScroll = useCallback((currentCount: number): boolean => {
    return currentCount > lastMessageCount && (!isUserScrolledUp || lastMessageCount === 0)
  }, [lastMessageCount, isUserScrolledUp])

  useEffect(() => {
    const currentCount = messages.length
    if (currentCount !== lastMessageCount) {
      if (shouldAutoScroll(currentCount)) {
        const scrollBehavior: ScrollBehavior = lastMessageCount === 0 ? 'instant' : 'smooth'
        scrollToBottom(scrollBehavior)
      }
      setLastMessageCount(currentCount)
    }
  }, [messages.length, lastMessageCount, shouldAutoScroll])

  const handleMessageSent = () => {
    setIsUserScrolledUp(false)
    refetch()
    setTimeout(() => {
      refetch()
      scrollToBottom()
    }, 1000)
  }

  if (isLoading) {
    return <ChatViewLoading />
  }

  if (error) {
    return <ChatViewError />
  }

  const chatTitle = getChatTitle(chatInfo)

  const ChatViewBody = () => (
    <>
      <ChatHeader
        chatTitle={chatTitle}
        username={chatInfo?.username}
        isFetching={isFetching}
        onRefresh={() => refetch()}
      />

      <ChatViewContent
        messagesContainerRef={messagesContainerRef}
        onScroll={handleScroll}
        messages={messages}
        messageGroups={messageGroups}
        messagesEndRef={messagesEndRef}
      />

      <ChatViewFooterInfo
        messagesCount={messages.length}
        hasMore={historyData?.has_more}
      />

      <MessageInput
        sessionId={sessionId}
        chatId={chatId}
        onMessageSent={handleMessageSent}
      />
    </>
  )

  return (
    <Card className="flex flex-col h-[calc(100vh-200px)] sm:h-[650px] overflow-hidden p-0">
      <ChatViewBody />
    </Card>
  )
}
