# 🐳 Docker Deployment Guide

Complete guide to deploy the full Telegram API stack with Docker.

## 📦 Full Stack

```
┌─────────────────────────────────────────┐
│         Frontend (React + Nginx)        │
│           http://localhost:3000         │
└────────────────┬────────────────────────┘
                  │
┌────────────────▼────────────────────────┐
│          Backend API (Go)               │
│           http://localhost:7789         │
└────────────┬──────────────┬─────────────┘
              │              │
┌────────────▼────┐  ┌──────▼─────────────┐
│   PostgreSQL    │  │      Redis         │
│  localhost:5649 │  │  localhost:7954    │
└─────────────────┘  └────────────────────┘
```

## 🚀 Quick Deploy

### 1. Set up environment variables

```bash
# Export required variables
export JWT_SECRET="your_jwt_secret_at_least_32_secure_characters"
export ENCRYPTION_KEY="exactly_32_characters_key!!"
```

### 2. Deploy full stack

```bash
# Using automated script
./docker-deploy.sh

# Or manually
docker-compose up -d --build
```

## 📋 Services

### PostgreSQL
- **Port:** 5649 (external) → 5432 (internal)
- **User:** admin
- **Password:** password123
- **Database:** telegram_db
- **Healthcheck:** `pg_isready`

### Redis
- **Port:** 7954 (external) → 6379 (internal)
- **Persistence:** AOF enabled
- **Healthcheck:** `redis-cli ping`

### Backend API (Go)
- **Port:** 7789 (external) → 8080 (internal)
- **Image:** `ghmedinac/telegram-api:latest`
- **Healthcheck:** `wget http://localhost:8080/health`
- **Depends on:** PostgreSQL, Redis

### Frontend (React)
- **Port:** 3000 (external) → 80 (internal)
- **Image:** `ghmedinac/telegram-frontend:latest`
- **Server:** Nginx
- **Healthcheck:** `wget http://localhost/`
- **Depends on:** Backend API

## 🛠️ Useful Commands

### View logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f frontend
docker-compose logs -f api
docker-compose logs -f postgres
docker-compose logs -f redis
```

### Check status
```bash
docker-compose ps
```

### Restart services
```bash
# All
docker-compose restart

# Specific one
docker-compose restart frontend
```

### Stop services
```bash
# Stop without deleting volumes
docker-compose down

# Stop and delete volumes (⚠️ deletes data)
docker-compose down -v
```

### Rebuild images
```bash
# Without cache
docker-compose build --no-cache

# Single service
docker-compose build --no-cache frontend
```

### Execute commands in containers
```bash
# Shell in frontend
docker exec -it telegram_frontend sh

# Shell in backend
docker exec -it telegram_api_app sh

# Connect to PostgreSQL
docker exec -it tg_postgres psql -U admin -d telegram_db

# Connect to Redis
docker exec -it tg_redis redis-cli
```

## 🔒 Environment Variables

### Backend API
```bash
API_PORT=8080
API_ENV=production
LOG_LEVEL=info
DB_URL=postgres://admin:password123@tg_postgres:5432/telegram_db?sslmode=disable
REDIS_ADDR=tg_redis:6379
REDIS_PASSWORD=
JWT_SECRET=${JWT_SECRET}
JWT_EXPIRY=24h
ENCRYPTION_KEY=${ENCRYPTION_KEY}
```

## 📊 Monitoring

### Check service health
```bash
# View healthchecks
docker-compose ps

# Inspect a container
docker inspect telegram_frontend | grep -A 10 Health
```

### Resources used
```bash
docker stats
```

## 🐛 Troubleshooting

### Frontend doesn't load
```bash
# Check logs
docker-compose logs frontend

# Verify backend is available
docker exec -it telegram_frontend wget -O- http://api:8080/health
```

### Backend doesn't connect to DB
```bash
# Check PostgreSQL
docker exec -it tg_postgres pg_isready -U admin -d telegram_db

# Check backend logs
docker-compose logs api
```

### Redis connection refused
```bash
# Check Redis
docker exec -it tg_redis redis-cli ping

# Check logs
docker-compose logs redis
```

### Rebuild from scratch
```bash
# Stop everything and clean
docker-compose down -v
docker system prune -a

# Rebuild
./docker-deploy.sh
```

## 🚀 Production Deploy

### 1. Upload images to Docker Hub

```bash
# Login to Docker Hub
docker login

# Build and push
docker-compose build
docker-compose push
```

### 2. On production server

```bash
# Clone repository
git clone https://github.com/your-user/telegram-api.git
cd telegram-api

# Configure variables
export JWT_SECRET="..."
export ENCRYPTION_KEY="..."

# Deploy
./docker-deploy.sh
```

## 📁 Persistent Volumes

Data is persisted in Docker volumes:

- `postgres_data` - PostgreSQL database
- `redis_data` - Redis data (AOF)

### Backup
```bash
# PostgreSQL
docker exec tg_postgres pg_dump -U admin telegram_db > backup.sql

# Restore
docker exec -i tg_postgres psql -U admin telegram_db < backup.sql
```

## 🔄 Update Versions

```bash
# Pull new images
docker-compose pull

# Recreate containers
docker-compose up -d --force-recreate

# Or use the script
./docker-deploy.sh 0.2.0
```

## 📝 Notes

- Frontend proxies to backend through Nginx (see `frontend/nginx.conf`)
- Healthchecks ensure services depend correctly
- Images use multi-stage builds to optimize size
- Frontend uses Node 22 Alpine + Nginx Alpine (very lightweight)
