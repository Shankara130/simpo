# simpo - Pharmacy Management System

**Cost-effective Pharmacy Management System for Indonesian SME Pharmacies**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev/)
[![React Native](https://img.shields.io/badge/React_Native-0.73+-61DAFB?logo=react)](https://reactnative.dev/)
[![Next.js](https://img.shields.io/badge/Next.js-14+-000000?logo=next.js)](https://nextjs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14+-336791?logo=postgresql)](https://www.postgresql.org/)

---

## 🏥 About simpo

simpo adalah sistem manajemen apotek **self-hosted** yang dirancang khusus untuk apotek SME di Indonesia. Sistem ini menyediakan:

- 💊 **Point of Sale (POS)** - Aplikasi mobile Android untuk kasir
- 📊 **Admin Dashboard** - Web dashboard untuk pemilik apotek  
- 📦 **Inventory Management** - Manajemen stok real-time dengan pelacakan expiry
- 💰 **Financial Reporting** - Laporan penjualan harian dan laba rugi
- 👥 **Supplier Management** - Manajemen supplier dan faktur pembelian
- 🔐 **Role-Based Access** - 3 roles: Admin, Owner, Cashier
- 🌐 **Offline Mode** - Sinkronisasi otomatis saat koneksi tersedia

---

## 📁 Monorepo Structure

```
simpo/
├── apps/                    # Application modules
│   ├── backend/             # Go REST API (Gin + GORM)
│   ├── mobile/              # React Native CLI (POS Android)
│   └── web/                 # Next.js (Admin Dashboard)
│
├── packages/                # Shared packages (future)
│   ├── shared-types/        # TypeScript types
│   └── ui-components/       # Shared UI components
│
├── docs/                    # Documentation
│   ├── planning/            # PRD, Architecture, HLD, LLD
│   ├── implementation/      # Stories, Sprint status
│   └── api/                 # API documentation
│
├── scripts/                 # Utility scripts
│   ├── setup.sh             # Initial setup script
│   └── dev.sh               # Development commands
│
├── .github/                 # GitHub configuration
│   └── workflows/           # CI/CD workflows
│
├── docker-compose.yml       # Local development
├── .gitignore              # Root gitignore
└── README.md               # This file
```

---

## 🚀 Quick Start

### Prerequisites

- **Backend**: Go 1.21+, PostgreSQL 14+
- **Mobile**: Node.js 18+, React Native CLI, Android Studio
- **Web**: Node.js 18+, npm/yarn
- **Docker** (optional): Docker, Docker Compose

### Initial Setup

```bash
# Clone repository
git clone <repository-url>
cd simpo

# Run setup script
chmod +x scripts/setup.sh
./scripts/setup.sh
```

The setup script will:
- ✓ Check prerequisites (Go, Node.js, PostgreSQL)
- ✓ Install backend dependencies
- ✓ Install mobile dependencies  
- ✓ Create database
- ✓ Generate secure configuration
- ✓ Setup environment files

### Development

#### Option 1: Using Development Scripts

```bash
# Start all services
./scripts/dev.sh all

# Start individual services
./scripts/dev.sh backend    # Backend API (port 8080)
./scripts/dev.sh mobile     # Mobile Metro (port 8081)
./scripts/dev.sh web        # Web dashboard (port 3000)
```

#### Option 2: Manual Startup

**Start Backend (Go API):**
```bash
cd apps/backend
cp .env.example .env  # If not exists
# Edit .env with your configuration
export $(cat .env | grep -v '^#' | xargs)
go run cmd/server/main.go
```
Backend: http://localhost:8080

**Start Mobile (React Native POS):**
```bash
cd apps/mobile
yarn install      # If not installed
yarn start        # Metro bundler (port 8081)
yarn android      # Run on Android emulator/device
```
Metro: http://localhost:8081

**Start Web (Next.js Admin):**
```bash
cd apps/web
yarn install      # If not installed
yarn dev
```
Web: http://localhost:3000

#### Option 3: Docker Compose

```bash
# Start all services in Docker
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

---

## 📚 Documentation

### Planning Documents
- [Product Requirements Document](docs/planning/planning-artifacts/prd.md)
- [Architecture Decisions](docs/planning/planning-artifacts/architecture.md)
- [High Level Design (HLD)](docs/planning/planning-artifacts/HLD-simpo-pharmacy-management.md)
- [Low Level Design (LLD)](docs/planning/planning-artifacts/LLD-simpo-pharmacy-management.md)

### Implementation
- [Sprint Status](docs/implementation/implementation-artifacts/sprint-status.yaml)
- [Story Files](docs/implementation/implementation-artifacts/)
- [API Documentation](http://localhost:8080/swagger/index.html)

---

## 🏗️ Architecture

### Technology Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Backend API** | Go 1.24 + Gin | RESTful API server |
| **Database** | PostgreSQL 14 + GORM | Primary data storage |
| **Mobile POS** | React Native CLI 0.73+ | Android cashier app |
| **Web Admin** | Next.js 14+ | Owner dashboard |
| **Authentication** | JWT (8-hour expiry) | Stateless auth |
| **Cache** | Redis 7+ | Session management |
| **Documentation** | Swagger/OpenAPI | API docs |

### System Architecture

```
┌─────────────────┐    ┌─────────────────┐
│   Mobile POS    │    │   Web Admin     │
│  (React Native) │    │    (Next.js)    │
└────────┬────────┘    └────────┬────────┘
         │                     │
         │    REST API         │
         └─────────┬───────────┘
                   │
         ┌─────────▼───────────┐
         │   Backend API       │
         │   (Go + Gin)        │
         └─────────┬───────────┘
                   │
         ┌─────────▼───────────┐
         │  PostgreSQL + Redis │
         └─────────────────────┘
```

---

## 🔐 Security

- **Authentication**: JWT dengan 8-hour session expiration
- **Authorization**: Role-Based Access Control (RBAC)
- **Password Hashing**: bcrypt dengan cost factor 12
- **API Security**: Rate limiting (100 req/min), CORS
- **Audit Trail**: Append-only logging untuk compliance Badan POM

---

## 📱 Features

### Point of Sale (Mobile)
- 📱 Barcode scanning integration
- 🛒 Cart management
- 💳 Multiple payment methods
- 🧾 Thermal printer support (ESC/POS)
- 📡 Offline mode dengan sync queue
- ⏱️ Sub-30 second transaction processing

### Inventory Management
- 📦 Real-time stock visibility
- ⚠️ Low stock notifications
- 📅 Expiry date alerts
- 🔢 Batch/lot number tracking
- 📊 Multi-branch support

### Financial Reporting
- 📈 Daily sales summaries
- 💹 Profit & Loss statements
- 📄 Export functionality (CSV/PDF)
- 📝 Append-only audit trail
- 🏪 Supplier aging reports

---

## 🛠️ Development

### Code Standards

- **Backend**: Go standard formatting + Air hot-reload
- **Mobile**: TypeScript strict mode + ESLint
- **Web**: TypeScript + Prettier + ESLint

### Testing

- **Backend**: Unit tests dengan Go testing framework
- **Mobile**: Jest + React Native Testing Library
- **Web**: Jest + React Testing Library

### Git Workflow

```bash
# Feature branch workflow
git checkout -b feature/story-1-2-user-auth
git commit -m "feat(auth): implement user authentication"
git push origin feature/story-1-2-user-auth
```

Commit message format: `type(scope): description`

---

## 📦 Deployment

### Self-Hosted Deployment

Minimum requirements:
- **CPU**: 2 cores
- **RAM**: 4GB
- **Storage**: 20GB
- **OS**: Linux/macOS/Windows

```bash
# Build and deploy
docker-compose -f docker-compose.prod.yml up -d
```

### Backup Strategy

- Automated daily backups
- Database dumps: `/backups/db/`
- Config backups: `/backups/config/`
- Retention: 30 days

---

## 🔧 Configuration

### Environment Variables

**Backend** (`apps/backend/.env`):
```bash
JWT_SECRET=your-secret-key
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_NAME=simpo_db
SERVER_PORT=8080
```

**Mobile** (`apps/mobile/.env`):
```bash
API_BASE_URL=http://localhost:8080/api/v1
```

**Web** (`apps/web/.env.local`):
```bash
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

---

## 🤝 Contributing

1. Fork repository
2. Create feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'feat: Add AmazingFeature'`)
4. Push to branch (`git push origin feature/AmazingFeature`)
5. Open Pull Request

---

## 📄 License

MIT License - see LICENSE file for details

---

## 👥 Team

- **Product Owner**: Shankara
- **Development**: simpo Development Team

---

## 📞 Support

- **Documentation**: [docs/](docs/)
- **Issues**: GitHub Issues
- **Email**: support@simpo.pharmacy

---

**Status**: 🚀 Active Development  
**Version**: 0.1.0-alpha  
**Last Updated**: 2026-05-09
