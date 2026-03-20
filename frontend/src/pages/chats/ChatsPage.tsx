import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ArrowLeft, Loader2, AlertCircle, MessageCircle, ChevronLeft } from 'lucide-react'
import { Layout } from '@/components/layout'
import { Button, Alert } from '@/components/common'
import { useChats, useSession } from '@/hooks'
import { ChatList } from './components/ChatList'
import { ChatView } from './components/ChatView'

// Tailwind classes for empty state
const EMPTY_STATE_CLASSES = 'flex items-center justify-center h-full min-h-[500px] bg-gray-50 dark:bg-gray-800/50 rounded-xl border-2 border-dashed border-gray-300 dark:border-gray-700'

/* eslint-disable max-lines-per-function, complexity */
export const ChatsPage = () => {
  const { sessionId } = useParams<{ sessionId: string }>()
  const navigate = useNavigate()
  const [selectedChatId, setSelectedChatId] = useState<number | null>(null)

  const { data: sessionData, isLoading: sessionLoading } = useSession(sessionId!)
  const { data: chatsData, isLoading: chatsLoading, error } = useChats(sessionId!, { limit: 100 })

  const isLoading = sessionLoading || chatsLoading

  // Handle back from chat view on mobile
  const handleBackToList = () => {
    setSelectedChatId(null)
  }

  if (!sessionId) {
    return (
      <Layout>
        <Alert variant="error">Invalid session ID</Alert>
      </Layout>
    )
  }

  if (isLoading) {
    return (
      <Layout>
        <div className="flex flex-col items-center justify-center py-12 gap-3">
          <Loader2 className="w-8 h-8 animate-spin text-primary-600" />
          <p className="text-sm text-gray-500 dark:text-gray-400">Loading chats...</p>
        </div>
      </Layout>
    )
  }

  if (!sessionData) {
    return (
      <Layout>
        <Alert variant="error">Session not found</Alert>
      </Layout>
    )
  }

  const session = sessionData.session

  if (!session.is_active) {
    return (
      <Layout>
        <div className="max-w-2xl mx-auto text-center py-12">
          <AlertCircle className="w-16 h-16 text-yellow-500 mx-auto mb-4" />
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
            Session not active
          </h2>
          <p className="text-gray-600 dark:text-gray-400 mb-6">
            This session is not active. Please verify the session first.
          </p>
          <Button variant="primary" onClick={() => navigate('/dashboard')}>
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Dashboard
          </Button>
        </div>
      </Layout>
    )
  }

  return (
    <Layout>
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-4 sm:mb-6">
          <div className="flex items-center gap-3">
            <Button variant="ghost" onClick={() => navigate('/dashboard')} className="shrink-0">
              <ArrowLeft className="w-4 h-4" />
            </Button>
            <div className="min-w-0">
              <h1 className="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white truncate">
                Chats
              </h1>
              <p className="text-sm text-gray-600 dark:text-gray-400 truncate">
                {session.phone_number || session.session_name}
              </p>
            </div>
          </div>
        </div>

        {error && (
          <Alert variant="error" className="mb-6">
            <div className="flex items-center gap-2">
              <AlertCircle className="w-5 h-5" />
              <span>Error loading chats. Please try again.</span>
            </div>
          </Alert>
        )}

        {chatsData && chatsData.chats.length === 0 && (
          <div className="text-center py-12">
            <div className="inline-flex items-center justify-center w-16 h-16 bg-gray-100 dark:bg-gray-800 rounded-full mb-4">
              <MessageCircle className="w-8 h-8 text-gray-400" />
            </div>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
              No chats
            </h3>
            <p className="text-gray-600 dark:text-gray-400">
              No conversations found in this session
            </p>
          </div>
        )}

        {chatsData && chatsData.chats.length > 0 && (
          <>
            {/* Desktop layout */}
            <div className="hidden lg:grid lg:grid-cols-3 gap-6">
              <div className="lg:col-span-1">
                <ChatList
                  chats={chatsData.chats}
                  selectedChatId={selectedChatId}
                  onSelectChat={setSelectedChatId}
                  totalCount={chatsData.total_count}
                  hasMore={chatsData.has_more}
                />
              </div>

              <div className="lg:col-span-2">
                {selectedChatId ? (
                  <ChatView sessionId={sessionId} chatId={selectedChatId} />
                ) : (
                  <div className={EMPTY_STATE_CLASSES}>
                    <div className="text-center">
                      <MessageCircle className="w-12 h-12 text-gray-400 mx-auto mb-3" />
                      <p className="text-gray-500 dark:text-gray-400">
                        Select a chat to view the conversation
                      </p>
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* Mobile layout - shows either list or chat view */}
            <div className="lg:hidden">
              {selectedChatId ? (
                <div>
                  {/* Back button for mobile */}
                  <button
                    onClick={handleBackToList}
                    className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white mb-4 transition-colors"
                  >
                    <ChevronLeft className="w-4 h-4" />
                    Back to list
                  </button>
                  <ChatView sessionId={sessionId} chatId={selectedChatId} />
                </div>
              ) : (
                <ChatList
                  chats={chatsData.chats}
                  selectedChatId={selectedChatId}
                  onSelectChat={setSelectedChatId}
                  totalCount={chatsData.total_count}
                  hasMore={chatsData.has_more}
                />
              )}
            </div>
          </>
        )}
      </div>
    </Layout>
  )
}
