 
import { useState, useMemo } from 'react'
import { Users, User, Hash, Volume2, Archive, Pin, Search, X, MessageSquare } from 'lucide-react'
import { Chat, ChatType } from '@/api/chats.api'
import { Card } from '@/components/common'

interface ChatListProps {
  chats: Chat[]
  selectedChatId: number | null
  onSelectChat: (chatId: number) => void
  totalCount?: number
  hasMore?: boolean
}

// Tailwind classes for search input
const SEARCH_INPUT_CLASSES = 'w-full pl-9 pr-9 py-2 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm text-gray-900 dark:text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-shadow'

const getChatIcon = (type: ChatType) => {
  switch (type) {
    case 'private':
      return User
    case 'group':
    case 'supergroup':
      return Users
    case 'channel':
      return Hash
    default:
      return User
  }
}

const getChatTitle = (chat: Chat): string => {
  if (chat.title) return chat.title
  const firstName = chat.first_name || ''
  const lastName = chat.last_name || ''
  return `${firstName} ${lastName}`.trim() || 'No name'
}

const isDateValid = (date: Date): boolean => date.getFullYear() >= 2000

const getTimeDiff = (date: Date): { minutes: number; hours: number; days: number } => {
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  return {
    minutes: Math.floor(diff / (1000 * 60)),
    hours: Math.floor(diff / (1000 * 60 * 60)),
    days: Math.floor(diff / (1000 * 60 * 60 * 24)),
  }
}

const formatRelativeTime = (minutes: number, hours: number, days: number, date: Date): string => {
  if (minutes < 1) return 'Now'
  if (minutes < 60) return `${minutes}m`
  if (hours < 24) return `${hours}h`
  if (days === 1) return 'Yesterday'
  if (days < 7) return `${days}d`
  return date.toLocaleDateString('en-US', { day: '2-digit', month: '2-digit' })
}

const formatLastMessageTime = (dateStr?: string): string => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  if (!isDateValid(date)) return ''
  const { minutes, hours, days } = getTimeDiff(date)

  return formatRelativeTime(minutes, hours, days, date)
}

const getChatSearchValues = (chat: Chat): string[] => {
  const title = getChatTitle(chat).toLowerCase()
  const username = chat.username?.toLowerCase() || ''
  const lastMessage = chat.last_message?.toLowerCase() || ''
  return [title, username, lastMessage]
}

const matchesSearchTerm = (chat: Chat, term: string): boolean => {
  const values = getChatSearchValues(chat)
  return values.some((value) => value.includes(term))
}

// Helper component for chat avatar
const ChatAvatar = ({
  chat,
  isSelected,
}: {
  chat: Chat
  isSelected: boolean
}) => {
  const ChatIcon = getChatIcon(chat.type)
  return (
    <div
      className={`shrink-0 w-10 h-10 rounded-full flex items-center justify-center ${
        isSelected ? 'bg-primary-100 dark:bg-primary-900/30' : 'bg-gray-100 dark:bg-gray-800'
      }`}
    >
      <ChatIcon
        className={`w-5 h-5 ${
          isSelected ? 'text-primary-600 dark:text-primary-400' : 'text-gray-600 dark:text-gray-400'
        }`}
      />
    </div>
  )
}

// Helper component for chat title
const ChatTitle = ({ chat }: { chat: Chat }) => (
  <div className="flex items-center gap-1.5">
    <h3 className="font-semibold text-gray-900 dark:text-white truncate text-sm">
      {getChatTitle(chat)}
    </h3>
    {chat.is_pinned && <Pin className="w-3 h-3 text-primary-600 dark:text-primary-400 shrink-0" />}
  </div>
)

// Helper component for chat username
const ChatUsername = ({ username }: { username?: string }) => {
  if (!username) return null
  return (
    <p className="text-xs text-gray-500 dark:text-gray-500 truncate">@{username}</p>
  )
}

// Helper component for chat time and unread count
const ChatTimeInfo = ({
  time,
  unreadCount,
}: {
  time: string
  unreadCount: number
}) => (
  <div className="flex flex-col items-end gap-1 shrink-0">
    {time && <span className="text-xs text-gray-500 dark:text-gray-500">{time}</span>}
    {unreadCount > 0 && (
      <span className="inline-flex items-center justify-center px-1.5 py-0.5 text-xs font-medium bg-primary-600 text-white rounded-full min-w-[18px]">
        {unreadCount > 99 ? '99+' : unreadCount}
      </span>
    )}
  </div>
)

// Helper component for chat header
const ChatHeader = ({
  chat,
  time,
}: {
  chat: Chat
  time: string
}) => (
  <div className="flex items-start justify-between gap-2">
    <div className="flex-1 min-w-0">
      <ChatTitle chat={chat} />
      <ChatUsername username={chat.username} />
    </div>
    <ChatTimeInfo time={time} unreadCount={chat.unread_count} />
  </div>
)

// Helper component for chat status indicators
const ChatStatusIndicators = ({ chat }: { chat: Chat }) => {
  if (!chat.is_muted && !chat.is_archived) return null

  return (
    <div className="flex items-center gap-2 mt-1.5">
      {chat.is_muted && (
        <div className="flex items-center gap-1 text-xs text-gray-400">
          <Volume2 className="w-3 h-3" />
        </div>
      )}
      {chat.is_archived && (
        <div className="flex items-center gap-1 text-xs text-gray-400">
          <Archive className="w-3 h-3" />
        </div>
      )}
    </div>
  )
}

// Helper component for chat item
const ChatItem = ({
  chat,
  isSelected,
  onSelectChat,
}: {
  chat: Chat
  isSelected: boolean
  onSelectChat: (chatId: number) => void
}) => {
  const time = formatLastMessageTime(chat.last_message_at)

  return (
    <Card
      hover
      className={`cursor-pointer transition-all duration-150 p-3 ${
        isSelected
          ? 'bg-primary-50 dark:bg-primary-900/20 border-primary-500 ring-1 ring-primary-500'
          : 'hover:bg-gray-50 dark:hover:bg-gray-800/50'
      }`}
      onClick={() => onSelectChat(chat.id)}
    >
      <div className="flex items-start gap-3">
        <ChatAvatar chat={chat} isSelected={isSelected} />

        <div className="flex-1 min-w-0">
          <ChatHeader chat={chat} time={time} />

          {chat.last_message && (
            <p className="text-xs text-gray-600 dark:text-gray-400 truncate mt-1">{chat.last_message}</p>
          )}

          <ChatStatusIndicators chat={chat} />
        </div>
      </div>
    </Card>
  )
}

const ChatListHeader = ({ totalCount, chatsLength }: { totalCount?: number; chatsLength: number }) => (
  <div className="flex items-center justify-between">
    <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
      Conversations
    </h2>
    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-primary-100 dark:bg-primary-900/30 text-primary-700 dark:text-primary-400">
      {totalCount || chatsLength}
    </span>
  </div>
)

const ChatSearchInput = ({
  searchTerm,
  onSearchChange,
  onClear,
}: {
  searchTerm: string
  onSearchChange: (value: string) => void
  onClear: () => void
}) => (
  <div className="relative">
    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
    <input
      type="text"
      placeholder="Search chats..."
      value={searchTerm}
      onChange={(e) => onSearchChange(e.target.value)}
      className={SEARCH_INPUT_CLASSES}
    />
    {searchTerm && (
      <button
        onClick={onClear}
        className="absolute right-3 top-1/2 -translate-y-1/2 p-0.5 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
      >
        <X className="w-3.5 h-3.5 text-gray-400" />
      </button>
    )}
  </div>
)

const ChatResults = ({
  filteredChats,
  selectedChatId,
  onSelectChat,
  searchTerm,
}: {
  filteredChats: Chat[]
  selectedChatId: number | null
  onSelectChat: (chatId: number) => void
  searchTerm: string
}) => (
  <div className="space-y-1 max-h-[calc(100vh-320px)] overflow-y-auto">
    {filteredChats.length === 0 ? (
      <div className="text-center py-8">
        <MessageSquare className="w-8 h-8 text-gray-400 mx-auto mb-2" />
        <p className="text-sm text-gray-500 dark:text-gray-400">
          {searchTerm ? `No chats found for "${searchTerm}"` : 'No chats'}
        </p>
      </div>
    ) : (
      filteredChats.map((chat) => (
        <ChatItem
          key={chat.id}
          chat={chat}
          isSelected={chat.id === selectedChatId}
          onSelectChat={onSelectChat}
        />
      ))
    )}
  </div>
)

export const ChatList = ({ chats, selectedChatId, onSelectChat, totalCount, hasMore }: ChatListProps) => {
  const [searchTerm, setSearchTerm] = useState('')

  const filteredChats = useMemo(() => {
    if (!searchTerm) return chats
    const term = searchTerm.toLowerCase()
    return chats.filter((chat) => matchesSearchTerm(chat, term))
  }, [chats, searchTerm])

  return (
    <div className="space-y-3">
      <ChatListHeader totalCount={totalCount} chatsLength={chats.length} />
      <ChatSearchInput
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        onClear={() => setSearchTerm('')}
      />

      {searchTerm && (
        <p className="text-xs text-gray-500 dark:text-gray-400">
          {filteredChats.length} result{filteredChats.length !== 1 ? 's' : ''}
        </p>
      )}

      <ChatResults
        filteredChats={filteredChats}
        selectedChatId={selectedChatId}
        onSelectChat={onSelectChat}
        searchTerm={searchTerm}
      />

      {hasMore && !searchTerm && (
        <p className="text-xs text-center text-gray-500 dark:text-gray-400 py-2">
          There are more chats available
        </p>
      )}
    </div>
  )
}
