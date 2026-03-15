# 🚀 Telegram API

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://docker.com)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Version](https://img.shields.io/badge/Version-0.1.0-blue.svg)](https://github.com/ghmedinac/telegram-api)

Multi-session REST API for Telegram using MTProto. Manages multiple Telegram accounts, sends bulk messages, and receives real-time events via webhooks.

## 📋 Features

- ✅ **Multi-session** - Manage multiple Telegram accounts simultaneously
- ✅ **JWT Authentication** - Registration, login, refresh tokens
- ✅ **Telegram Auth** - Via SMS or QR code with automatic regeneration
- ✅ **Messaging** - Text, photos, videos, audio, documents
- ✅ **Bulk Messaging** - Bulk messaging with configurable delay
- ✅ **Webhooks** - Receive real-time events (messages, statuses, etc)
- ✅ **Chats & Contacts** - List dialogs, history, contacts
- ✅ **AES-256 Encryption** - Sensitive data encrypted
- ✅ **Rate Limiting** - Flood protection
- ✅ **Documentation** - Swagger UI, ReDoc, Postman Collection

## 📚 Documentation

| URL | Description |
|-----|-------------|
| [http://localhost:7789/docs/](http://localhost:7789/docs/) | **Swagger UI** - Interactive documentation |
| [http://localhost:7789/redoc](http://localhost:7789/redoc) | **ReDoc** - Elegant documentation |
| [http://localhost:7789/health](http://localhost:7789/health) | Health check + version |

## 🏗️ Architecture

```
telegram-api/
├── cmd/api/main.go              # Entry point
├── internal/
│   ├── config/                  # Configuration
│   ├── domain/                  # Entities and DTOs
│   ├── handler/                 # HTTP controllers (Fiber)
│   ├── middleware/              # JWT, CORS, Logger, RateLimit
│   ├── repository/
│   │   ├── postgres/            # PostgreSQL repositories
│   │   └── redis/               # Redis cache
│   ├── service/                 # Business logic
│   └── telegram/                # MTProto client (gotd/td)
├── pkg/                         # Reusable packages
├── db/migrations/               # SQL migrations
├── docs/                        # Swagger, ReDoc, Postman
└── docker-compose.yml
```

## 🚀 Installation

### Requirements
- Go 1.23+
- PostgreSQL 16+
- Redis 7+
- Docker (recommended)

### Option 1: Docker (recommended)

```bash
# Clone
git clone https://github.com/ghmedinac/telegram-api.git
cd telegram-api

# Configure
cp .env.example .env
# Edit .env with your values

# Run everything
docker-compose up -d

# View logs
docker-compose logs -f api
```

### Option 2: Local

```bash
# Start only DB and Redis
docker-compose up -d postgres redis

# Build and run
go build ./cmd/api && ./api
```

### Option 3: From Docker Hub

```bash
docker pull ghmedinac/telegram-api:latest

docker run -d \
  --name telegram-api \
  -p 7789:8080 \
  -e DB_URL="postgres://user:pass@host:5432/db" \
  -e REDIS_ADDR="redis:6379" \
  -e JWT_SECRET="your_secret_32_chars" \
  -e ENCRYPTION_KEY="your_key_32_chars!!" \
  ghmedinac/telegram-api:latest
```

## ⚙️ Configuration

```env
# API
API_PORT=7789
API_ENV=production
LOG_LEVEL=info

# PostgreSQL
DB_URL=postgres://admin:password123@localhost:5432/telegram_db?sslmode=disable

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=

# JWT (minimum 32 characters)
JWT_SECRET=your_very_long_and_secure_jwt_secret!
JWT_EXPIRY=24h

# Encryption (exactly 32 characters)
ENCRYPTION_KEY=exactly_32_characters_key!!
```

## 📖 Endpoints

### 🔐 Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register user |
| POST | `/api/v1/auth/login` | Login → JWT |
| POST | `/api/v1/auth/refresh` | Refresh token |
| POST | `/api/v1/auth/logout` | Logout |
| GET | `/api/v1/auth/me` | Current user |

### 📱 Telegram Sessions

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/sessions` | Create session (SMS/QR) |
| GET | `/api/v1/sessions` | List sessions |
| GET | `/api/v1/sessions/:id` | Get session |
| POST | `/api/v1/sessions/:id/verify` | Verify SMS code |
| DELETE | `/api/v1/sessions/:id` | Delete session |

### 💬 Messages

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/sessions/:id/messages/text` | Send text |
| POST | `/api/v1/sessions/:id/messages/photo` | Send photo |
| POST | `/api/v1/sessions/:id/messages/video` | Send video |
| POST | `/api/v1/sessions/:id/messages/audio` | Send audio |
| POST | `/api/v1/sessions/:id/messages/file` | Send file |
| POST | `/api/v1/sessions/:id/messages/bulk` | Bulk send |
| GET | `/api/v1/messages/:jobId/status` | Send status |

### 📋 Chats & Contacts

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/sessions/:id/chats` | List chats |
| GET | `/api/v1/sessions/:id/chats/:chatId` | Chat info |
| GET | `/api/v1/sessions/:id/chats/:chatId/history` | History |
| GET | `/api/v1/sessions/:id/contacts` | List contacts |
| POST | `/api/v1/sessions/:id/resolve` | Resolve @username |

### 🔔 Webhooks

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/sessions/:id/webhook` | Configure webhook |
| GET | `/api/v1/sessions/:id/webhook` | Get config |
| DELETE | `/api/v1/sessions/:id/webhook` | Delete |
| POST | `/api/v1/sessions/:id/webhook/start` | Start listening |
| POST | `/api/v1/sessions/:id/webhook/stop` | Stop listening |
| GET | `/api/v1/pool/status` | Pool status |

## 🔐 Authentication Flows

### SMS Flow

```bash
# 1. Create session
curl -X POST http://localhost:7789/api/v1/sessions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+573001234567",
    "api_id": 12345678,
    "api_hash": "your_api_hash",
    "session_name": "my_account"
  }'

# 2. Verify SMS code
curl -X POST http://localhost:7789/api/v1/sessions/{id}/verify \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"code": "12345"}'
```

### QR Flow

```bash
# 1. Create QR session
curl -X POST http://localhost:7789/api/v1/sessions \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "api_id": 12345678,
    "api_hash": "your_api_hash",
    "auth_method": "qr",
    "session_name": "my_qr_account"
  }'
# Response includes qr_image_base64

# QR regenerates automatically (max 3 attempts)
```

## 📤 Sending Messages

```bash
# Simple text
curl -X POST http://localhost:7789/api/v1/sessions/{id}/messages/text \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"to": "@username", "text": "Hello!"}'

# With photo
curl -X POST http://localhost:7789/api/v1/sessions/{id}/messages/photo \
  -d '{"to": "@username", "photo_url": "https://...", "caption": "Look!"}'

# Bulk
curl -X POST http://localhost:7789/api/v1/sessions/{id}/messages/bulk \
  -d '{
    "recipients": ["@user1", "@user2", "+57300..."],
    "text": "Message for everyone",
    "delay_ms": 3000
  }'
```

## 🔔 Configure Webhook

```bash
# Configure URL
curl -X POST http://localhost:7789/api/v1/sessions/{id}/webhook \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "url": "https://your-server.com/webhook",
    "secret": "my_secret",
    "events": ["message.new", "user.online"]
  }'

# Start listening
curl -X POST http://localhost:7789/api/v1/sessions/{id}/webhook/start
```

### Available events:
- `message.new` - New message
- `message.edit` - Message edited
- `message.delete` - Message deleted
- `user.online` - User online
- `user.offline` - User offline
- `user.typing` - User typing
- `session.started` - Session started
- `session.stopped` - Session stopped
- `session.error` - Session error

## 🐳 Deploy

```bash
# Deploy new version
./deploy.sh 0.1.0

# The script:
# 1. Stops current container
# 2. Rebuilds image
# 3. Pushes to Docker Hub
# 4. Starts new container
# 5. Verifies health
```

## 📝 Get Telegram API ID

1. Go to https://my.telegram.org
2. Login with your number
3. Go to "API development tools"
4. Create new application
5. Copy `api_id` and `api_hash`

## 🛠️ Development

```bash
# Regenerate Swagger
swag init -g cmd/api/main.go -o docs

# Generate Postman collection
./generate-postman.sh

# Tests
go test ./...

# Build
go build ./cmd/api
```

## 📚 Tech Stack

| Technology | Usage |
|------------|-----|
| [Go 1.23](https://golang.org) | Language |
| [Fiber v2](https://gofiber.io) | HTTP Framework |
| [gotd/td](https://github.com/gotd/td) | Telegram MTProto Client |
| [pgx v5](https://github.com/jackc/pgx) | PostgreSQL Driver |
| [go-redis v9](https://github.com/redis/go-redis) | Redis Client |
| [zerolog](https://github.com/rs/zerolog) | Structured Logger |
| [swaggo](https://github.com/swaggo/swag) | OpenAPI Documentation |

## 📄 License

MIT License - see [LICENSE](LICENSE)

## 👤 Author

**ghmedinac** - [GitHub](https://github.com/ghmedinac)

---

⭐ If you find it useful, give the repo a star!
