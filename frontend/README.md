# Telegram API Manager - Frontend

Modern and minimalist interface for managing Telegram sessions, built with React 19, TypeScript, Tailwind CSS, and TanStack Query.

## Features

- **Modern UI** - Minimalist design with Tailwind CSS and custom components
- **Dark/Light Mode** - Persistent automatic theme with system support
- **Responsive** - Mobile-first design with collapsible sidebar
- **Performance** - Optimized with TanStack Query and lazy loading
- **Authentication** - JWT with protected routes and automatic refresh
- **TypeScript** - Type-safe throughout the application
- **Vite 7** - Fast build and instant HMR
- **Toast Notifications** - Contextual notification system
- **File Upload** - Support for images, videos, audio, and files

## Architecture

```
src/
├── api/                    # API clients (axios)
│   ├── client.ts               # HTTP client configured with interceptors
│   ├── auth.api.ts             # Authentication endpoints
│   ├── sessions.api.ts         # Telegram session endpoints
│   ├── messages.api.ts         # Message endpoints (text, photo, video, audio, file, bulk)
│   ├── chats.api.ts            # Chats and contacts endpoints
│   └── webhooks.api.ts         # Webhook endpoints
│
├── components/             # Reusable components
│   ├── common/                 # UI Components
│   │   ├── Alert.tsx               # Alerts (success, error, warning, info)
│   │   ├── Badge.tsx               # Badges with variants
│   │   ├── Button.tsx              # Buttons (primary, secondary, danger, ghost)
│   │   ├── Card.tsx                # Cards with optional hover
│   │   ├── FileUpload.tsx          # File upload with preview
│   │   ├── Input.tsx               # Inputs with label and error
│   │   ├── Modal.tsx               # Responsive modal with sizes
│   │   └── Tabs.tsx                # Tab system
│   ├── layout/                 # Layout Components
│   │   ├── Header.tsx              # Header with search and user menu
│   │   ├── Layout.tsx              # Main layout with sidebar
│   │   └── Sidebar.tsx             # Collapsible sidebar with navigation
│   └── sessions/               # Session Components
│       ├── CreateSessionModal.tsx  # Create session modal (SMS/QR)
│       ├── QRCodeModal.tsx         # Modal to scan QR
│       ├── SessionCard.tsx         # Session card
│       └── VerifyCodeModal.tsx     # Verify SMS code modal
│
├── contexts/               # React Contexts
│   ├── AuthContext.tsx         # Global authentication state
│   ├── ThemeContext.tsx        # Dark/light theme
│   └── ToastContext.tsx        # Toast notification system
│
├── hooks/                  # Custom hooks
│   ├── index.ts                # Re-exports
│   ├── useSessions.ts          # Session CRUD with TanStack Query
│   ├── useMessages.ts          # Message sending (all types)
│   └── useChats.ts             # Chats, history, and contacts
│
├── pages/                  # Main pages
│   ├── auth/                   # Authentication
│   │   ├── LoginPage.tsx           # Login with split design
│   │   └── RegisterPage.tsx        # User registration
│   ├── dashboard/              # Dashboard
│   │   ├── DashboardPage.tsx       # Main view with stats
│   │   └── components/
│   │       └── SessionList.tsx     # Session list
│   ├── messages/               # Messaging
│   │   ├── MessagesPage.tsx        # Message sending page
│   │   └── components/
│   │       ├── SendTextForm.tsx        # Send text
│   │       ├── SendPhotoForm.tsx       # Send photo
│   │       ├── SendVideoForm.tsx       # Send video
│   │       ├── SendAudioForm.tsx       # Send audio
│   │       ├── SendFileForm.tsx        # Send file
│   │       └── SendBulkForm.tsx        # Bulk send
│   ├── chats/                  # Chats
│   │   ├── ChatsPage.tsx           # Chats view
│   │   └── components/
│   │       ├── ChatList.tsx            # Chat list
│   │       └── ChatView.tsx            # Conversation view
│   ├── contacts/               # Contacts
│   │   └── ContactsPage.tsx        # Contact list
│   ├── webhooks/               # Webhooks
│   │   └── WebhooksPage.tsx        # Webhook configuration
│   ├── profile/                # Profile
│   │   └── ProfilePage.tsx         # User profile
│   └── settings/               # Settings
│       └── SettingsPage.tsx        # App settings
│
├── routes/                 # Route configuration
│   ├── ProtectedRoute.tsx      # HOC for protected routes
│   └── index.tsx               # All routes definition
│
├── types/                  # TypeScript types
│   ├── auth.types.ts           # Authentication types
│   ├── session.types.ts        # Session types
│   └── api.types.ts            # Generic API types
│
├── config/                 # Configuration
│   └── constants.ts            # URLs, webhook events, file types
│
├── utils/                  # Utilities
│   └── upload.ts               # File validation and processing
│
└── styles/                 # Global styles
    └── index.css               # Tailwind + custom animations
```

## Installation

### Requirements

- Node.js 18+
- pnpm 8+

### 1. Install dependencies

```bash
cd frontend
pnpm install
```

### 2. Configure environment variables

```bash
cp .env.example .env
```

Edit `.env`:

```env
VITE_API_URL=/api/v1
```

### 3. Run in development

```bash
pnpm dev
```

The application will be available at `http://localhost:3000`

### 4. Build for production

```bash
pnpm build
```

Compiled files will be in `/dist`

## Available Scripts

```bash
pnpm dev        # Start development server
pnpm build      # Build for production (tsc + vite build)
pnpm preview    # Preview production build
pnpm lint       # Run ESLint
```

## Main Dependencies

| Package | Version | Purpose |
|---------|---------|-----------|
| React | 19.x | UI Library |
| TypeScript | 5.x | Type Safety |
| Vite | 7.x | Build Tool |
| React Router | 7.x | Routing |
| TanStack Query | 5.x | Data Fetching & Caching |
| Axios | 1.x | HTTP Client |
| Tailwind CSS | 4.x | Styling |
| Lucide React | 0.x | Icons |

## UI Components

### Button
```tsx
<Button variant="primary" isLoading={false} fullWidth>
  Click me
</Button>
// Variants: primary, secondary, danger, ghost
```

### Input
```tsx
<Input
  label="Username"
  type="text"
  error="Error message"
  icon={<User />}
/>
```

### Card
```tsx
<Card hover onClick={() => {}}>
  Content
</Card>
```

### Alert
```tsx
<Alert variant="success">
  Success message
</Alert>
// Variants: success, error, warning, info
```

### Modal
```tsx
<Modal isOpen={open} onClose={close} title="Title" size="lg">
  Content
</Modal>
// Sizes: sm, md, lg, xl
```

### FileUpload
```tsx
<FileUpload
  type="image"
  value={url}
  onChange={setUrl}
  label="Image"
/>
// Types: image, video, audio, file
```

### Toast (via context)
```tsx
const toast = useToast()
toast.success('Title', 'Message')
toast.error('Error', 'Description')
toast.info('Info', 'Informational message')
toast.warning('Warning', 'Be careful')
```

## Authentication

The system uses JWT tokens with refresh tokens:

1. **Login/Register** - POST `/api/v1/auth/login` or `/api/v1/auth/register`
2. **Tokens saved** in `localStorage`
3. **Auto-refresh** when token is about to expire
4. **Protected routes** with `ProtectedRoute`
5. **Axios interceptor** adds token automatically

## API Integration

### Interceptors

- **Request**: Adds JWT token automatically to all requests
- **Response**: Handles errors globally, redirects to login if 401

### TanStack Query

All requests use custom hooks with cache:

```tsx
import { useSessions, useCreateSession } from '@/hooks'

// Query with cache
const { data, isLoading, error, refetch } = useSessions()

// Mutation
const createSession = useCreateSession()
createSession.mutate(data, {
  onSuccess: () => toast.success('Success', 'Session created'),
  onError: (err) => toast.error('Error', err.message)
})
```

## Available Pages

| Route | Page | Description |
|------|--------|-------------|
| `/login` | LoginPage | Login |
| `/register` | RegisterPage | User registration |
| `/dashboard` | DashboardPage | Main panel with sessions |
| `/messages/:sessionId` | MessagesPage | Send messages |
| `/chats/:sessionId` | ChatsPage | View chats and conversations |
| `/contacts/:sessionId` | ContactsPage | Contact list |
| `/webhooks/:sessionId` | WebhooksPage | Configure webhooks |
| `/profile` | ProfilePage | User profile |
| `/settings` | SettingsPage | App settings |

## Dark/Light Theme

The theme is automatically saved in `localStorage` and respects system preferences:

```tsx
import { useTheme } from '@/contexts'

const { theme, toggleTheme } = useTheme()
// theme: 'light' | 'dark'
```

## Responsive Design

Designed mobile-first with Tailwind breakpoints:

- `sm`: 640px - Large mobile
- `md`: 768px - Tablets
- `lg`: 1024px - Desktop
- `xl`: 1280px - Large desktop

The sidebar automatically collapses on small screens.

## Webhooks

Available events to configure:

- `new_message` - New message received
- `message_edited` - Message edited
- `message_deleted` - Message deleted
- `user_status` - User status change
- `user_typing` - User typing
- `chat_action` - Chat actions

## Upload File Structure

```
/uploads/
├── images/     # Images (jpg, png, gif, webp)
├── videos/     # Videos (mp4, webm, mov)
├── audio/      # Audio (mp3, ogg, wav)
└── files/      # Documents (pdf, doc, docx, txt)
```

Size limits:
- Images: 10MB
- Videos: 50MB
- Audio: 20MB
- Files: 50MB

## Code Conventions

- **Components**: PascalCase (`LoginPage.tsx`)
- **Hooks**: camelCase with `use` prefix (`useSessions.ts`)
- **Types**: PascalCase (`SessionStatus`, `AuthContextType`)
- **Constants**: UPPER_SNAKE_CASE (`API_BASE_URL`)
- **CSS Classes**: Tailwind utilities
- **Files**: kebab-case for utils, PascalCase for components

## License

MIT
