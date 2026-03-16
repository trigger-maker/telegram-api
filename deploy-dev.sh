#!/bin/bash
set -e

# ==================== CONFIG ====================
VERSION="${1:-0.1.0}"
BACKEND_IMAGE="ghmedinac/telegram-api"
FRONTEND_IMAGE="ghmedinac/telegram-frontend"

# Colores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# ==================== BANNER ====================
clear
echo -e "${CYAN}"
echo "╔════════════════════════════════════════════════╗"
echo "║     🚀 Telegram API - Smart Deployment       ║"
echo "║              Version: ${VERSION}                  ║"
echo "╚════════════════════════════════════════════════╝"
echo -e "${NC}"

# ==================== SELECCIÓN DE DESARROLLO ====================
echo -e "${BLUE}📋 ¿Qué estás desarrollando?${NC}"
echo ""
echo "  1) Backend (Go API)"
echo "  2) Frontend (React)"
echo "  3) Ambos (Full Stack)"
echo "  4) Solo infraestructura (PostgreSQL + Redis)"
echo ""
read -p "Selecciona una opción [1-4]: " DEV_OPTION

case $DEV_OPTION in
    1)
        DEV_MODE="backend"
        echo -e "${GREEN}✓ Modo: Desarrollo Backend${NC}"
        ;;
    2)
        DEV_MODE="frontend"
        echo -e "${GREEN}✓ Modo: Desarrollo Frontend${NC}"
        ;;
    3)
        DEV_MODE="fullstack"
        echo -e "${GREEN}✓ Modo: Full Stack${NC}"
        ;;
    4)
        DEV_MODE="infra"
        echo -e "${GREEN}✓ Modo: Solo Infraestructura${NC}"
        ;;
    *)
        echo -e "${RED}✗ Opción inválida${NC}"
        exit 1
        ;;
esac

echo ""

# ==================== VERIFICAR VARIABLES ====================
echo -e "${BLUE}🔍 Verificando variables de entorno...${NC}"

if [ -z "$JWT_SECRET" ]; then
    echo -e "${YELLOW}⚠️  JWT_SECRET no definido, usando valor por defecto${NC}"
    export JWT_SECRET="tu_jwt_secret_32_caracteres_min!"
fi

if [ -z "$ENCRYPTION_KEY" ]; then
    echo -e "${YELLOW}⚠️  ENCRYPTION_KEY no definido, usando valor por defecto${NC}"
    export ENCRYPTION_KEY="clave_32_caracteres_exactos!!"
fi

# ==================== FUNCIONES ====================
build_backend() {
    echo -e "${CYAN}🔨 Construyendo Backend (Go)...${NC}"
    docker compose build --no-cache api
    docker tag ghmedinac/telegram-api:latest ${BACKEND_IMAGE}:${VERSION}
    echo -e "${GREEN}✓ Backend construido: ${VERSION}${NC}"
}

build_frontend() {
    echo -e "${CYAN}🔨 Construyendo Frontend (React)...${NC}"
    docker compose build --no-cache frontend
    docker tag ghmedinac/telegram-frontend:latest ${FRONTEND_IMAGE}:${VERSION}
    echo -e "${GREEN}✓ Frontend construido: ${VERSION}${NC}"
}

push_images() {
    read -p "¿Deseas subir las imágenes a Docker Hub? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${CYAN}📤 Subiendo a Docker Hub...${NC}"

        if [ "$DEV_MODE" = "backend" ] || [ "$DEV_MODE" = "fullstack" ]; then
            docker push ${BACKEND_IMAGE}:latest
            docker push ${BACKEND_IMAGE}:${VERSION}
            echo -e "${GREEN}✓ Backend subido${NC}"
        fi

        if [ "$DEV_MODE" = "frontend" ] || [ "$DEV_MODE" = "fullstack" ]; then
            docker push ${FRONTEND_IMAGE}:latest
            docker push ${FRONTEND_IMAGE}:${VERSION}
            echo -e "${GREEN}✓ Frontend subido${NC}"
        fi
    fi
}

deploy_services() {
    echo -e "${CYAN}▶️  Desplegando servicios...${NC}"

    case $DEV_MODE in
        backend)
            echo -e "${YELLOW}  → PostgreSQL, Redis, Backend${NC}"
            docker compose up -d postgres redis
            sleep 3
            docker compose up -d api
            ;;
        frontend)
            echo -e "${YELLOW}  → PostgreSQL, Redis, Backend, Frontend${NC}"
            docker compose up -d postgres redis api
            sleep 5
            docker compose up -d frontend
            ;;
        fullstack)
            echo -e "${YELLOW}  → Stack completo${NC}"
            docker compose up -d
            ;;
        infra)
            echo -e "${YELLOW}  → PostgreSQL, Redis${NC}"
            docker compose up -d postgres redis
            ;;
    esac
}

show_logs() {
    echo ""
    echo -e "${BLUE}📋 Logs de servicios:${NC}"

    case $DEV_MODE in
        backend)
            docker compose logs --tail=20 api
            ;;
        frontend)
            docker compose logs --tail=20 frontend
            ;;
        fullstack)
            docker compose logs --tail=10
            ;;
        infra)
            docker compose logs --tail=10 postgres redis
            ;;
    esac
}

# ==================== DETENER SERVICIOS ANTIGUOS ====================
echo -e "${YELLOW}⏹️  Deteniendo servicios anteriores...${NC}"

case $DEV_MODE in
    backend)
        docker compose stop api 2>/dev/null || true
        ;;
    frontend)
        docker compose stop frontend 2>/dev/null || true
        ;;
    fullstack)
        docker compose down 2>/dev/null || true
        ;;
    infra)
        docker compose stop postgres redis 2>/dev/null || true
        ;;
esac

echo ""

# ==================== BUILD ====================
if [ "$DEV_MODE" != "infra" ]; then
    echo -e "${BLUE}🏗️  Construyendo imágenes...${NC}"
    echo ""

    if [ "$DEV_MODE" = "backend" ] || [ "$DEV_MODE" = "fullstack" ]; then
        build_backend
    fi

    if [ "$DEV_MODE" = "frontend" ] || [ "$DEV_MODE" = "fullstack" ]; then
        build_frontend
    fi

    echo ""

    # ==================== PUSH ====================
    push_images
fi

echo ""

# ==================== DEPLOY ====================
deploy_services

# ==================== ESPERAR ====================
echo ""
echo -e "${BLUE}⏳ Esperando que los servicios estén listos...${NC}"
sleep 8

# ==================== VERIFICAR ====================
echo ""
echo -e "${GREEN}📊 Estado de servicios:${NC}"
docker compose ps

# ==================== LOGS ====================
show_logs

# ==================== RESUMEN ====================
echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║           ✅ DEPLOYMENT EXITOSO ✅            ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${CYAN}📍 Servicios activos (Modo: ${DEV_MODE}):${NC}"
echo ""

if [ "$DEV_MODE" = "infra" ] || [ "$DEV_MODE" = "backend" ] || [ "$DEV_MODE" = "frontend" ] || [ "$DEV_MODE" = "fullstack" ]; then
    echo -e "   ${GREEN}PostgreSQL:${NC}  localhost:5649"
    echo -e "   ${GREEN}Redis:${NC}       localhost:7954"
fi

if [ "$DEV_MODE" = "backend" ] || [ "$DEV_MODE" = "fullstack" ]; then
    echo -e "   ${GREEN}Backend:${NC}     http://localhost:7789"
    echo -e "                ${BACKEND_IMAGE}:${VERSION}"
fi

if [ "$DEV_MODE" = "frontend" ] || [ "$DEV_MODE" = "fullstack" ]; then
    echo -e "   ${GREEN}Frontend:${NC}    http://localhost:3000"
    echo -e "                ${FRONTEND_IMAGE}:${VERSION}"
fi

echo ""
echo -e "${CYAN}📊 Commandos útiles:${NC}"
echo ""
echo -e "   ${YELLOW}Ver logs:${NC}        docker compose logs -f"

case $DEV_MODE in
    backend)
        echo -e "   ${YELLOW}Logs backend:${NC}    docker compose logs -f api"
        echo -e "   ${YELLOW}Reiniciar:${NC}       docker compose restart api"
        ;;
    frontend)
        echo -e "   ${YELLOW}Logs frontend:${NC}   docker compose logs -f frontend"
        echo -e "   ${YELLOW}Reiniciar:${NC}       docker compose restart frontend"
        ;;
    fullstack)
        echo -e "   ${YELLOW}Logs backend:${NC}    docker compose logs -f api"
        echo -e "   ${YELLOW}Logs frontend:${NC}   docker compose logs -f frontend"
        echo -e "   ${YELLOW}Reiniciar todo:${NC}  docker compose restart"
        ;;
esac

echo -e "   ${YELLOW}Detener:${NC}         docker compose down"
echo -e "   ${YELLOW}Estado:${NC}          docker compose ps"
echo ""
echo -e "${GREEN}🎉 Happy coding!${NC}"
echo ""
