#!/bin/bash

# simpo Development Scripts
# Usage: ./scripts/dev.sh [command]

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project root
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && cd "$PROJECT_ROOT" && pwd)"

echo -e "${BLUE}simpo Development Environment${NC}"
echo -e "${BLUE}=====================================${NC}\n"

# Function to show usage
show_usage() {
    echo "Usage: ./scripts/dev.sh [command]"
    echo ""
    echo "Commands:"
    echo "  backend     Start backend API server"
    echo "  mobile      Start mobile Metro bundler"
    echo "  web         Start web admin dashboard"
    echo "  all         Start all services"
    echo "  install     Install all dependencies"
    echo "  clean       Clean build artifacts"
    echo "  help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./scripts/dev.sh backend"
    echo "  ./scripts/dev.sh all"
}

# Function to start backend
start_backend() {
    echo -e "${GREEN}▶ Starting Backend API...${NC}"
    cd "$PROJECT_ROOT/apps/backend"

    if [ ! -f .env ]; then
        echo -e "${YELLOW}⚠ .env file not found. Copying from .env.example...${NC}"
        cp .env.example .env
        echo -e "${YELLOW}⚠ Please edit .env with your configuration${NC}"
    fi

    # Export environment variables
    export $(cat .env 2>/dev/null | grep -v '^#' | xargs)

    echo -e "${GREEN}✓ Backend starting on http://localhost:8080${NC}"
    go run cmd/server/main.go
}

# Function to start mobile
start_mobile() {
    echo -e "${GREEN}▶ Starting Mobile Metro Bundler...${NC}"
    cd "$PROJECT_ROOT/apps/mobile"

    if [ ! -d node_modules ]; then
        echo -e "${YELLOW}⚠ node_modules not found. Installing dependencies...${NC}"
        yarn install
    fi

    echo -e "${GREEN}✓ Metro bundler starting on http://localhost:8081${NC}"
    yarn start
}

# Function to start web
start_web() {
    echo -e "${GREEN}▶ Starting Web Admin Dashboard...${NC}"
    cd "$PROJECT_ROOT/apps/web"

    if [ ! -d node_modules ]; then
        echo -e "${YELLOW}⚠ node_modules not found. Installing dependencies...${NC}"
        yarn install
    fi

    if [ ! -f .env.local ]; then
        echo -e "${YELLOW}⚠ .env.local not found. Creating from template...${NC}"
        cat > .env.local << EOF
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
EOF
    fi

    echo -e "${GREEN}✓ Web dashboard starting on http://localhost:3000${NC}"
    yarn dev
}

# Function to start all services
start_all() {
    echo -e "${GREEN}▶ Starting All Services...${NC}"

    # Start backend in background
    cd "$PROJECT_ROOT/apps/backend"
    if [ -f .env ]; then
        export $(cat .env | grep -v '^#' | xargs)
        nohup go run cmd/server/main.go > /tmp/simpo_backend.log 2>&1 &
        echo -e "${GREEN}✓ Backend started in background (port 8080)${NC}"
    fi

    # Start mobile in new terminal (macOS)
    if [[ "$OSTYPE" == "darwin"* ]]; then
        osascript -e 'tell application "Terminal" to do script "cd '"$PROJECT_ROOT"' && cd apps/mobile && yarn start" activate' 2>/dev/null || true
    fi

    echo -e "${GREEN}✓ Services starting...${NC}"
    echo -e ""
    echo -e "${BLUE}Access Points:${NC}"
    echo -e "  • Backend API:  http://localhost:8080"
    echo -e "  • Mobile Metro: http://localhost:8081"
    echo -e "  • Web Dashboard: http://localhost:3000"
    echo -e "  • API Docs:     http://localhost:8080/swagger/index.html"
}

# Function to install all dependencies
install_deps() {
    echo -e "${GREEN}▶ Installing All Dependencies...${NC}"

    # Backend dependencies
    echo -e "${BLUE}Installing backend dependencies...${NC}"
    cd "$PROJECT_ROOT/apps/backend"
    go mod download

    # Mobile dependencies
    echo -e "${BLUE}Installing mobile dependencies...${NC}"
    cd "$PROJECT_ROOT/apps/mobile"
    yarn install

    # Web dependencies (if exists)
    if [ -d "$PROJECT_ROOT/apps/web" ]; then
        echo -e "${BLUE}Installing web dependencies...${NC}"
        cd "$PROJECT_ROOT/apps/web"
        yarn install
    fi

    echo -e "${GREEN}✓ All dependencies installed${NC}"
}

# Function to clean build artifacts
clean_artifacts() {
    echo -e "${GREEN}▶ Cleaning Build Artifacts...${NC}"

    # Backend clean
    cd "$PROJECT_ROOT/apps/backend"
    rm -rf tmp/ bin/

    # Mobile clean
    cd "$PROJECT_ROOT/apps/mobile"
    rm -rf node_modules/.cache/

    # Web clean (if exists)
    if [ -d "$PROJECT_ROOT/apps/web" ]; then
        cd "$PROJECT_ROOT/apps/web"
        rm -rf .next/ node_modules/.cache/
    fi

    echo -e "${GREEN}✓ Build artifacts cleaned${NC}"
}

# Main script logic
case "${1:-help}" in
    backend)
        start_backend
        ;;
    mobile)
        start_mobile
        ;;
    web)
        start_web
        ;;
    all)
        start_all
        ;;
    install)
        install_deps
        ;;
    clean)
        clean_artifacts
        ;;
    help|*)
        show_usage
        ;;
esac
