# 🚀 Deployment Guide - Telegram API

## 📋 Prerequisites

- Docker installed
- Docker Compose (without dash: `docker compose`)
- Docker Hub account (optional, for push)

## ⚡ Quick Deploy

### 1. Set up environment variables

```bash
export JWT_SECRET="your_jwt_secret_at_least_32_characters"
export ENCRYPTION_KEY="exactly_32_characters_key!!"
```

### 2. Run deploy

```bash
./deploy-dev.sh [version]

# Examples:
./deploy-dev.sh           # Uses version 0.1.0 by default
./deploy-dev.sh 0.2.0     # Specifies version
```

## 🎯 Development Modes

The script will ask you what you're developing:

### 1️⃣ Backend (Go API)
- Builds only the backend image
- Deploys: PostgreSQL + Redis + Backend
- Port: 7789

```bash
./deploy-dev.sh
# Select: 1) Backend
```

### 2️⃣ Frontend (React)
- Builds only the frontend image
- Deploys: PostgreSQL + Redis + Backend + Frontend
- Port: 3000

```bash
./deploy-dev.sh
# Select: 2) Frontend
```

### 3️⃣ Full Stack (Both)
- Builds backend and frontend
- Deploys entire stack
- Ports: 3000 (frontend), 7789 (backend)

```bash
./deploy-dev.sh
# Select: 3) Both
```

### 4️⃣ Infrastructure Only
- Doesn't build images
- Deploys only PostgreSQL + Redis
- Useful for local development without Docker

```bash
./deploy-dev.sh
# Select: 4) Infrastructure only
```

## 🔄 Workflow

### Backend Development

```bash
# 1. Modify backend code
vim cmd/api/main.go

# 2. Deploy
./deploy-dev.sh 0.1.1
# Select: 1) Backend

# 3. View logs
docker compose logs -f api
```

### Frontend Development

```bash
# 1. Modify frontend code
vim frontend/src/pages/dashboard/DashboardPage.tsx

# 2. Deploy
./deploy-dev.sh 0.1.1
# Select: 2) Frontend

# 3. View logs
docker compose logs -f frontend
```

### Full Stack Development

```bash
# 1. Modify backend and frontend
vim cmd/api/main.go
vim frontend/src/App.tsx

# 2. Deploy both
./deploy-dev.sh 0.1.2
# Select: 3) Both

# 3. View all logs
docker compose logs -f
```

## 🏷️ Versioning

The script uses **semantic versioning**:

```bash
# Initial development
./deploy-dev.sh 0.1.0

# Bug fixes
./deploy-dev.sh 0.1.1
./deploy-dev.sh 0.1.2

# New features
./deploy-dev.sh 0.2.0
./deploy-dev.sh 0.3.0

# Stable version
./deploy-dev.sh 1.0.0
```

### Images on Docker Hub

The script creates tags:
- `ghmedinac/telegram-api:latest`
- `ghmedinac/telegram-api:0.1.0`
- `ghmedinac/telegram-frontend:latest`
- `ghmedinac/telegram-frontend:0.1.0`

## 📦 What the script does

1. **Detects development mode** (backend/frontend/fullstack/infra)
2. **Verifies required environment variables**
3. **Stops old services** based on mode
4. **Builds only necessary images**
   - Backend: `docker compose build --no-cache api`
   - Frontend: `docker compose build --no-cache frontend`
5. **Creates version tags**
   - `latest` and specific `version`
6. **Asks if you want to push** to Docker Hub
7. **Deploys only necessary services**
8. **Shows logs and status**

## 🛠️ Docker Compose Commands

### View services
```bash
docker compose ps
```

### View logs
```bash
# All
docker compose logs -f

# Single service
docker compose logs -f api
docker compose logs -f frontend
docker compose logs -f postgres
docker compose logs -f redis
```

### Restart services
```bash
# All
docker compose restart

# Single service
docker compose restart api
docker compose restart frontend
```

### Stop
```bash
# Stop without deleting
docker compose stop

# Stop and remove containers
docker compose down

# Stop and remove EVERYTHING (⚠️ including volumes)
docker compose down -v
```

### Manual rebuild
```bash
# Backend
docker compose build --no-cache api

# Frontend
docker compose build --no-cache frontend

# Both
docker compose build --no-cache
```

## 🐛 Troubleshooting

### Error: "docker-compose: command not found"
Use `docker compose` (without dash):
```bash
docker compose --version
```

### Frontend doesn't update
```bash
# Rebuild without cache
docker compose build --no-cache frontend
docker compose up -d frontend

# Clear browser cache
Ctrl + Shift + R
```

### Backend doesn't connect to DB
```bash
# Check PostgreSQL
docker compose logs postgres

# Check environment variables
docker compose exec api env | grep DB_URL
```

### Clean everything and start from scratch
```bash
# Stop and remove EVERYTHING
docker compose down -v

# Clean old images
docker image prune -a

# Redeploy
./deploy-dev.sh
```

## 📊 Monitoring

### Container status
```bash
docker compose ps
```

### Resources used
```bash
docker stats
```

### Inspect a container
```bash
docker compose exec api sh
docker compose exec frontend sh
```

### View healthchecks
```bash
docker inspect telegram_api_app | grep -A 10 Health
docker inspect telegram_frontend | grep -A 10 Health
```

## 🔐 Security

### Sensitive variables

**NEVER** commit these variables:
```bash
JWT_SECRET
ENCRYPTION_KEY
```

Use `.env` file (already in `.gitignore`):
```bash
# .env
JWT_SECRET=your_super_secure_secret_at_least_32_chars
ENCRYPTION_KEY=exactly_32_characters_key!!
```

Load automatically:
```bash
source .env
./deploy-dev.sh
```

## 🚀 Production Deploy

### 1. Local build
```bash
./deploy-dev.sh 1.0.0
# Select: 3) Both
# Push: y (yes)
```

### 2. On production server
```bash
# Pull images
docker pull ghmedinac/telegram-api:1.0.0
docker pull ghmedinac/telegram-frontend:1.0.0

# Configure variables
export JWT_SECRET="..."
export ENCRYPTION_KEY="..."

# Deploy
docker compose up -d
```

## 📝 Notes

- The script uses `docker compose` (without dash)
- Only rebuilds what you're developing
- Handles versioning automatically
- Asks before pushing to Docker Hub
- Shows relevant logs based on mode
- Friendly colors and formatting in terminal

## 🎯 Recommended Workflow

```bash
# 1. Local development
./deploy-dev.sh 0.1.0
# Mode: depending on what you're modifying

# 2. Testing
# Test the application

# 3. Increment version
./deploy-dev.sh 0.1.1

# 4. Push to Docker Hub when ready
# The script will ask: y/N

# 5. Repeat until stable version
./deploy-dev.sh 1.0.0
```

Happy coding! 🎉
