#!/bin/bash

# simpo Initial Setup Script
# This script sets up the development environment for simpo monorepo

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=====================================${NC}"
echo -e "${BLUE}  simpo Pharmacy Management System  ${NC}"
echo -e "${BLUE}  Initial Setup Script               ${NC}"
echo -e "${BLUE}=====================================${NC}\n"

# Project root
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo -e "${GREEN}📁 Project Root: $PROJECT_ROOT${NC}\n"

# Check prerequisites
echo -e "${BLUE}🔍 Checking Prerequisites...${NC}"

# Check Go
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo -e "${GREEN}✓${NC} Go installed: $GO_VERSION"
else
    echo -e "${RED}✗${NC} Go not found. Please install Go 1.21+"
    exit 1
fi

# Check Node.js
if command -v node &> /dev/null; then
    NODE_VERSION=$(node -v)
    echo -e "${GREEN}✓${NC} Node.js installed: $NODE_VERSION"
else
    echo -e "${RED}✗${NC} Node.js not found. Please install Node.js 18+"
    exit 1
fi

# Check PostgreSQL
if command -v psql &> /dev/null; then
    PG_VERSION=$(psql --version | awk '{print $3}')
    echo -e "${GREEN}✓${NC} PostgreSQL installed: $PG_VERSION"
else
    echo -e "${YELLOW}⚠${NC} PostgreSQL not found. Please install PostgreSQL 14+"
fi

# Check Docker
if command -v docker &> /dev/null; then
    DOCKER_VERSION=$(docker --version | awk '{print $3}')
    echo -e "${GREEN}✓${NC} Docker installed: $DOCKER_VERSION"
else
    echo -e "${YELLOW}⚠${NC} Docker not found. Optional but recommended."
fi

echo ""

# Setup backend
echo -e "${BLUE}🔧 Setting up Backend (Go API)...${NC}"
cd "$PROJECT_ROOT/apps/backend"

if [ ! -f .env ]; then
    echo -e "${YELLOW}⚠ Creating .env from .env.example...${NC}"
    cp .env.example .env

    # Generate secure JWT secret
    JWT_SECRET=$(openssl rand -base64 32)
    sed -i.bak "s/JWT_SECRET=.*/JWT_SECRET=$JWT_SECRET/" .env
    rm .env.bak

    echo -e "${GREEN}✓${NC} .env created with secure JWT secret"
else
    echo -e "${GREEN}✓${NC} .env already exists"
fi

# Download Go dependencies
echo -e "${YELLOW}⚠ Downloading Go dependencies...${NC}"
go mod download
echo -e "${GREEN}✓${NC} Go dependencies downloaded"

# Create database if it doesn't exist
if command -v psql &> /dev/null; then
    DB_USER=$(grep ^DATABASE_USER .env | cut -d'=' -f2)
    DB_NAME=$(grep ^DATABASE_NAME .env | cut -d'=' -f2)

    if psql -U "$DB_USER" -lqt | cut -d \| -f 1 | grep -w "$DB_NAME" > /dev/null; then
        echo -e "${GREEN}✓${NC} Database '$DB_NAME' already exists"
    else
        echo -e "${YELLOW}⚠ Creating database '$DB_NAME'...${NC}"
        createdb -U "$DB_USER" "$DB_NAME" 2>/dev/null || echo -e "${YELLOW}⚠ Database creation may have failed - please create manually${NC}"
    fi
fi

echo ""

# Setup mobile
echo -e "${BLUE}🔧 Setting up Mobile (React Native)...${NC}"
cd "$PROJECT_ROOT/apps/mobile"

if [ ! -d node_modules ]; then
    echo -e "${YELLOW}⚠ Installing mobile dependencies...${NC}"
    yarn install
    echo -e "${GREEN}✓${NC} Mobile dependencies installed"
else
    echo -e "${GREEN}✓${NC} Mobile dependencies already installed"
fi

# Create mobile .env if it doesn't exist
if [ ! -f .env ]; then
    cat > .env << EOF
# Mobile Configuration
API_BASE_URL=http://localhost:8080/api/v1
EOF
    echo -e "${GREEN}✓${NC} Mobile .env created"
fi

echo ""

# Setup web (if exists)
if [ -d "$PROJECT_ROOT/apps/web" ]; then
    echo -e "${BLUE}🔧 Setting up Web (Next.js)...${NC}"
    cd "$PROJECT_ROOT/apps/web"

    if [ ! -d node_modules ]; then
        echo -e "${YELLOW}⚠ Installing web dependencies...${NC}"
        yarn install
        echo -e "${GREEN}✓${NC} Web dependencies installed"
    else
        echo -e "${GREEN}✓${NC} Web dependencies already installed"
    fi

    # Create web .env.local if it doesn't exist
    if [ ! -f .env.local ]; then
        cat > .env.local << EOF
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
EOF
        echo -e "${GREEN}✓${NC} Web .env.local created"
    fi

    echo ""
fi

# Setup scripts
echo -e "${BLUE}🔧 Making scripts executable...${NC}"
chmod +x "$PROJECT_ROOT/scripts/"*.sh
echo -e "${GREEN}✓${NC} Scripts executable"

echo ""

# Setup documentation
echo -e "${BLUE}📚 Setting up documentation...${NC}"
cd "$PROJECT_ROOT"

# Create docs structure
mkdir -p docs/planning docs/implementation docs/api

if [ -d docs/implementation-artifacts ]; then
    echo -e "${GREEN}✓${NC} Documentation already exists"
else
    echo -e "${YELLOW}⚠ Creating documentation structure...${NC}"
fi

echo ""

# Final summary
echo -e "${GREEN}=====================================${NC}"
echo -e "${GREEN}  ✅ Setup Complete!                    ${NC}"
echo -e "${GREEN}=====================================${NC}\n"

echo -e "${BLUE}🚀 Quick Start:${NC}\n"

echo -e "1. Start Backend API:"
echo -e "   ${YELLOW}./scripts/dev.sh backend${NC}"
echo -e "   Or: ${YELLOW}cd apps/backend && go run cmd/server/main.go${NC}\n"

echo -e "2. Start Mobile (React Native):"
echo -e "   ${YELLOW}./scripts/dev.sh mobile${NC}"
echo -e "   Or: ${YELLOW}cd apps/mobile && yarn start${NC}\n"

echo -e "3. Start All Services:"
echo -e "   ${YELLOW}./scripts/dev.sh all${NC}\n"

echo -e "${BLUE}📚 Documentation:${NC}"
echo -e "  • README.md - Project overview"
echo -e "  • docs/planning/ - PRD, Architecture, HLD, LLD"
echo -e "  • docs/implementation/ - Stories, Sprint status"
echo -e "  • http://localhost:8080/swagger/index.html - API docs\n"

echo -e "${BLUE}🔧 Configuration Files:${NC}"
echo -e "  • apps/backend/.env - Backend configuration"
echo -e "  • apps/mobile/.env - Mobile configuration"
echo -e "  • apps/web/.env.local - Web configuration\n"

echo -e "${BLUE}⚠️  Next Steps:${NC}"
echo -e "  1. Review configuration in .env files"
echo -e "  2. Start PostgreSQL database"
echo -e "  3. Run database migrations: ${YELLOW}cd apps/backend && migrate up${NC}"
echo -e "  4. Start development servers"
echo -e ""

echo -e "${GREEN}Happy coding! 🎉${NC}\n"
