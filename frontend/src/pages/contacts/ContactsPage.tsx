import { useState, useMemo, useCallback, useEffect, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  Loader2,
  AlertCircle,
  Users,
  Phone,
  Clock,
  UserCheck,
  UserX,
  Search,
  X,
  Database,
  ChevronDown,
} from 'lucide-react'
import { Layout } from '@/components/layout'
import { Button, Alert, Card } from '@/components/common'
import { useInfiniteContacts, useSession } from '@/hooks'
import { Contact } from '@/api/chats.api'

const getContactSearchValue = (contact: Contact, term: string): boolean => {
  const lowerTerm = term.toLowerCase()
  const searchFields = [
    contact.first_name?.toLowerCase(),
    contact.last_name?.toLowerCase(),
    contact.username?.toLowerCase(),
    contact.phone,
  ]

  return searchFields.some((field) => field?.includes(lowerTerm) || field?.includes(term))
}

const getStatusColor = (status?: string): string => {
  if (!status) return 'text-gray-500'
  const statusLower = status.toLowerCase()
  const colorMap: Record<string, string> = {
    online: 'text-green-600 dark:text-green-400',
    recently: 'text-blue-600 dark:text-blue-400',
    offline: 'text-gray-500 dark:text-gray-500',
  }
  return colorMap[statusLower] || 'text-gray-500 dark:text-gray-500'
}

const getStatusBadge = (status?: string) => {
  if (!status || status.toLowerCase() !== 'online') return null
  return (
    <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400">
      Online
    </span>
  )
}

const formatTimeDiff = (minutes: number, hours: number, days: number): string => {
  if (minutes < 1) return 'Just now'
  if (minutes < 60) return `${minutes} min ago`
  if (hours < 24) return `${hours}h ago`
  if (days === 1) return 'Yesterday'
  if (days < 7) return `${days} days ago`
  return ''
}

const formatLastSeen = (lastSeenAt?: string) => {
  if (!lastSeenAt) return null
  const date = new Date(lastSeenAt)
  if (date.getFullYear() < 2000) return null
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / (1000 * 60))
  const hours = Math.floor(diff / (1000 * 60 * 60))
  const days = Math.floor(hours / 24)

  const formatted = formatTimeDiff(minutes, hours, days)
  return formatted || date.toLocaleDateString('en-US', { day: 'numeric', month: 'short' })
}

const ContactAvatar = ({ contact }: { contact: Contact }) => {
  const isOnline = contact.status?.toLowerCase() === 'online'
  return (
    <div className="shrink-0">
      <div
        className={`w-12 h-12 rounded-full flex items-center justify-center text-lg font-semibold ${
          isOnline
            ? 'bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400'
            : 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400'
        }`}
      >
        {contact.first_name?.charAt(0).toUpperCase() || '?'}
      </div>
    </div>
  )
}

const ContactName = ({ contact }: { contact: Contact }) => (
  <div className="flex items-start justify-between gap-2">
    <h3 className="font-semibold text-gray-900 dark:text-white truncate">
      {contact.first_name} {contact.last_name || ''}
    </h3>
    {getStatusBadge(contact.status)}
  </div>
)

const ContactPhone = ({ phone }: { phone?: string }) => {
  if (!phone) return null
  return (
    <div className="flex items-center gap-1.5 mt-1.5 text-sm text-gray-600 dark:text-gray-400">
      <Phone className="w-3.5 h-3.5 shrink-0" />
      <span className="truncate">+{phone}</span>
    </div>
  )
}

const ContactStatus = ({ contact }: { contact: Contact }) => {
  if (!contact.status || contact.status.toLowerCase() === 'online') return null
  return (
    <span className={`text-xs ${getStatusColor(contact.status)}`}>
      {contact.status === 'recently' ? 'Recent' : contact.status}
    </span>
  )
}

const ContactBadges = ({ contact }: { contact: Contact }) => (
  <div className="flex items-center gap-2 mt-2">
    {contact.is_mutual && (
      <div className="flex items-center gap-1 text-xs text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-900/20 px-2 py-0.5 rounded-full">
        <UserCheck className="w-3 h-3" />
        <span>Mutual</span>
      </div>
    )}
    {contact.is_blocked && (
      <div className="flex items-center gap-1 text-xs text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20 px-2 py-0.5 rounded-full">
        <UserX className="w-3 h-3" />
        <span>Blocked</span>
      </div>
    )}
  </div>
)

const ContactInfo = ({ contact }: { contact: Contact }) => (
  <div className="flex-1 min-w-0">
    <ContactName contact={contact} />

    {contact.username && (
      <p className="text-sm text-primary-600 dark:text-primary-400 truncate">
        @{contact.username}
      </p>
    )}

    <ContactPhone phone={contact.phone} />

    <div className="flex items-center gap-3 mt-2 flex-wrap">
      <ContactStatus contact={contact} />
      {formatLastSeen(contact.last_seen_at) && (
        <div className="flex items-center gap-1 text-xs text-gray-500 dark:text-gray-500">
          <Clock className="w-3 h-3" />
          <span>{formatLastSeen(contact.last_seen_at)}</span>
        </div>
      )}
    </div>

    <ContactBadges contact={contact} />
  </div>
)

const ContactCard = ({ contact }: { contact: Contact }) => (
  <Card hover className="p-4">
    <div className="flex items-start gap-3">
      <ContactAvatar contact={contact} />
      <ContactInfo contact={contact} />
    </div>
  </Card>
)

const PageHeader = ({ session, totalCount, fromCache, onBack }: {
  session: { phone_number?: string; session_name?: string }
  totalCount: number
  fromCache: boolean
  onBack: () => void
}) => (
  <div className="mb-6">
    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div className="flex items-center gap-3">
        <Button variant="ghost" onClick={onBack} className="searchTerm-0">
          <ArrowLeft className="w-4 h-4" />
        </Button>
        <div className="min-w-0">
          <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white truncate">
            Contacts
          </h1>
          <p className="text-sm text-gray-600 dark:text-gray-400 truncate">
            {session.phone_number || session.session_name}
          </p>
        </div>
      </div>

      <div className="flex items-center gap-2 flex-wrap">
        <span className="inline-flex items-center px-3 py-1.5 rounded-lg text-sm font-medium bg-primary-50 dark:bg-primary-900/20 text-primary-700 dark:text-primary-400">
          <Users className="w-4 h-4 mr-1.5" />
          {totalCount.toLocaleString()} contacts
        </span>
        {fromCache && (
          <span className="inline-flex items-center px-3 py-1.5 rounded-lg text-sm font-medium bg-yellow-50 dark:bg-yellow-900/20 text-yellow-700 dark:text-yellow-400">
            <Database className="w-4 h-4 mr-1.5" />
            Cache
          </span>
        )}
      </div>
    </div>
  </div>
)

const SearchBar = ({ searchTerm, onSearchChange, onClear }: {
  searchTerm: string
  onSearchChange: (value: string) => void
  onClear: () => void
}) => (
  <div className="mb-6">
    <div className="relative">
      <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
      <input
        type="text"
        placeholder="Search by name, username or phone..."
        value={searchTerm}
        onChange={(e) => onSearchChange(e.target.value)}
        className="w-full pl-10 pr-10 py-3 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-shadow"
      />
      {searchTerm && (
        <button
          onClick={onClear}
          className="absolute right-3 top-1/2 -translate-y-1/2 p-1 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
        >
          <X className="w-4 h-4 text-gray-400" />
        </button>
      )}
    </div>
  </div>
)

const EmptyState = ({ searchTerm }: { searchTerm: string }) => (
  <div className="text-center py-12">
    <div className="inline-flex items-center justify-center w-16 h-16 bg-gray-100 dark:bg-gray-800 rounded-full mb-4">
      <Users className="w-8 h-8 text-gray-400" />
    </div>
    <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
      {searchTerm ? 'No results' : 'No contacts'}
    </h3>
    <p className="text-gray-600 dark:text-gray-400">
      {searchTerm
        ? `No contacts found for "${searchTerm}"`
        : 'No contacts found in this session'}
    </p>
  </div>
)

const LoadMoreSection = ({ isFetchingNextPage, hasNextPage, onLoadMore, hasContacts }: {
  isFetchingNextPage: boolean
  hasNextPage: boolean
  onLoadMore: () => void
  hasContacts: boolean
}) => (
  <div className="mt-6 flex justify-center">
    {isFetchingNextPage ? (
      <div className="flex items-center gap-2 text-gray-500 dark:text-gray-400">
        <Loader2 className="w-5 h-5 animate-spin" />
        <span>Loading more contacts...</span>
      </div>
    ) : hasNextPage ? (
      <Button
        variant="secondary"
        onClick={onLoadMore}
        className="flex items-center gap-2"
      >
        <ChevronDown className="w-4 h-4" />
        Load more
      </Button>
    ) : hasContacts ? (
      <p className="text-sm text-gray-500 dark:text-gray-400">
        You have seen all contacts
      </p>
    ) : null}
  </div>
)

const getContactsList = (contactsData?: { pages?: Array<{ contacts: Contact[] }> }): Contact[] => {
  if (!contactsData?.pages) return []
  return contactsData.pages.flatMap((page) => page.contacts)
}

const getFirstPage = (contactsData?: { pages?: Array<{ total_count?: number; from_cache?: boolean }> }) => {
  return contactsData?.pages?.[0]
}

const getTotalCount = (firstPage?: { total_count?: number }) => {
  return firstPage?.total_count || 0
}

const getFromCache = (firstPage?: { from_cache?: boolean }) => {
  return firstPage?.from_cache || false
}

const useContactsMetadata = (contactsData?: { pages?: Array<{ total_count?: number; from_cache?: boolean }> }) => {
  const firstPage = getFirstPage(contactsData)
  const totalCount = getTotalCount(firstPage)
  const fromCache = getFromCache(firstPage)
  return { totalCount, fromCache }
}

const useContactsData = (sessionId: string, _debouncedSearch: string) => {
  const {
    data: contactsData,
    isLoading: contactsLoading,
    error,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteContacts(sessionId, 50)

  const contacts = getContactsList(contactsData)
  const { totalCount, fromCache } = useContactsMetadata(contactsData)

  return {
    contactsData,
    contactsLoading,
    error,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    contacts,
    totalCount,
    fromCache,
  }
}

const ContactsPageLoading = () => (
  <Layout>
    <div className="flex flex-col items-center justify-center py-12 gap-3">
      <Loader2 className="w-8 h-8 animate-spin text-primary-600" />
      <p className="text-sm text-gray-500 dark:text-gray-400">Loading contacts...</p>
    </div>
  </Layout>
)

const ContactsPageInactiveSession = ({ onBack }: { onBack: () => void }) => (
  <Layout>
    <div className="max-w-2xl mx-auto text-center py-12">
      <AlertCircle className="w-16 h-16 text-yellow-500 mx-auto mb-4" />
      <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
        Session not active
      </h2>
      <p className="text-gray-600 dark:text-gray-400 mb-6">
        This session is not active. Please verify the session first.
      </p>
      <Button variant="primary" onClick={onBack}>
        <ArrowLeft className="w-4 h-4 mr-2" />
        Back to Dashboard
      </Button>
    </div>
  </Layout>
)

const ContactsPageSearchResults = ({ searchTerm, filteredContacts }: {
  searchTerm: string
  filteredContacts: Contact[]
}) => (
  <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
    {filteredContacts.length} resultados para &apos;{searchTerm}&apos;
  </p>
)

const ContactsPageError = () => (
  <Alert variant="error" className="mb-6">
    <div className="flex items-center gap-2">
      <AlertCircle className="w-5 h-5" />
      <span>Error loading contacts. Please try again.</span>
    </div>
  </Alert>
)

const ContactsPageGrid = ({ filteredContacts }: { filteredContacts: Contact[] }) => (
  <div className="grid gap-3 sm:gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
    {filteredContacts.map((contact: Contact) => (
      <ContactCard key={contact.id} contact={contact} />
    ))}
  </div>
)

const ContactsPageLoadMore = ({
  loadMoreRef,
  isFetchingNextPage,
  hasNextPage,
  fetchNextPage,
  contacts,
}: {
  loadMoreRef: React.RefObject<HTMLDivElement | null>
  isFetchingNextPage: boolean
  hasNextPage: boolean
  fetchNextPage: () => void
  contacts: Contact[]
}) => (
  <div ref={loadMoreRef}>
    <LoadMoreSection
      isFetchingNextPage={isFetchingNextPage}
      hasNextPage={hasNextPage}
      onLoadMore={() => fetchNextPage()}
      hasContacts={contacts.length >= 50}
    />
  </div>
)

const ContactsPageContentBody = ({
  searchTerm,
  setSearchTerm,
  filteredContacts,
  error,
  isLoading,
  loadMoreRef,
  isFetchingNextPage,
  hasNextPage,
  fetchNextPage,
  contacts,
}: {
  searchTerm: string
  setSearchTerm: (value: string) => void
  filteredContacts: Contact[]
  error: unknown
  isLoading: boolean
  loadMoreRef: React.RefObject<HTMLDivElement | null>
  isFetchingNextPage: boolean
  hasNextPage: boolean
  fetchNextPage: () => void
  contacts: Contact[]
}) => (
  <>
    <SearchBar
      searchTerm={searchTerm}
      onSearchChange={setSearchTerm}
      onClear={() => setSearchTerm('')}
    />

    {searchTerm && <ContactsPageSearchResults searchTerm={searchTerm} filteredContacts={filteredContacts} />}

    {error && <ContactsPageError />}

    {filteredContacts.length === 0 && !isLoading && <EmptyState searchTerm={searchTerm} />}

    {filteredContacts.length > 0 && (
      <>
        <ContactsPageGrid filteredContacts={filteredContacts} />
        <ContactsPageLoadMore
          loadMoreRef={loadMoreRef}
          isFetchingNextPage={isFetchingNextPage}
          hasNextPage={hasNextPage}
          fetchNextPage={fetchNextPage}
          contacts={contacts}
        />
      </>
    )}
  </>
)

interface ContactsPageContentProps {
  session: { phone_number?: string; session_name?: string }
  searchTerm: string
  setSearchTerm: (value: string) => void
  filteredContacts: Contact[]
  totalCount: number
  fromCache: boolean
  error: unknown
  isLoading: boolean
  loadMoreRef: React.RefObject<HTMLDivElement | null>
  isFetchingNextPage: boolean
  hasNextPage: boolean
  fetchNextPage: () => void
  contacts: Contact[]
  onBack: () => void
}

const ContactsPageContent = ({
  session,
  searchTerm,
  setSearchTerm,
  filteredContacts,
  totalCount,
  fromCache,
  error,
  isLoading,
  loadMoreRef,
  isFetchingNextPage,
  hasNextPage,
  fetchNextPage,
  contacts,
  onBack,
}: ContactsPageContentProps) => (
  <Layout>
    <div className="max-w-7xl mx-auto">
      <PageHeader
        session={session}
        totalCount={totalCount}
        fromCache={fromCache}
        onBack={onBack}
      />

      <ContactsPageContentBody
        searchTerm={searchTerm}
        setSearchTerm={setSearchTerm}
        filteredContacts={filteredContacts}
        error={error}
        isLoading={isLoading}
        loadMoreRef={loadMoreRef}
        isFetchingNextPage={isFetchingNextPage}
        hasNextPage={hasNextPage}
        fetchNextPage={fetchNextPage}
        contacts={contacts}
      />
    </div>
  </Layout>
)

const useFilteredContacts = (contacts: Contact[], searchTerm: string) => {
  return useMemo(() => {
    if (!searchTerm) return contacts
    return contacts.filter((contact) => getContactSearchValue(contact, searchTerm))
  }, [contacts, searchTerm])
}

const useContactsObserver = (
  loadMoreRef: React.RefObject<HTMLDivElement | null>,
  hasNextPage: boolean,
  isFetchingNextPage: boolean,
  fetchNextPage: () => void
) => {
  const handleObserver = useCallback(
    (entries: IntersectionObserverEntry[]) => {
      const target = entries[0]
      if (target.isIntersecting && hasNextPage && !isFetchingNextPage) {
        fetchNextPage()
      }
    },
    [fetchNextPage, hasNextPage, isFetchingNextPage]
  )

  useEffect(() => {
    const option = {
      root: null,
      rootMargin: '100px',
      threshold: 0,
    }
    const observer = new IntersectionObserver(handleObserver, option)
    if (loadMoreRef.current) observer.observe(loadMoreRef.current)
    return () => observer.disconnect()
  }, [handleObserver, loadMoreRef])
}

const ContactsPageInvalidSession = () => (
  <Layout>
    <Alert variant="error">Invalid session ID</Alert>
  </Layout>
)

const ContactsPageNotFound = () => (
  <Layout>
    <Alert variant="error">Session not found</Alert>
  </Layout>
)

const ContactsPageMain = ({
  session,
  searchTerm,
  setSearchTerm,
  filteredContacts,
  totalCount,
  fromCache,
  error,
  isLoading,
  loadMoreRef,
  isFetchingNextPage,
  hasNextPage,
  fetchNextPage,
  contacts,
  onBack,
}: {
  session: { phone_number?: string; session_name?: string; is_active: boolean }
  searchTerm: string
  setSearchTerm: (value: string) => void
  filteredContacts: Contact[]
  totalCount: number
  fromCache: boolean
  error: unknown
  isLoading: boolean
  loadMoreRef: React.RefObject<HTMLDivElement | null>
  isFetchingNextPage: boolean
  hasNextPage: boolean
  fetchNextPage: () => void
  contacts: Contact[]
  onBack: () => void
}) => (
  <ContactsPageContent
    session={session}
    searchTerm={searchTerm}
    setSearchTerm={setSearchTerm}
    filteredContacts={filteredContacts}
    totalCount={totalCount}
    fromCache={fromCache}
    error={error}
    isLoading={isLoading}
    loadMoreRef={loadMoreRef}
    isFetchingNextPage={isFetchingNextPage}
    hasNextPage={hasNextPage}
    fetchNextPage={fetchNextPage}
    contacts={contacts}
    onBack={onBack}
  />
)

const useContactsPageState = (sessionId: string, searchTerm: string) => {
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const loadMoreRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchTerm)
    }, 300)
    return () => clearTimeout(timer)
  }, [searchTerm])

  const { data: sessionData, isLoading: sessionLoading } = useSession(sessionId)
  const contactsData = useContactsData(sessionId, debouncedSearch)

  const isLoading = sessionLoading || contactsData.contactsLoading

  const filteredContacts = useFilteredContacts(contactsData.contacts, searchTerm)

  useContactsObserver(
    loadMoreRef,
    contactsData.hasNextPage,
    contactsData.isFetchingNextPage,
    contactsData.fetchNextPage
  )

  return {
    sessionData,
    contactsData,
    isLoading,
    filteredContacts,
    loadMoreRef,
  }
}

const ContactsPageValidation = ({
  sessionId,
  isLoading,
  sessionData,
}: {
  sessionId?: string
  isLoading: boolean
  sessionData?: { session: { is_active: boolean } }
}) => {
  if (!sessionId) {
    return <ContactsPageInvalidSession />
  }

  if (isLoading && !sessionData) {
    return <ContactsPageLoading />
  }

  if (!sessionData) {
    return <ContactsPageNotFound />
  }

  const session = sessionData.session

  if (!session.is_active) {
    return <ContactsPageInactiveSession onBack={() => {
      // Navigate will be handled by parent component
      window.location.href = '/dashboard'
    }} />
  }

  return null
}

export const ContactsPage = () => {
  const { sessionId } = useParams<{ sessionId: string }>()
  const navigate = useNavigate()
  const [searchTerm, setSearchTerm] = useState('')

  const { sessionData, contactsData, isLoading, filteredContacts, loadMoreRef } =
    useContactsPageState(sessionId!, searchTerm)

  const validation = ContactsPageValidation({
    sessionId,
    isLoading,
    sessionData,
  })

  if (validation) {
    return validation
  }

  const session = sessionData!.session

  return (
    <ContactsPageMain
      session={session}
      searchTerm={searchTerm}
      setSearchTerm={setSearchTerm}
      filteredContacts={filteredContacts}
      totalCount={contactsData.totalCount}
      fromCache={contactsData.fromCache}
      error={contactsData.error}
      isLoading={isLoading}
      loadMoreRef={loadMoreRef}
      isFetchingNextPage={contactsData.isFetchingNextPage}
      hasNextPage={contactsData.hasNextPage}
      fetchNextPage={contactsData.fetchNextPage}
      contacts={contactsData.contacts}
      onBack={() => navigate('/dashboard')}
    />
  )
}
