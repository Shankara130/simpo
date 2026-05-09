# Story 1.1: Initialize Backend Project with GRAB Boilerplate

Status: ready-for-dev

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

- [ ] **Task 1: Clone and Configure GRAB Boilerplate** (AC: 1, 7)
  - [ ] Clone GRAB repository from `https://github.com/vahiiiid/go-rest-api-boilerplate.git` into `backend/` directory
  - [ ] Copy `.env.example` to `.env` and configure database variables
  - [ ] Update project name and descriptions in README and configuration files
  - [ ] Verify Go version compatibility (Go 1.21+ required)
  - [ ] Download dependencies with `go mod download`

- [ ] **Task 2: Verify Clean Architecture Structure** (AC: 1)
  - [ ] Confirm Handler layer exists with proper HTTP handlers
  - [ ] Confirm Service layer exists with business logic
  - [ ] Confirm Repository layer exists with data access
  - [ ] Verify dependency injection is properly implemented
  - [ ] Document the layer structure in project README

- [ ] **Task 3: Configure JWT Authentication** (AC: 2)
  - [ ] Set JWT secret key in .env configuration
  - [ ] Verify 8-hour session expiration is configured
  - [ ] Test token generation with a sample login
  - [ ] Verify token validation middleware is working
  - [ ] Document JWT configuration for team reference

- [ ] **Task 4: Configure Role-Based Access Control (RBAC)** (AC: 3)
  - [ ] Verify three roles exist: Admin, Owner, Cashier
  - [ ] Check role permissions are defined in code
  - [ ] Test authorization middleware with different roles
  - [ ] Document RBAC rules and permission matrix

- [ ] **Task 5: Set Up PostgreSQL and GORM** (AC: 4)
  - [ ] Configure PostgreSQL connection in .env
  - [ ] Verify GORM is properly initialized
  - [ ] Test database connection with health check
  - [ ] Configure connection pooling settings
  - [ ] Set up golang-migrate for future migrations

- [ ] **Task 6: Verify Swagger Documentation** (AC: 5)
  - [ ] Access Swagger UI at configured endpoint
  - [ ] Verify swaggo is integrated in the project
  - [ ] Test API documentation generation from Go annotations
  - [ ] Document how to update Swagger docs for team

- [ ] **Task 7: Configure Development Environment** (AC: 6)
  - [ ] Run `go run main.go` and verify server starts
  - [ ] Test hot reload with Air (`air` command)
  - [ ] Verify all tests pass with `go test ./...`
  - [ ] Check logging output is properly formatted

- [ ] **Task 8: Finalize Project Configuration** (AC: 7)
  - [ ] Create comprehensive .env.example with all variables documented
  - [ ] Verify Docker configuration exists for local development
  - [ ] Test Docker Compose for PostgreSQL and Redis
  - [ ] Verify structured logging with zap is configured
  - [ ] Create initial commit with backend foundation

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

_Generated by create-story workflow_

### Debug Log References

_Story not yet implemented - no debug log_

### Completion Notes List

_Story not yet implemented - awaiting dev-story workflow_

### File List

_Files to be created/modified during implementation:_

**NEW Files:**
- `backend/` - GRAB boilerplate cloned here
- `backend/.env` - Local configuration (git-ignored)
- `backend/.env.example` - Configuration template
- `backend/Dockerfile` - Container definition
- `docker-compose.yml` - Local development infrastructure (if not exists)

**MODIFIED Files:**
- `.gitignore` - Ensure .env is ignored

**VERIFIED Files:**
- All GRAB boilerplate files should be present and unmodified unless specified

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2026-05-09 | Story created via create-story workflow | BMad System |
