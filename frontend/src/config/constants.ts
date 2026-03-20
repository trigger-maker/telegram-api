export const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1'

export const AUTH_TOKEN_KEY = 'auth_token'
export const REFRESH_TOKEN_KEY = 'refresh_token'
export const USER_KEY = 'user_data'

// Base frontend URL for public files
export const FRONTEND_URL = 'https://frontend.telegram-api.fututel.com'

// Base URL for uploads
export const UPLOADS_BASE_URL = `${FRONTEND_URL}/uploads`

export const ROUTES = {
  LOGIN: '/login',
  REGISTER: '/register',
  DASHBOARD: '/dashboard',
  SESSIONS: '/sessions',
  MESSAGES: '/messages',
  CHATS: '/chats',
  CONTACTS: '/contacts',
  WEBHOOKS: '/webhooks',
  PROFILE: '/profile',
  SETTINGS: '/settings',
} as const

export const AUTH_STATES = {
  PENDING: 'pending',
  CODE_SENT: 'code_sent',
  PASSWORD_REQUIRED: 'password_required',
  AUTHENTICATED: 'authenticated',
  FAILED: 'failed',
} as const

export const AUTH_METHODS = {
  SMS: 'sms',
  QR: 'qr',
} as const

// Allowed file types
export const ALLOWED_FILE_TYPES = {
  image: ['image/jpeg', 'image/png', 'image/gif', 'image/webp'],
  video: ['video/mp4', 'video/webm', 'video/quicktime'],
  audio: ['audio/mpeg', 'audio/ogg', 'audio/wav', 'audio/mp3'],
  file: [
    'application/pdf',
    'application/msword',
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    'text/plain',
  ],
}

// Maximum sizes (in bytes)
export const MAX_FILE_SIZES = {
  image: 10 * 1024 * 1024, // 10MB
  video: 50 * 1024 * 1024, // 50MB
  audio: 20 * 1024 * 1024, // 20MB
  file: 50 * 1024 * 1024,  // 50MB
}

// Available webhook events
export const WEBHOOK_EVENTS = [
  { id: 'message.new', label: 'New message', description: 'When a new message arrives' },
  { id: 'message.edit', label: 'Message edited', description: 'When a message is edited' },
  { id: 'message.delete', label: 'Message deleted', description: 'When a message is deleted' },
  { id: 'user.online', label: 'User connected', description: 'When a user connects' },
  { id: 'user.offline', label: 'User disconnected', description: 'When a user disconnects' },
  { id: 'user.typing', label: 'User typing', description: 'When a user is typing' },
  { id: 'session.started', label: 'Session started', description: 'When a session starts' },
  { id: 'session.stopped', label: 'Session stopped', description: 'When a session stops' },
  { id: 'session.error', label: 'Session error', description: 'When there is a session error' },
]

// Session states with colors
export const SESSION_STATE_CONFIG = {
  pending: {
    label: 'Pending',
    color: 'yellow',
    bg: 'bg-yellow-100 dark:bg-yellow-900/30',
    text: 'text-yellow-600 dark:text-yellow-400',
  },
  code_sent: {
    label: 'Code sent',
    color: 'blue',
    bg: 'bg-blue-100 dark:bg-blue-900/30',
    text: 'text-blue-600 dark:text-blue-400',
  },
  password_required: {
    label: 'Password required',
    color: 'orange',
    bg: 'bg-orange-100 dark:bg-orange-900/30',
    text: 'text-orange-600 dark:text-orange-400',
  },
  authenticated: {
    label: 'Authenticated',
    color: 'green',
    bg: 'bg-green-100 dark:bg-green-900/30',
    text: 'text-green-600 dark:text-green-400',
  },
  failed: {
    label: 'Failed',
    color: 'red',
    bg: 'bg-red-100 dark:bg-red-900/30',
    text: 'text-red-600 dark:text-red-400',
  },
}

// Chat types
export const CHAT_TYPES = {
  private: { label: 'Private', icon: 'User' },
  group: { label: 'Group', icon: 'Users' },
  supergroup: { label: 'Supergroup', icon: 'Users' },
  channel: { label: 'Channel', icon: 'Radio' },
}
