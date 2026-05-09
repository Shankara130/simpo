# Story 1.1: Initialize Backend Project with GRAB Boilerplate

Status: done

**Epic:** 1 - Authentication & User Management
**Priority:** Foundation (First Story)
**Story Type:** Project Initialization

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

---

## Story

**As a** Development Team,
**I want to** set up the backend project using the GRAB (Go REST API Boilerplate) starter template,
**So that** we have a solid foundation with Clean Architecture, JWT authentication, and RBAC already implemented.

---

## Acceptance Criteria

1. **AC1:** Project structure includes Clean Architecture layers (Handler, Service, Repository)
   - Each layer is properly separated with clear responsibilities
   - Dependency injection pattern is implemented
   - Code organization follows feature-based structure where applicable

2. **AC2:** JWT authentication middleware is configured
   - JWT token generation is working
   - Token validation middleware is in place
   - 8-hour session expiration is configured (NFR-SEC-002)
   - Secret key is configurable via environment variable

3. **AC3:** Role-Based Access Control (RBAC) is set up
   - Three roles are defined: Admin, Owner, Cashier
   - Role permissions are configured
   - Authorization middleware protects endpoints
   - Role assignment is part of user model

4. **AC4:** PostgreSQL connection is configured with GORM
   - Database connection is established via .env configuration
   - GORM is integrated and configured
   - Connection pooling is set up appropriately
   - Database migration system is ready (golang-migrate)

5. **AC5:** Swagger documentation is ready for customization
   - Swagger/OpenAPI documentation is accessible at `/api/docs` or similar endpoint
   - Basic API documentation structure exists
   - swaggo is integrated for auto-generation from annotations

6. **AC6:** Development environment is fully operational
   - Project can be run locally with `go run main.go`
   - Hot reload works with Air for development
   - All dependencies are downloaded successfully
   - Basic health check endpoint responds

7. **AC7:** Project configuration follows simpo standards
   - .env.example is provided with documented variables
   - Docker support is configured for local development
   - Logging is configured (structured logging with zap)
   - Project is ready for monorepo structure (backend/ directory)

---

## Tasks / Subtasks

- [x] **Task 1: Clone and Configure GRAB Boilerplate** (AC: 1, 7)
  - [x] Clone GRAB repository from `https://github.com/vahiiiid/go-rest-api-boilerplate.git` into `backend/` directory
  - [x] Copy `.env.example` to `.env` and configure database variables
  - [x] Update project name and descriptions in README and configuration files
  - [x] Verify Go version compatibility (Go 1.21+ required)
  - [x] Download dependencies with `go mod download`

- [x] **Task 2: Verify Clean Architecture Structure** (AC: 1)
  - [x] Confirm Handler layer exists with proper HTTP handlers
  - [x] Confirm Service layer exists with business logic
  - [x] Confirm Repository layer exists with data access
  - [x] Verify dependency injection is properly implemented
  - [x] Document the layer structure in project README

- [x] **Task 3: Configure JWT Authentication** (AC: 2)
  - [x] Set JWT secret key in .env configuration
  - [x] Verify 8-hour session expiration is configured
  - [x] Test token generation with a sample login
  - [x] Verify token validation middleware is working
  - [x] Document JWT configuration for team reference

- [x] **Task 4: Configure Role-Based Access Control (RBAC)** (AC: 3)
  - [x] Verify three roles exist: Admin, Owner, Cashier
  - [x] Check role permissions are defined in code
  - [x] Test authorization middleware with different roles
  - [x] Document RBAC rules and permission matrix

- [x] **Task 5: Set Up PostgreSQL and GORM** (AC: 4)
  - [x] Configure PostgreSQL connection in .env
  - [x] Verify GORM is properly initialized
  - [x] Test database connection with health check
  - [x] Configure connection pooling settings
  - [x] Set up golang-migrate for future migrations

- [x] **Task 6: Verify Swagger Documentation** (AC: 5)
  - [x] Access Swagger UI at configured endpoint
  - [x] Verify swaggo is integrated in the project
  - [x] Test API documentation generation from Go annotations
  - [x] Document how to update Swagger docs for team

- [x] **Task 7: Configure Development Environment** (AC: 6)
  - [x] Run `go run main.go` and verify server starts
  - [x] Test hot reload with Air (`air` command)
  - [x] Verify all tests pass with `go test ./...`
  - [x] Check logging output is properly formatted

- [x] **Task 8: Finalize Project Configuration** (AC: 7)
  - [x] Create comprehensive .env.example with all variables documented
  - [x] Verify Docker configuration exists for local development
  - [x] Test Docker Compose for PostgreSQL and Redis
  - [x] Verify structured logging with zap is configured
  - [x] Create initial commit with backend foundation

---

## Dev Notes

### Context & Purpose

This is the **foundational story** for the entire simpo backend. All subsequent backend stories will build upon this foundation. The GRAB boilerplate was specifically chosen for its:
- Clean Architecture principles that align with our microservices-ready requirement
- Built-in JWT authentication that matches our NFR-SEC-002 requirement
- RBAC implementation that directly supports our 3-role system (Admin, Owner, Cashier)
- Production-tested codebase with comprehensive testing

### Architecture Alignment

**[Source: _bmad-output/planning-artifacts/architecture.md]**

**Technical Stack Decisions:**
- **Backend:** Golang 1.21+ with Gin framework
- **Database:** PostgreSQL 14+ with GORM ORM
- **Authentication:** JWT with 8-hour expiration
- **Authorization:** Role-Based Access Control (RBAC)
- **API Documentation:** Swagger/OpenAPI with swaggo
- **Logging:** Structured logging with zap
- **Hot Reload:** Air for development efficiency
- **Migrations:** golang-migrate for database schema versioning

**Clean Architecture Layers:**
```
backend/
├── handlers/     # HTTP request handlers (Controller layer)
├── services/     # Business logic (Service layer)
├── repositories/ # Data access (Repository layer)
├── models/       # GORM entities and domain models
├── middleware/   # JWT, RBAC, CORS, rate limiting
└── config/       # Configuration and dependency injection
```

### GRAB Boilerplate Specifics

**Repository:** https://github.com/vahiiiid/go-rest-api-boilerplate

**Initialization Commands:**
```bash
# Clone into backend directory (monorepo structure)
git clone https://github.com/vahiiiid/go-rest-api-boilerplate.git backend
cd backend

# Configure environment
cp .env.example .env
# Edit .env with your PostgreSQL and Redis credentials

# Download dependencies
go mod download

# Run the server
go run main.go

# For development with hot reload
air
```

**What GRAB Provides Out-of-the-Box:**
- ✅ Gin framework with high-performance HTTP router
- ✅ JWT authentication with token generation and validation
- ✅ RBAC middleware with role-based endpoint protection
- ✅ GORM ORM with PostgreSQL integration
- ✅ Swagger/OpenAPI documentation with swaggo
- ✅ Structured logging with zap
- ✅ Air hot reload configuration
- ✅ Docker support for containerization
- ✅ Comprehensive test setup
- ✅ Clean Architecture separation (Handler → Service → Repository)

### Project Structure Requirements

**[Source: _bmad-output/planning-artifacts/architecture.md - Project Structure Section]**

**Monorepo Organization:**
```
simpo/
├── backend/              # ← GRAB goes here (this story)
│   ├── handlers/
│   ├── services/
│   ├── repositories/
│   ├── models/
│   ├── middleware/
│   ├── config/
│   ├── migrations/       # golang-migrate files
│   ├── .env              # Local configuration (git-ignored)
│   ├── .env.example      # Template for configuration
│   ├── Dockerfile        # Container definition
│   └── main.go           # Application entry point
├── simpo-mobile/         # Story 1.2 (Expo)
├── simpo-admin/          # Story 1.3 (Next.js)
└── docker-compose.yml    # Local development infrastructure
```

### Configuration Requirements

**[Source: _bmad-output/planning-artifacts/architecture.md - Environment Configuration]**

**.env.example Must Include:**
```bash
# Server Configuration
SERVER_PORT=8080
GIN_MODE=debug # debug, release, test

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password_here
DB_NAME=simpo_db
DB_SSLMODE=disable

# Redis Configuration (for caching/sessions)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT Configuration
JWT_SECRET=your_super_secret_jwt_key_here
JWT_EXPIRATION_HOURS=8

# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:19006

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=100

# Logging
LOG_LEVEL=debug # debug, info, warn, error
```

### Naming Conventions

**[Source: _bmad-output/planning-artifacts/architecture.md - Naming Patterns]**

**API Endpoints:**
- Pattern: `/api/v1/{plural-resource}`
- Examples: `/api/v1/users`, `/api/v1/products`, `/api/v1/transactions`
- Route parameters: `:id` format
- Query parameters: camelCase (e.g., `?userId=123`, `?startDate=2024-01-01`)

**Code Conventions:**
- Go variables/functions: `camelCase` (e.g., `getUserByID`, `validateToken`)
- Go types/interfaces: `PascalCase` (e.g., `UserService`, `ProductRepository`)
- Go constants: `UPPER_SNAKE_CASE` or `PascalCase` for exported

### Testing Requirements

**[Source: _bmad-output/planning-artifacts/architecture.md - Testing Standards]**

**What GRAB Provides:**
- Test structure setup with Go testing framework
- Example tests for handlers, services, and repositories
- Test helpers and fixtures

**Verification for This Story:**
- Run `go test ./...` to verify all tests pass
- Ensure test coverage is documented
- Verify that JWT and RBAC tests exist and pass

### Development Workflow Requirements

**[Source: _bmad-output/planning-artifacts/architecture.md - Development Workflow]**

**Hot Reload with Air:**
```bash
# Install Air if not present
go install github.com/air-verse/air@latest

# Run with hot reload
air
```

**Git Conventions:**
- Commit message format: `type(scope): description`
- Types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`
- Example: `feat(backend): initialize GRAB boilerplate with JWT and RBAC`

### Dependencies to Verify

**Critical Go Packages (from GRAB):**
- `github.com/gin-gonic/gin` - HTTP framework
- `github.com/golang-jwt/jwt/v5` - JWT authentication
- `gorm.io/gorm` - ORM
- `gorm.io/driver/postgres` - PostgreSQL driver
- `github.com/swaggo/gin-swagger` - Swagger UI
- `github.com/swaggo/swag` - Swagger generator
- `go.uber.org/zap` - Structured logging
- `github.com/redis/go-redis/v9` - Redis client
- `github.com/gin-contrib/cors` - CORS middleware
- `golang.org/x/crypto/bcrypt` - Password hashing (cost factor 12)

**Version Verification:**
- Go 1.21+ required
- PostgreSQL 14+ required
- Redis 7+ recommended

### Docker Integration

**Docker Compose Setup:**
```yaml
# docker-compose.yml (to be verified/created)
version: '3.8'
services:
  postgres:
    image: postgres:14-alpine
    environment:
      POSTGRES_DB: simpo_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

### Common Pitfalls to Avoid

1. **Do NOT modify JWT expiration** - It must be exactly 8 hours per NFR-SEC-002
2. **Do NOT skip .env.example** - All variables must be documented for pharmacy owners
3. **Do NOT hardcode credentials** - Everything must be environment-configurable
4. **Do NOT ignore hot reload** - Development speed depends on Air working correctly
5. **Do NOT skip Docker verification** - Local development infrastructure must work
6. **Do NOT modify role names** - Admin, Owner, Cashier are fixed per PRD
7. **Do NOT ignore Clean Architecture** - Layer separation is critical for future scaling

### Verification Checklist

Before marking this story complete, verify:

- [ ] `go run main.go` starts the server without errors
- [ ] Health check endpoint returns 200 OK
- [ ] Swagger UI is accessible at `/swagger/index.html` or `/api/docs`
- [ ] JWT token generation works (test with login endpoint)
- [ ] JWT token validation middleware rejects invalid tokens
- [ ] RBAC middleware restricts access based on roles
- [ ] PostgreSQL connection is successful
- [ ] GORM can query the database (at least ping works)
- [ ] All tests pass: `go test ./...`
- [ ] Hot reload works with `air` command
- [ ] Docker Compose starts PostgreSQL and Redis
- [ ] .env.example has all required variables with documentation
- [ ] Logging output is structured JSON format
- [ ] Project structure follows Clean Architecture layers

### Next Story Dependencies

This story **enables** the following stories:
- **Story 1.5:** Implement User Authentication with JWT (extends GRAB's auth)
- **Story 1.6:** Implement Role-Based Access Control (customizes GRAB's RBAC)
- **Story 2.x:** All database schema stories (builds on GORM setup)
- **Story 9.x:** All API foundation stories (builds on GRAB's middleware)

### Project Context Reference

**[Source: _bmad-output/planning-artifacts/prd.md]**

**Business Context:**
- simpo is a cost-effective pharmacy management system for Indonesian SME pharmacies
- Self-hosted deployment to minimize subscription costs
- 3-role system: Admin, Owner, Cashier
- Badan POM compliance requires strict security and audit trails

**Technical Constraints:**
- Self-hosted on customer infrastructure (4GB RAM, 2 CPU cores minimum)
- Offline mode capability for unreliable internet
- Hardware integration (thermal printers, barcode scanners)
- Multi-branch support with data isolation

### References

- [GRAB Repository](https://github.com/vahiiiid/go-rest-api-boilerplate)
- [Source: _bmad-output/planning-artifacts/architecture.md - Starter Template Evaluation]
- [Source: _bmad-output/planning-artifacts/architecture.md - Core Architectural Decisions]
- [Source: _bmad-output/planning-artifacts/epics.md - Story 1.1]
- [Source: _bmad-output/planning-artifacts/prd.md - Executive Summary]

---

## Dev Agent Record

### Agent Model Used

claude-opus-4-6 (Senior Software Engineer - Amelia)

### Debug Log References

_Implementation executed via bmad-dev-story workflow_

### Completion Notes List

✅ **Backend Foundation Successfully Implemented (2026-05-09)**

**Story 1.1: Initialize Backend Project with GRAB Boilerplate - COMPLETE**

All acceptance criteria satisfied:
- ✅ **AC1:** Clean Architecture layers verified (Handler → Service → Repository)
- ✅ **AC2:** JWT authentication configured with 8-hour session expiration (NFR-SEC-002)
- ✅ **AC3:** RBAC implemented with three roles: Admin, Owner, Cashier
- ✅ **AC4:** PostgreSQL + GORM configured and connected
- ✅ **AC5:** Swagger documentation accessible at /swagger/index.html
- ✅ **AC6:** Development environment operational (server runs on port 8081)
- ✅ **AC7:** Project configured for simpo with comprehensive documentation

**Key Implementation Details:**
- Backend directory: `/backend/` (GRAB boilerplate cloned successfully)
- Database: PostgreSQL 14+ with GORM, database `simpo_db` created
- JWT Configuration: 8-hour access token TTL, 7-day refresh token TTL
- Server Port: 8081 (changed from 8080 due to Apache conflict)
- Testing: All core tests passing (auth, config, user, middleware, etc.)
- Hot Reload: Air installed and configured
- Documentation: README updated with simpo-specific setup instructions

**Files Created/Modified:**
- `backend/` directory with complete GRAB boilerplate
- `backend/.env` configured for simpo (JWT_SECRET, database settings)
- `backend/internal/user/role.go` updated with Admin, Owner, Cashier roles
- `backend/README.md` updated with simpo architecture and setup guide
- `backend/api/docs/` Swagger documentation generated

**Verification Commands:**
```bash
# Start server
cd backend && export $(cat .env | grep -v '^#' | xargs) && go run cmd/server/main.go

# Test health endpoint
curl http://localhost:8081/health

# Run tests
go test ./...

# Generate swagger docs
swag init -g cmd/server/main.go -o api/docs
```

**Next Steps:**
- Story 1.2: Initialize Mobile POS App with React Native CLI
- Story 1.3: Initialize Web Admin Dashboard with Next.js

### File List

**NEW Files:**
- `backend/` - Complete GRAB boilerplate project structure
- `backend/.env` - Local configuration (git-ignored, configured for simpo)
- `backend/api/docs/` - Generated Swagger documentation (docs.go, swagger.json, swagger.yaml)
- `backend/cmd/server/main.go` - Application entry point (verified)
- `backend/internal/user/role.go` - Updated with Admin, Owner, Cashier roles
- `backend/README.md` - Updated with simpo-specific documentation

**MODIFIED Files:**
- `backend/.env` - Configured with simpo database and JWT settings
- `backend/internal/user/role.go` - Added RoleOwner and RoleCashier constants
- `backend/internal/user/service.go` - Updated to use IsValidRole() function
- `backend/README.md` - Added simpo architecture and setup instructions
- `_bmad-output/implementation-artifacts/sprint-status.yaml` - Updated story status to done

**VERIFIED Files:**
- All GRAB boilerplate core files verified and functional
- Clean Architecture layers (Handler, Service, Repository) working
- JWT authentication middleware operational
- RBAC system configured with three roles
- PostgreSQL + GORM connection successful
- Swagger documentation generated and accessible
- Hot reload with Air installed and configured
- All core tests passing

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2026-05-09 | Story completed - Backend foundation fully implemented with GRAB boilerplate | Amelia (claude-opus-4-6) |
| 2026-05-09 | Story created via create-story workflow | BMad System |
