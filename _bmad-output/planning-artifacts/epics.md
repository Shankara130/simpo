---
stepsCompleted: ["step-01-extract", "step-02-design-epics", "step-03-generate-stories", "step-04-final-validation"]
inputDocuments: ["prd.md", "architecture.md"]
workflowType: 'epics-and-stories'
project_name: 'simpo'
user_name: 'Shankara'
date: '2026-05-08'
lastStep: 4
status: 'complete'
completedAt: '2026-05-08T23:15:00+07:00'
---

# simpo - Epic Breakdown

This document provides the complete epic and story breakdown for simpo, decomposing the requirements from the PRD and Architecture requirements into implementable stories.

## Requirements Inventory

### Functional Requirements

FR1: System Administrator can create new user accounts with role assignment (Admin, Owner, Cashier)
FR2: System Administrator can deactivate existing user accounts
FR3: Users can authenticate into the system using username and password credentials
FR4: System can enforce role-based access control, restricting users to actions permitted for their assigned role
FR5: System can terminate user sessions after 8 hours of inactivity for security
FR6: Cashier can process sales transactions by scanning product barcodes or manual item entry
FR7: Cashier can apply payment methods including cash, bank transfer, and e-wallet for each transaction
FR8: Cashier can generate and print transaction receipts using thermal printer
FR9: System can calculate transaction totals including item prices, discounts, and taxes
FR10: System can complete end-to-end sales transactions within 30 seconds
FR11: Owner can view current stock levels for all products in real-time
FR12: Cashier can check product availability during sales transactions
FR13: System Administrator can manually adjust stock quantities with required reason logging
FR14: System can automatically deduct sold items from inventory during sales transactions
FR15: System can maintain stock reconciliation accuracy greater than 99%
FR16: System can automatically generate low stock notifications when product quantity falls below configurable reorder threshold
FR17: System can automatically generate expiry date alerts at 30-day, 14-day, and 7-day intervals before expiration
FR18: Owner can view aggregated alerts and notifications in a centralized dashboard
FR19: System can prevent sale of expired medications
FR20: Owner can generate daily sales summary reports showing total sales, payment methods, and top-selling products
FR21: Owner can generate basic profit/loss reports showing revenue, cost of goods sold, and gross profit
FR22: Owner can export financial reports in PDF and Excel formats
FR23: System can maintain complete audit trail of all financial transactions for compliance purposes
FR24: System Administrator can configure system settings including business name, address, and contact information
FR25: System Administrator can monitor system health and uptime through admin dashboard
FR26: System can maintain complete append-only audit trail logging all system changes with user, timestamp, and reason
FR27: System can automatically perform daily backups of all data with 30-day retention
FR28: System can interface with thermal printers supporting ESC/POS protocol for receipt printing
FR29: System can interface with USB and Bluetooth barcode scanners for product scanning
FR30: System can control cash drawers via printer kick command
FR31: System can generate barcode/sticker labels using label printer
FR32: Cashier can process sales transactions without internet connectivity
FR33: System can queue offline transactions for synchronization when connectivity is restored
FR34: System can automatically synchronize data when internet connectivity resumes
FR35: System can provide visual indicators showing synchronization status (synced, pending, failed)

### Non-Functional Requirements

**Performance:**
NFR-PERF-001: System shall complete end-to-end sales transactions within 30 seconds from product scan to receipt print
NFR-PERF-002: System shall process product barcode scans and return product information within 1 second
NFR-PERF-003: System shall generate daily sales summary reports within 10 seconds
NFR-PERF-004: System shall synchronize offline transactions to server within 5 seconds of connectivity restoration
NFR-PERF-005: System shall load admin dashboard within 3 seconds on standard broadband connection
NFR-PERF-006: System shall respond to user interactions (button clicks, form submissions) within 500 milliseconds
NFR-PERF-007: System shall support concurrent usage by 5 cashiers across multiple terminals with <2 second response degradation
NFR-PERF-008: System shall maintain stock reconciliation accuracy greater than 99% during ongoing operations

**Security:**
NFR-SEC-001: System shall enforce role-based access control with three roles: System Admin, Owner, Cashier
NFR-SEC-002: System shall terminate user sessions after 8 hours of inactivity
NFR-SEC-003: System shall require strong password storage using industry-standard hashing algorithms
NFR-SEC-004: System shall maintain append-only audit trail for all system changes with user identification, timestamp, and reason
NFR-SEC-005: System shall encrypt all data at rest using AES-256 encryption or equivalent
NFR-SEC-006: System shall encrypt all data in transit using TLS 1.3 or higher
NFR-SEC-007: System shall support daily automated backups with 30-day retention
NFR-SEC-008: System shall prevent data deletion for audit compliance (append-only logs)
NFR-SEC-009: System shall maintain complete audit trail for all inventory transactions (purchase, sale, adjustment, disposal) for minimum 5 years
NFR-SEC-010: System shall provide read-only audit mode for compliance verification without data alteration capabilities
NFR-SEC-011: System shall prevent sale of expired medications through blocking logic at transaction time

**Reliability:**
NFR-REL-001: System shall achieve 99.5% uptime excluding planned maintenance windows
NFR-REL-002: System shall provide automatic data backup every 24 hours with 30-day retention
NFR-REL-003: System shall maintain failed transaction rate less than 0.1% of total transactions
NFR-REL-004: System shall provide offline mode capability for continued POS operations during internet outages
NFR-REL-005: System shall automatically queue offline transactions when internet connectivity is lost
NFR-REL-006: System shall synchronize offline data when connectivity resumes with conflict resolution (last-write-wins with manual override)
NFR-REL-007: System shall provide visual indicators for synchronization status (synced, pending, failed)

**Integration:**
NFR-INT-001: System shall support thermal printers using ESC/POS protocol for receipt printing
NFR-INT-002: System shall support USB HID barcode scanners for product scanning
NFR-INT-003: System shall support Bluetooth barcode scanners for wireless product scanning
NFR-INT-004: System shall control cash drawers via printer kick command (RJ-12 interface)
NFR-INT-005: System architecture shall support future API integrations with accounting software (Jurnal, Accurate, Zoho Books)
NFR-INT-006: System architecture shall support future payment gateway integrations (QRIS, e-wallets, debit cards)

**Scalability:**
NFR-SCAL-001: System shall support 5 concurrent cashier users across multiple POS terminals with <2 second response degradation
NFR-SCAL-002: System shall support up to 10,000 product SKUs in inventory database
NFR-SCAL-003: System shall support up to 50,000 transaction records per month
NFR-SCAL-004: System architecture shall support horizontal scaling to 5 pharmacy branches in Phase 2
NFR-SCAL-005: Database architecture shall support growth to 100 pharmacy customers by Year 3

### Additional Requirements (from Architecture)

**Starter Templates:**
- Backend: GRAB (Go REST API Boilerplate) - Clean Architecture with JWT auth, RBAC, PostgreSQL migrations
- Mobile: Expo (TypeScript) - blank-typescript template with EAS Build support
- Web: Next.js 15 + TypeScript + Tailwind - create-next-app with TypeScript and Tailwind flags

**Technical Stack Decisions:**
- Backend: Golang 1.21+ with Gin framework
- Mobile: React Native via Expo SDK 50+
- Web: Next.js 15 with App Router
- Database: PostgreSQL 14+ with GORM ORM
- Caching: Redis 7+ for sessions, queries, and pub/sub
- Deployment: Self-hosted Docker Compose

**Infrastructure Requirements:**
- Self-hosted deployment on customer infrastructure (4GB RAM, 2 CPU cores, 50GB storage minimum)
- Docker Compose for PostgreSQL and Redis
- Manual update mechanism for MVP (scripts), automated for future versions
- Environment configuration via .env files with .env.example template

**Data Architecture Requirements:**
- Code-First data modeling with GORM
- Hybrid validation (database-level + application-level)
- golang-migrate for database migrations with explicit migration files
- Layered caching strategy (session, query, pub/sub)

**Security Implementation Requirements:**
- bcrypt password hashing with cost factor 12
- JWT authentication with 8-hour session timeout
- RFC 7807 (Problem Details) for error responses
- Rate limiting: 100 req/min per user using Gin middleware
- TLS 1.3 enforcement for all API communication

**API Design Requirements:**
- REST API with plural resource naming (/api/v1/users, /api/v1/products)
- Swagger/OpenAPI documentation with swaggo
- API versioning with /api/v1/ prefix
- Direct success responses (no wrapper)
- RFC 7807 error responses

**Hardware Integration Requirements:**
- ESC/POS protocol support for 58mm and 80mm thermal printers
- USB HID and Bluetooth barcode scanner compatibility
- RJ-12 cash drawer interface via printer kick command
- Label printer support for barcode/sticker generation

**Offline Architecture Requirements:**
- Local SQLite storage on mobile devices for offline transaction processing
- Bidirectional synchronization queue with conflict resolution
- Background sync with exponential backoff retry logic
- Visual sync status indicators (synced, pending, failed)

**Project Structure Requirements:**
- Feature-based organization for all codebases
- Co-located test files (file_test.go, file.test.tsx)
- Monorepo structure with backend/, simpo-mobile/, simpo-admin/ directories
- Clean Architecture layers (Handler → Service → Repository)

**Development Workflow Requirements:**
- GitHub Actions CI/CD pipeline for all three components
- Hot reload for development (Air for Go, Expo dev server, Next.js dev server)
- Docker Compose for local development infrastructure
- Structured logging (zap for Go, pino for Node.js)

### UX Design Requirements

No UX Design specification document was provided. UX requirements will be derived from the PRD user journeys and Architecture patterns.

### FR Coverage Map

All 42 functional requirements are categorized and will be mapped to epics and stories:

**Authentication & User Management (FR1-5):** Will be covered in Authentication Epic
**Point of Sale (FR6-10):** Will be covered in Point of Sale Epic
**Inventory Management (FR11-15):** Will be covered in Inventory Management Epic
**Alerts & Notifications (FR16-19):** Will be covered in Alerts & Notifications Epic
**Financial Reporting (FR20-23):** Will be covered in Reporting Epic
**System Administration (FR24-27):** Will be covered in System Administration Epic
**Hardware Integration (FR28-31):** Will be covered in Hardware Integration Epic (Mobile)
**Offline Mode & Synchronization (FR32-35):** Will be covered in Offline Sync Epic (Mobile)
**Supplier & Purchase Management (FR36-42):** Will be covered in Supplier & Purchase Management Epic

## Epic List

Epic 1: Authentication & User Management
Epic 2: Database Schema & Migrations
Epic 3: Point of Sale (Mobile)
Epic 4: Inventory Management
Epic 5: Financial Reporting
Epic 6: System Administration & Configuration
Epic 7: Hardware Integration (Mobile)
Epic 8: Offline Mode & Synchronization (Mobile)
Epic 9: API Foundation & Core Services
Epic 10: Supplier & Purchase Management

## Epic 1: Authentication & User Management

Enable secure access to the system with role-based permissions for three user types (System Admin, Owner, Cashier), ensuring appropriate access control and audit trails.

### Story 1.1: Initialize Backend Project with GRAB Boilerplate

**As a** Development Team,
**I want to** set up the backend project using the GRAB (Go REST API Boilerplate) starter template,
**So that** we have a solid foundation with Clean Architecture, JWT authentication, and RBAC already implemented.

**Acceptance Criteria:**

**Given** the GRAB boilerplate repository exists
**When** the development team clones the repository
**Then** the project structure includes Clean Architecture layers (Handler, Service, Repository)
**And** JWT authentication middleware is configured
**And** Role-Based Access Control (RBAC) is set up
**And** PostgreSQL connection is configured with GORM
**And** Swagger documentation is ready for customization

### Story 1.2: Initialize Mobile POS App with Expo

**As a** Development Team,
**I want to** create the mobile POS application using Expo with TypeScript template,
**So that** cashiers have a modern React Native foundation for the point-of-sale interface.

**Acceptance Criteria:**

**Given** the Expo CLI is available
**When** the development team creates a new Expo project
**Then** the project is initialized with TypeScript
**And** the project structure includes features/ directory for feature-based organization
**And** app.json is configured for Android-first deployment
**And** EAS Build configuration is set up for future Android builds

### Story 1.3: Initialize Web Admin Dashboard with Next.js

**As a** Development Team,
**I want to** create the web admin dashboard using Next.js with TypeScript and Tailwind CSS,
**So that** pharmacy owners and system admins have a modern web interface for business oversight.

**Acceptance Criteria:**

**Given** the create-next-app CLI is available
**When** the development team creates a new Next.js project
**Then** the project is initialized with TypeScript
**And** Tailwind CSS is configured for styling
**And** App Router structure is enabled
**And** ESLint is configured for code quality
**And** pages/ directory structure is established for route organization

### Story 1.4: Set Up Development Infrastructure

**As a** Development Team,
**I want to** configure local development infrastructure with Docker Compose,
**So that** all team members can run PostgreSQL and Redis services locally for development.

**Acceptance Criteria:**

**Given** Docker and Docker Compose are installed on development machines
**When** the development team runs docker-compose up
**Then** PostgreSQL 14+ container is running on port 5432
**And** Redis 7+ container is running on port 6379
**And** Environment variables are configured for database and Redis connections
**And** Data persists across container restarts via Docker volumes

### Story 1.5: Implement User Authentication with JWT

**As a** System,
**I want** users to authenticate securely into the system using JWT tokens,
**So that** only authorized users can access pharmacy data and perform actions based on their role.

**Acceptance Criteria:**

**Given** a user has valid credentials (username and password)
**When** the user logs in via API or mobile/web client
**Then** the system validates credentials using bcrypt password comparison
**And** generates a JWT token with 8-hour expiration
**And** returns the token to the client for subsequent requests
**And** the token includes user role information for authorization

### Story 1.6: Implement Role-Based Access Control (RBAC)

**As a** System,
**I want** to enforce role-based permissions so users can only access appropriate features and data,
**So that** cashiers cannot access owner-level reports, and data isolation is maintained across branches.

**Acceptance Criteria:**

**Given** a JWT token includes user role information (Admin, Owner, Cashier)
**When** a user makes an API request
**Then** the system validates the token and extracts user role
**And** checks if the requested endpoint is permitted for that role
**And** grants access if permitted, or returns 403 Forbidden if not
**And** branch-level access control is enforced (cashiers can only access their assigned branch)

### Story 1.7: Implement User Registration with Admin Approval

**As a** System Administrator,
**I want** to register new staff accounts with role assignment,
**So that** new cashiers and admins can be onboarded with appropriate permissions.

**Acceptance Criteria:**

**Given** a System Administrator is logged in with Admin role
**When** creating a new user account
**Then** the admin provides username, password, email, and role selection
**And** the system validates the input data
**And** hashes the password using bcrypt with cost factor 12
**And** stores the user in the database with inactive status (if approval workflow) or active status
**And** assigns the selected role with appropriate permissions
**And** logs the user creation action in the audit trail with admin user ID

### Story 1.8: Implement Session Management with Timeout

**As a** System,
**I want** to automatically terminate user sessions after 8 hours of inactivity for security,
**So that** unauthorized access is prevented if a device is left unattended.

**Acceptance Criteria:**

**Given** a user is actively making requests within the 8-hour window
**When** the user continues activity
**Then** the session remains active and valid
**When** the user has no activity for 8 hours
**Then** the system automatically invalidates the JWT token
**And** requires the user to re-authenticate to continue
**And** returns 401 Unauthorized for subsequent requests until re-authentication

### Story 1.9: Implement Staff Registration via Whitelist

**As a** System Administrator,
**I want** to allow staff registration via email domain whitelist as an alternative to manual account creation,
**So that** new staff can self-register with their work email if the domain is approved.

**Acceptance Criteria:**

**Given** the System Administrator has configured approved email domains in the whitelist
**When** a new staff member attempts to register with an approved domain email
**Then** the system validates the email domain against the whitelist
**And** allows registration with email verification
**And** assigns a default role (typically Cashier) with appropriate permissions
**And** requires the user to set their password on first login
**And** logs the self-registration action in the audit trail

### Story 1.10: Implement User Deactivation

**As a** System Administrator,
**I want** to deactivate user accounts when staff leave the organization,
**So that** former employees cannot access the system and data remains secure.

**Acceptance Criteria:**

**Given** a System Administrator is logged in with Admin role
**When** deactivating a user account
**Then** the system marks the user account as inactive
**And** prevents the user from authenticating (invalid credentials on login attempt)
**And** revokes all active JWT tokens for that user
**And** logs the deactivation action in the audit trail with admin user ID and reason

## Epic 2: Database Schema & Migrations

Define and implement the complete database schema with support for pharmacies, products, transactions, and multi-branch operations, using PostgreSQL with GORM ORM.

### Story 2.1: Design Database Schema for MVP

**As a** Development Team,
**I want** to design a complete database schema that supports all Phase 1 MVP features,
**So that** we have a clear blueprint for implementing data models with GORM.

**Acceptance Criteria:**

**Given** the PRD defines all entities and relationships for Phase 1 MVP
**When** designing the database schema
**Then** all entities are identified with attributes and relationships:
  - Users (id, username, password_hash, email, role, branch_id, status, created_at, updated_at)
  - Products (id, sku, name, description, stock_qty, price, expiry_date, branch_id, created_at, updated_at)
  - Transactions (id, transaction_number, cashier_id, total, payment_method, status, created_at, updated_at)
  - TransactionItems (id, transaction_id, product_id, quantity, unit_price, subtotal, created_at)
  - Branches (id, name, address, phone, created_at, updated_at)
**And** all relationships are defined (Users → Branches, Transactions → Users, TransactionItems → Transactions)
**And** indexes are identified for performance (user uniqueness, product SKU, transaction dates)
**And** constraints are defined (NOT NULL on required fields, foreign key relationships, check constraints)
**And** schema supports branch-level data isolation

### Story 2.2: Create Initial Migration with golang-migrate

**As a** Development Team,
**I want** to create the initial database migration using golang-migrate,
**So that** we can version control database schema changes and support rollback capabilities.

**Acceptance Criteria:**

**Given** the database schema design is complete
**When** creating migration files
**Then** SQL UP migration files are created for each table
**And** SQL DOWN migration files are created for rollback capability
**And** migrations follow naming convention: 000001_create_{table}_table.{up|down}.sql
**And** migrations include all columns, indexes, and constraints
**And** migrations use PostgreSQL syntax appropriate for version 14+
**And** migration files are stored in backend/migrations/ directory

### Story 2.3: Implement GORM Models with Struct Tags

**As a** Development Team,
**I want** to implement GORM model structs with proper tags for database mapping and JSON serialization,
**So that** the ORM layer can interact with PostgreSQL and serialize data for API responses.

**Acceptance Criteria:**

**Given** the database schema is defined
**When** implementing GORM models
**Then** each entity has a corresponding Go struct with GORM tags:
  - Users model with gorm tags (table: users, primaryKey, uniqueIndex, not null, etc.)
  - Products model with price as decimal string for precision
  - Transactions model with transaction_number field
  - TransactionItems model with foreign key to Transactions and Products
  - Branches model for multi-location support
**And** JSON tags use camelCase for API responses (id, sku, stockQty, price, expiryDate, createdAt, updatedAt)
**And** models include CreatedAt and UpdatedAt timestamp fields
**And** foreign key relationships are defined with BelongsTo and HasMany GORM methods

### Story 2.4: Implement Database Connection and Pooling

**As a** Development Team,
**I want** to configure PostgreSQL connection with connection pooling for optimal performance,
**So that** the backend can handle concurrent requests from multiple cashiers efficiently.

**Acceptance Criteria:**

**Given** PostgreSQL is running via Docker Compose
**When** the backend application starts
**Then** GORM connects to PostgreSQL using configured credentials from .env
**And** connection pooling is configured with appropriate settings (max open connections, max idle connections, connection lifetime)
**And** database connection is established successfully
**And** the application logs successful connection on startup
**And** connection errors are handled gracefully with appropriate error messages

### Story 2.5: Implement Repository Layer for Data Access

**As a** Development Team,
**I want** to implement the repository layer that abstracts database operations,
**So that** business logic in services is decoupled from data access concerns.

**Acceptance Criteria:**

**Given** GORM models are defined
**When** implementing repositories
**Then** each entity has a corresponding repository (UserRepository, ProductRepository, TransactionRepository, BranchRepository)
**And** repositories provide CRUD methods (Create, Read, Update, Delete, List)
**And** repositories use GORM to interact with PostgreSQL
**And** repositories return domain entities or errors (not database-specific errors)
**And** complex queries (like filtering, pagination) are implemented in repositories
**And** repositories are injected into services via dependency injection

## Epic 3: Point of Sale (Mobile)

Enable cashiers to process sales transactions efficiently with barcode scanning, payment handling, and receipt printing, with sub-30-second transaction processing time.

### Story 3.1: Design POS Screen Layout and Navigation

**As a** Cashier,
**I want** a clean, intuitive POS screen layout optimized for fast transaction processing,
**So that** I can process customers quickly during peak hours without struggling with the interface.

**Acceptance Criteria:**

**Given** the mobile app is launched and cashier is authenticated
**When** the POS screen is displayed
**Then** the layout shows key controls at the top (product search/barcode scan, cart summary, payment)
**And** the center area shows product list with barcode scan input prominently
**And** the bottom area shows action buttons (add to cart, checkout, clear cart)
**And** cart items are displayed in a summary panel
And** the layout is optimized for one-handed operation in portrait mode
**And** the interface follows mobile app conventions (large touch targets, clear visual hierarchy)

### Story 3.2: Implement Barcode Scanner Integration

**As a** Cashier,
**I want** to scan product barcodes using USB or Bluetooth scanners to quickly add items to transactions,
**So that** I can process customers faster without manual product lookup.

**Acceptance Criteria:**

**Given** a barcode scanner is connected to the mobile device (USB HID or Bluetooth)
**When** the cashier scans a product barcode
**Then** the mobile app receives the barcode input
**And** queries the backend API to find product by SKU/barcode
**And** adds the product to the transaction cart if found
**Or** displays a "product not found" error if SKU is not in the system
**And** shows product details (name, price, current stock) in the cart
**And** plays a confirmation sound or vibration on successful scan

### Story 3.3: Implement Cart Management

**As a** a Cashier,
**I want** to manage the transaction cart with add, remove, and quantity adjustment capabilities,
**So that** I can build up transactions accurately and make corrections before finalizing.

**Acceptance Criteria:**

**Given** products have been scanned or added to the transaction
**When** the cashier views the cart
**Then** all cart items are displayed with product name, SKU, quantity, unit price, and subtotal
**And** the cashier can increase or decrease item quantities with +/- buttons
**And** the cashier can remove items from the cart
**And** the cart displays the running total amount
**And** the cart persists within the session (if transaction is not completed)

### Story 3.4: Implement Payment Method Selection

**As** a Cashier,
**I want** to select payment methods (cash, bank transfer, e-wallet) for each transaction,
**So that** I can accommodate customer preferences and complete transactions flexibly.

**Acceptance Criteria:**

**Given** a transaction cart has been built and customer is ready to pay
**When** the cashier initiates payment
**Then** payment method options are displayed (Cash, Bank Transfer, E-Wallet)
**And** the cashier selects the appropriate payment method
**And** for Cash: no additional input required
**And** for Bank Transfer: customer bank account name and reference number input is collected
**And** for E-Wallet: e-wallet app selection (GoPay, OVO, Dana, ShopeePay) and payment confirmation input is collected
**And** the selected payment method is stored with the transaction record

### Story 3.5: Implement Receipt Printing with Thermal Printer

**As** a Cashier,
**I want** to print transaction receipts automatically using a thermal printer after payment is complete,
**So that** customers receive proof of purchase and the pharmacy has transaction records.

**Acceptance Criteria:**

**Given** a transaction has been completed successfully
**When** the payment confirmation is received
**Then** the system generates a receipt in ESC/POS format
**And** sends the receipt to the thermal printer
**And** the receipt includes:
  - Pharmacy name and address
  - Transaction number and date/time
  - List of items with quantities and prices
  - Subtotal, tax (if applicable), and total
  - Payment method and change (if cash)
  - "Thank you" message
**And** the printer cuts the receipt after printing
**And** a success confirmation is displayed to the cashier

### Story 3.6: Implement Transaction Processing <30 Seconds

**As a** a Cashier,
**I want** to complete end-to-end sales transactions within 30 seconds from product scan to receipt print,
**So that** I can serve customers efficiently even during peak hours.

**Acceptance Criteria:**

**Given** a customer is ready to purchase products
**When** the cashier scans the first product
**Then** a timer starts tracking transaction duration
**And** product scanning and cart building completes in under 15 seconds
**And** payment processing completes in under 10 seconds
**And** receipt printing completes in under 5 seconds
**And** the total transaction time from first scan to receipt print is under 30 seconds
**And** the system logs the transaction completion time for performance monitoring

### Story 3.7: Implement Transaction History View

**As** a Cashier,
**I want** to view recent transaction history for reference or customer queries,
**So that** I can answer questions about recent purchases or reprint receipts if needed.

**Acceptance Criteria:**

**Given** the cashier is authenticated and has appropriate permissions
**When** accessing transaction history
**Then** the app displays a list of recent transactions for the current shift/day
**And** each transaction shows: transaction number, customer name (optional), total, status, timestamp
**And** tapping on a transaction shows full details including all items
**And** the cashier has an option to reprint receipts for recent transactions
**And** transaction history is filterable by date range and status

## Epic 4: Inventory Management

Enable real-time stock visibility across branches with stock adjustments, low stock notifications, and expiry date tracking to prevent stockouts and waste.

### Story 4.1: Implement Product List View with Search and Filters

**As a** Pharmacy Owner or Cashier,
**I want** to view a list of all products with search and filter capabilities,
**So that** I can quickly find specific products and check stock levels.

**Acceptance Criteria:**

**Given** the user is authenticated with appropriate permissions (Owner, Cashier)
**When** accessing the product list
**Then** products are displayed in a searchable list or grid view
**And** products can be searched by name or SKU
**And** products can be filtered by category or branch (for Owners)
And** each product displays: SKU, name, current stock quantity, price, expiry date
**And** low stock items are visually highlighted (red or orange indicator)
**And** expired items are visually marked and cannot be added to transactions
**And** the list supports pagination for large product catalogs (10K+ SKUs)

### Story 4.2: Implement Real-Time Stock Visibility

**As** a Pharmacy Owner,
I want** to see real-time stock levels across all branches,
So that** I can make informed decisions about stock transfers and reorders.

**Acceptance Criteria:**

**Given** the pharmacy owner is logged into the web dashboard or mobile app
**When** viewing product information
**Then** current stock quantity is displayed in real-time
**And** stock levels are updated immediately after sales transactions
**And** stock levels are updated immediately after stock adjustments
**And** owners can view stock levels by branch
**And** stock levels refresh automatically without manual refresh (real-time updates)
**And** the system maintains >99% stock reconciliation accuracy

### Story 4.3: Implement Manual Stock Adjustment

**As a** System Administrator,
**I want** to manually adjust stock quantities with reason logging for corrections,
**So that** inventory discrepancies can be resolved and audit trail compliance is maintained.

**Acceptance Criteria:**

**Given** the system administrator is authenticated and has appropriate permissions
**When** initiating a stock adjustment
**Then** the admin selects a product and branch location
**And** inputs the new stock quantity
**And** selects or enters a reason for the adjustment (damage, expiration, delivery receipt, etc.)
**And** the system updates the stock quantity in the database
**And** logs the adjustment in the append-only audit trail with admin user ID, timestamp, product, old quantity, new quantity, and reason
**And** triggers a stock level check for low stock notifications if applicable
**And** returns success confirmation to the administrator

### Story 4.4: Implement Low Stock Notifications

**As** a** Pharmacy Owner,
I want** to receive automatic notifications when products fall below configurable reorder thresholds,
**So that** I can reorder products before stockouts occur and avoid lost sales.

**Acceptance Criteria:**

**Given** reorder thresholds have been configured for products
**When** a product's stock quantity falls below its threshold after a sale
**Then** the system automatically detects the low stock condition
**And** generates a low stock notification event
**And** publishes the notification to Redis pub/sub with event type "stock.low"
**And** sends notifications to subscribed owners via:
  - Mobile app push notification
  - Web dashboard alert banner
  - Optional: Email notification (future enhancement)
**And** the notification includes: product SKU, product name, current stock, reorder threshold, and branch location
**And** the notification is actionable: "Order {quantity} units of {product} for {branch}"

### Story 4.5: Implement Expiry Date Alerts

**As** a** Pharmacy Owner,
I want** to receive advance alerts when products are approaching their expiry dates at 30, 14, and 7 days,
So that** I can discount or dispose of expiring medications proactively and comply with regulations.

**Acceptance Criteria:**

**Given** products have expiry dates recorded in the system
**When** the current date reaches 30 days before a product's expiry date
**Then** the system generates the first 30-day expiry alert
**And** when the current date reaches 14 days before expiry, the system generates a 14-day alert
**And** when the current date reaches 7 days before expiry, the system generates a 7-day alert
**And** each alert is published to Redis pub/sub with event type "product.expiry"
**And** notifications are displayed to owners via:
  - Mobile app alert banner
  - Web dashboard notifications
  - Optional: Email digest (future enhancement)
**And** the alert includes: product SKU, product name, expiry date, days remaining, branch location
And**But** the 7-day alert is marked as urgent with visual highlighting

### Story 4.6: Prevent Sale of Expired Medications

**As** a** System,
I want** to automatically block sales of expired medications to prevent regulatory compliance issues,
**So** the pharmacy avoids legal liability and protects public safety.

**Acceptance Criteria:**

**Given** a product has an expiry date recorded
**When** the current date is on or after the product's expiry date
**Then** the system marks the product as expired in the database
**And** the product is visually marked in product lists (grayed out with "EXPIRED" badge)
**And** attempting to add the product to a transaction cart is blocked
**And** an error message is displayed: "This product has expired and cannot be sold"
**And** the barcode scan for an expired product shows an error instead of adding to cart
**And** the audit trail logs any blocked sale attempt with user ID, timestamp, and product information

## Epic 5: Financial Reporting

Generate daily, weekly, and monthly financial reports with profit/loss analysis, sales summaries, and export capabilities for accounting and business oversight.

### Story 5.1: Implement Daily Sales Summary Report

**As a** Pharmacy Owner,
I want** to generate daily sales summary reports showing total sales, payment methods, and top-selling products,
So**that** I can track daily business performance and make operational decisions.

**Acceptance Criteria:**

**Given** the pharmacy owner is logged into the web dashboard
**When** generating a daily sales summary report
**Then** the report displays for the selected date and branches:
  - Total sales amount
  - Total number of transactions
  - Breakdown by payment method (Cash, Transfer, E-Wallet)
  - Top 10 selling products by quantity and revenue
  - Sales by hour (for operational insights)
**And** the report can be filtered by branch location
**And** the report is generated in under 10 seconds (NFR-PERF-003)
**And** the owner can export the report as PDF or Excel

### Story 5.2: Implement Profit/Loss Report

**As a** a** Pharmacy Owner,
I want** to generate basic profit/loss reports showing revenue, cost of goods sold, and gross profit,
So**that** I can understand business profitability and make informed decisions.

**Acceptance Criteria:**

**Given** the pharmacy owner is logged into the web dashboard
When** generating a profit/loss report
**Then** the report displays for the selected period (daily, weekly, monthly, custom range):
  - Total revenue (sales)
  - Cost of goods sold (COGS)
  - Gross profit (Revenue - COGS)
  - Gross profit margin percentage
**And** the report can be broken down by:
  - Product category
  - Branch location
  - Payment method
**And** the report data is calculated from transaction and product cost data
**And** the owner can export the report as PDF or Excel for accountant use

### Story 5.3: Implement Report Export Functionality

**As** a** Pharmacy Owner,
I**want** to export financial reports in PDF and Excel formats for sharing with accountants,
So**that** I can meet accounting and tax compliance requirements efficiently.

**Acceptance Criteria:**

**Given** a financial report has been generated and viewed
**When** the owner selects an export option (PDF or Excel)
**Then** the system generates the file in the selected format
**And** the PDF includes all report sections with proper formatting and company branding
**And** the Excel export includes raw data in spreadsheet format for further analysis
**And** the exported file is downloaded to the user's device
**And** the export includes metadata: report title, date range, generated timestamp, and branch location

### Story 5.4: Implement Append-Only Audit Trail for Compliance

As a System, I must maintain complete append-only audit trail of all financial transactions for Badan POM compliance and business accountability.

**Acceptance Criteria:**

**Given** any financial transaction (sale, return, adjustment) occurs
**When** the transaction is recorded in the database
**Then** the system automatically creates an immutable audit trail entry
**And** the audit entry includes:
  - Who: User ID and role who performed the action
  - When: Timestamp of the action
  - What: Description of the action (transaction created, stock adjusted, etc.)
  - Why: Reason for the action (if applicable)
**And** the audit entry is append-only (no modifications or deletions allowed)
**And** audit entries are stored in a separate audit_logs table with write-only access
**And** the audit trail is queryable for at least 5 years per Badan POM requirements
**And** audit logs can be exported for compliance inspections

## Epic 6: System Administration & Configuration

Enable system administrators to configure system settings, monitor system health, and manage the overall deployment, while giving owners business oversight capabilities.

### Story 6.1: Implement System Settings Configuration

**As a** System Administrator,
I**want** to configure system settings including business name, address, and contact information,
So**that** the system reflects pharmacy branding and contact information is accurate.

**Acceptance Criteria:**

**Given** the system administrator is logged in with Admin role
**When** accessing system configuration
**Then** the admin can view and edit:
  - Pharmacy business name
  - Pharmacy address
  - Phone number
  - Email address
  - Logo upload (future enhancement)
**And** changes are saved to the system configuration in the database
**And** updated information is reflected in receipts, reports, and UI
**And** all configuration changes are logged in the audit trail with admin user ID and timestamp

### Story 6.2: Implement System Health Monitoring Dashboard

**As a** System Administrator,
I**want** to monitor system health and uptime through an admin dashboard,
So**that** I can identify and resolve issues proactively before they impact operations.

**Acceptance Criteria:**

**Given** the system administrator is logged in with Admin role
**When** accessing the health monitoring dashboard
**Then** the dashboard displays:
  - System uptime percentage (NFR-REL-001: >99.5%)
  - Database connection status
  - Redis cache status
  - Active user sessions count
  - Recent error log entries
  - API response times
  - Disk storage usage
**And** health metrics refresh automatically every 30 seconds
**And** alerts are displayed for:
  - Database connection failures
  - Redis connection failures
  - Error rate exceeding 0.1% threshold
  - Disk space below 20% free
**And** the /health endpoint returns system status for external monitoring tools

### Story 6.3: Implement Automated Daily Backups

**As a** System,
I**want** to automatically perform daily backups of all data with 30-day retention,
So**that** pharmacy data is protected against data loss and Badan POM compliance requirements are met.

**Acceptance Criteria:**

**Given** the system is running and database changes occur
**When** the daily backup schedule triggers (configured time, e.g., 2:00 AM)
**Then** the system creates a full backup of the PostgreSQL database
**And** the backup is stored with timestamp in the configured backup location
**And** backups are retained for 30 days (older backups are automatically cleaned up)
**And** backup success or failure is logged in the system health log
**And** the system includes /api/v1/admin/backups endpoint for manual backup triggers
**And** the system supports restoration from any backup in the retention window

### Story 6.4: Implement Append-Only Audit Trail for System Changes

As a System, I must maintain complete append-only audit trail logging all system changes with user identification, timestamp, and reason for Badan POM compliance.

**Acceptance Criteria:**

**Given** any system change action is performed (user creation, configuration change, etc.)
**When** the action is executed
**Then** the system creates an immutable audit log entry
**And** the audit log includes:
  - Who: User ID and role
  - When: Timestamp with timezone
  - What: Description of the change (user created, setting updated, etc.)
  - Why: Reason for the change (from user input or system event)
**And** the audit log is stored in an append-only table (no modifications or deletions)
**And** audit logs are queryable via the admin dashboard with filters (date range, user, action type)
**And** audit logs can be exported in CSV/PDF for compliance inspections
**And** the audit trail retention period is minimum 5 years per Badan POM requirements

## Epic 7: Hardware Integration (Mobile)

Integrate mobile POS with essential hardware peripherals—thermal printers, barcode scanners, and cash drawers—enabling efficient physical checkout operations.

### Story 7.1: Implement Thermal Printer Support via ESC/POS Protocol

**As a** Cashier,
I**want** to print receipts on thermal printers (58mm and 80mm) using ESC/POS protocol,
So**that** customers receive professional transaction receipts and I don't struggle with printer compatibility.

**Acceptance Criteria:**

**Given** a thermal printer is connected to the mobile device
**When** a transaction is completed and receipt printing is triggered
**Then** the mobile app formats the receipt in ESC/POS format
**And** sends the formatted receipt to the printer via appropriate interface
**And** the receipt includes:
  - Pharmacy name and address (from system configuration)
  - Transaction number, date, and time
  - List of items with quantities and prices
  - Subtotal, tax, total, and change
  - Payment method
  - "Thank you" message
**And** the printer cuts the receipt after printing
**And** the system supports both 58mm and 80mm paper widths
**And** printer errors (out of paper, connection failure) are displayed to the cashier

### Story 7.2: Implement USB Barcode Scanner Integration

**As** a** Cashier,
I**want** to use USB HID barcode scanners to quickly add products to transactions,
So**that** I don't have to manually search for products and can serve customers faster.

**Acceptance Criteria:**

**Given** a USB HID barcode scanner is connected to the mobile device
**When** the cashier scans a product barcode
**Then** the mobile app receives the barcode input as keyboard events
**And** the app debounces the input to prevent duplicate scans
**And** the app queries the backend API to find the product by SKU
**And** the product is added to the cart if found
**Or** a "product not found" error is displayed if the SKU is not in the system
**And** successful scans provide visual feedback (sound/vibration)
**And** the scanner works seamlessly without requiring configuration

### Story 7.3: Implement Bluetooth Barcode Scanner Support

As a Cashier, I want the flexibility to use wireless Bluetooth barcode scanners so I'm not tethered by cables and can move freely around the pharmacy.

**Acceptance Criteria:**

**Given** a Bluetooth barcode scanner is paired with the mobile device
**When** the cashier scans a product barcode
**Then** the mobile app receives the barcode input via Bluetooth
**And** the app validates and debounces the input
**And** queries the backend API to find the product by SKU
**And** adds the product to the cart if found
**Or** displays a "product not found" error if SKU is not in system
**And** the scanner pairing process is intuitive (Bluetooth settings, discover, pair)
And**the app maintains the Bluetooth connection during POS operations

### Story 7.4: Implement Cash Drawer Control via Printer Kick

**As a** a Cashier,
I**want** to automatically open the cash drawer when cash payments are processed,
So**that** I don't have to manually open the drawer and can maintain transaction flow.

**Acceptance Criteria:**

**Given** a cash drawer is connected to the thermal printer via RJ-12 interface
**When** a cash payment is successfully processed and receipt printing is triggered
**Then** the mobile app sends the ESC/POS cash drawer kick command via the printer
**And** the cash drawer opens automatically
**And** the drawer opens at the appropriate moment in the receipt printing sequence
**And** the system logs cash drawer openings in the audit trail (transaction ID, timestamp, user)
**And** the system handles drawer open failures gracefully with user notification

## Epic 8: Offline Mode & Synchronization (Mobile)

Enable cashiers to process sales transactions without internet connectivity and automatically synchronize data when connectivity is restored, ensuring business continuity during unreliable internet outages.

### Story 8.1: Implement Local SQLite Storage for Offline Transactions

**As a** Cashier,
I**want** the mobile app to store transaction data locally when internet is unavailable,
So**that** I can continue serving customers even when internet goes down.

**Acceptance Criteria:**

**Given** the mobile app is running and internet connectivity is lost
**When** the cashier processes a transaction
**Then** the app saves the transaction data to local SQLite storage on the device
**And** the app saves transaction header data (transaction number, timestamp, cashier ID, payment method, total)
**And** the app saves transaction line items (product ID, quantity, price, subtotal)
And**the app marks the transaction as "pending_sync" status
And**the app continues to display current stock levels from the last sync (cached data)
And**the app allows product scanning and cart building with local stock data
And**the app prevents actions that require online connectivity (new user registration, multi-branch visibility)

### Story 8.2: Implement Transaction Sync Queue

As a mobile app, I need to queue offline transactions for synchronization when connectivity is restored, ensuring no data is lost and transactions are processed in the correct order.

**Acceptance Criteria:**

**Given** offline transactions have been saved to local SQLite storage
When**internet connectivity is restored
**Then**the app identifies all pending transactions in chronological order
And**queues them for synchronization with the backend API
And**each transaction in the queue is processed sequentially:
  - POST /api/v1/sync with transaction payload
  - Wait for successful response from backend
  - Mark transaction as "synced" if successful
  - Move to next transaction if successful, retry if failed
And**the app provides visual indicators of sync progress (number of pending transactions)
And**the app implements exponential backoff retry for failed syncs (retry after 1min, 2min, 4min, etc.)
And**the app sync queue is persistent across app restarts (survives app crashes)

### Story 8.3: Implement Bidirectional Data Synchronization

As a mobile app, I need to synchronize data bidirectionally when internet is restored—the backend receives offline transactions, and the mobile app gets the latest stock levels from the server.

**Acceptance Criteria:**

**Given** internet connectivity is restored
When**the sync process is triggered
Then**the app first uploads all pending offline transactions to the backend
And**the app then downloads the latest stock levels from the server
And**the app updates the local product cache with current stock quantities
And**the app downloads new products added since last sync
And**the app updates user data (profile, role, permissions) if changed
And**the app sync is incremental (only changed data since last sync, not full database dump)
And**the app implements conflict resolution for offline transactions using last-write-wins with manual override
And**the app provides visual sync status indicators (synced, pending sync, failed sync)
And**the app retries failed syncs automatically in the background

### Story 8.4: Implement Visual Sync Status Indicators

As a Cashier, I want clear visual indicators showing the synchronization status so I know whether my data is up-to-date or needs attention.

**Acceptance Criteria:**

**Given**the mobile app is running
When**sync status changes
Then**visual indicators are displayed prominently in the UI:
  - **Green checkmark (synced):** All data is up-to-date with server
  - **Yellow clock (pending):** Offline transactions waiting to sync
  - **Red exclamation (failed):** Sync failed, requires attention
And**the indicator is displayed in the app header or status bar
And**tapping the indicator shows sync details:
  - For pending: number of transactions waiting to sync
  - For failed: last error message and retry countdown
And**the app automatically retries failed syncs with exponential backoff
And**the app plays a notification when sync completes successfully

### Story 8.5: Implement Conflict Resolution for Offline Transactions

As a mobile app, I need to handle conflicts that arise when the same product is sold offline by multiple cashiers before synchronization, ensuring data integrity.

**Acceptance Criteria:**

**Given**multiple offline transactions include sales of the same product
When**syncing to the backend
Then**the backend processes transactions chronologically
And**for each transaction, the system checks if sufficient stock is available
And**if stock is sufficient, the transaction is processed normally
Or**if stock is insufficient, the transaction fails with an "insufficient stock" error
And**the failed transaction is marked in the mobile app with the specific error message
And**the app allows manual override for the transaction if needed (with admin authorization)
And**all conflict resolution attempts are logged in the audit trail with user IDs and timestamps

## Epic 9: API Foundation & Core Services

Build the REST API with Golang backend that serves mobile and web clients, implementing authentication, authorization, error handling, and core business logic services.

### Story 9.1: Implement API Health Check Endpoint

**As a** System,
I**want**to provide a /health endpoint that returns system status for monitoring and uptime tracking,
**So**that**operations teams can monitor system health and achieve 99.5% uptime target.

**Acceptance Criteria:**

**Given**the backend API is running
When**a GET request is made to /api/v1/health
**Then**the endpoint returns 200 OK status
**And**the response includes system status information:
  - status: "healthy" or "degraded"
  - database: "connected" or "disconnected"
  - redis: "connected" or "disconnected"
  - uptime: current uptime in seconds
  - version: API version number
And**the endpoint responds within 500ms for monitoring tools
And**the endpoint logs health check requests for audit purposes

### Story 9.2: Implement API Documentation with Swagger

**As a** Development Team,
I**want**auto-generated API documentation from Swagger annotations using swaggo,
So**that**API contracts are clearly documented and available for client SDK generation.

**Acceptance Criteria:**

**Given**API handlers are implemented with Swagger annotations
When**the development team runs swaggo command
**Then**Swagger generates a complete OpenAPI specification file
And**the specification includes all endpoints with:
  - HTTP methods and paths
  - Request parameters with types and validation rules
  - Response schemas with examples
  - Authentication requirements
  - Error responses (RFC 7807 format)
And**the Swagger UI is accessible at /api/docs endpoint
And**the documentation is exported to docs/swagger.yaml for version control
And**clients can generate API client SDKs from the specification

### Story 9.3: Implement Rate Limiting Middleware

**As a** System,
I**want**to implement API rate limiting to prevent abuse and ensure fair resource allocation,
So**that**the system remains stable even under heavy load and no single user can monopolize resources.

**Acceptance Criteria:**

**Given**the API is running and Gin middleware is configured
When**requests come in from authenticated users
**Then**each user is limited to 100 requests per minute
And**the rate limiter tracks requests by JWT token/user ID
And**requests exceeding the limit receive 429 Too Many Requests response
And**the response includes Retry-After header indicating when to retry
And**the rate limiter uses a sliding window algorithm for accurate rate limiting
And**rate limits are configurable per environment (can be increased for Enterprise tier)

### Story 9.4: Implement CORS Middleware for Cross-Origin Requests

**As a** System,
I**want**to implement CORS middleware so the web dashboard can call the API even if hosted on different domains,
So**that**we have deployment flexibility and don't force same-origin deployment.

**Acceptance Criteria:**

**Given**the API is running and CORS middleware is configured
When**cross-origin requests come from the web dashboard
**Then**the middleware validates the Origin header against allowed origins
And**requests from allowed origins receive appropriate CORS headers:
  - Access-Control-Allow-Origin: specific origin (not "*" for security)
  - Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
  - Access-Control-Allow-Headers: Authorization, Content-Type, etc.
And**pre-flight OPTIONS requests receive 200 OK response with appropriate CORS headers
And**actual API requests (GET, POST, etc.) proceed normally with CORS headers attached
And**CORS configuration is environment-based (localhost for development, specific domains for production)

### Story 9.5: Implement Structured Logging with Zap

As a Backend System, I want comprehensive structured logging with context so I can troubleshoot issues effectively and maintain operational visibility.

**Acceptance Criteria:**

**Given**the backend API is running with structured logging configured
When**log-worthy events occur (requests, errors, business events)
Then**the system logs events in JSON format with fields:
  - level (info, warn, error, debug)
  - timestamp (ISO 8601 with timezone)
  - caller (function name, file, line number)
  - message (descriptive message)
  - context (user_id, transaction_id, request_id, etc.)
And**logs are written to stdout for Docker container logging
And**logs are rotated daily to prevent disk space issues
And**sensitive data (passwords, tokens) is redacted from logs
And**log levels are configurable via environment variable

### Story 9.6: Implement Core Business Services

As a Development Team, I want to implement the foundational business logic services that power the application.

**Acceptance Criteria:**

**Given**the repository layer provides data access
When**implementing services
Then**core services are implemented with Clean Architecture separation:
  - AuthService: login, logout, token generation, token validation
  - UserService: create user, update user, deactivate user, list users
  - ProductService: list products, get product, create product, update stock, check availability
  - TransactionService: create transaction, process sale, calculate totals, generate receipt
  - ReportService: daily sales, profit/loss, export reports
  - AlertService: check low stock, check expiry, send notifications
  - SyncService: queue transaction sync, process sync, conflict resolution
And**services use repositories for data access (no direct database calls)
And**services return domain entities or errors (not database-specific errors)
And**business logic is well-tested and isolated from data access concerns

## Epic 10: Supplier & Purchase Management

Enable supplier management, purchase order tracking, and goods receipt processing to maintain accurate inventory costs and supplier relationships.

### Story 10.1: Implement Supplier Master Data Management

**As a** System Administrator,
**I want to** create and maintain supplier master data with contact information,
**So that** we can track our supplier relationships and communicate with them effectively.

**Acceptance Criteria:**

**Given** the system administrator is authenticated with Admin role
**When** creating a new supplier
**Then** the admin can input supplier name, contact person, phone number, email, and address
**And** the system validates required fields (name, phone)
**And** the system saves the supplier to the database with unique supplier ID
**And** the system logs supplier creation in the audit trail with admin user ID
**And** the admin can edit existing supplier information
**And** the admin can deactivate suppliers (no new purchases, maintain historical data)

### Story 10.2: Implement Purchase Invoice Recording

**As a** System Administrator or Owner,
**I want to** record purchase invoices from suppliers with item details and costs,
**So that** we can track purchases accurately and calculate cost of goods sold.

**Acceptance Criteria:**

**Given** the user is authenticated with Admin or Owner role
**When** recording a purchase invoice
**Then** the user can input invoice number, date, supplier, and invoice items
**And** each invoice item includes product, quantity, unit cost, and subtotal
**And** the system calculates total invoice amount automatically
**And** the system records the invoice with payment status set to "unpaid"
**And** the system maintains an append-only audit trail of the invoice recording
**And** the user can upload or attach invoice document images (optional)

### Story 10.3: Implement Goods Receipt Processing

**As a** System Administrator or Owner,
**I want to** process goods receipt from suppliers to increase stock and update costs,
**So that** inventory quantities are accurate and cost prices reflect latest purchases.

**Acceptance Criteria:**

**Given** a purchase invoice has been recorded
**When** goods are received from the supplier
**Then** the admin can initiate goods receipt for the invoice
**And** the system increases stock quantities for all items in the invoice
**And** the system updates product cost prices to the latest purchase cost
**And** the system marks the invoice as "received"
**And** the system logs the goods receipt in the audit trail with user ID and timestamp
**And** the system triggers stock level check for low stock notifications if applicable

### Story 10.4: Implement Supplier Payment Tracking

**As a** Pharmacy Owner,
**I want to** track supplier payment status including unpaid, partial, and fully paid invoices,
**So that** I can manage cash flow and avoid missing payment deadlines.

**Acceptance Criteria:**

**Given** unpaid purchase invoices exist in the system
**When** recording a supplier payment
**Then** the owner can select an unpaid invoice and input payment amount
**And** the system updates invoice payment status (unpaid → partial → fully paid)
**And** the system records payment date, payment method, and notes
**And** the system logs all payment transactions in the audit trail
**And** the owner can view payment history for each supplier
**And** the owner can filter invoices by payment status

### Story 10.5: Implement Supplier Product Catalog

**As a** System Administrator,
**I want to** maintain supplier product catalogs with purchase prices,
**So that** cost calculations are accurate and purchase orders can be created efficiently.

**Acceptance Criteria:**

**Given** suppliers are registered in the system
**When** managing supplier product catalogs
**Then** the admin can associate products with suppliers and specify purchase prices
**And** the system maintains current purchase price for each product-supplier combination
**And** the system uses supplier purchase prices when recording purchase invoices
**And** the system can display price history to track cost changes over time
**And** the admin can mark preferred suppliers for each product

### Story 10.6: Implement Supplier Aging Reports

**As a** Pharmacy Owner,
**I want to** generate supplier aging reports showing outstanding invoices by payment period,
**So that** I can prioritize payments and manage supplier relationships effectively.

**Acceptance Criteria:**

**Given** the pharmacy owner is logged into the web dashboard
**When** generating a supplier aging report
**Then** the report displays all unpaid and partially paid invoices grouped by supplier
**And** invoices are categorized by payment period: 0-30, 31-60, 61-90, 90+ days
**And** the report shows total outstanding amount per supplier
**And** the report can be filtered by date range and supplier
**And** the owner can export the aging report as PDF or Excel
**And** the report data is calculated from purchase invoices and payment records

### Story 10.7: Implement Supplier Transaction Audit Trail

**As a** System,
**I must** maintain complete append-only audit trail for all supplier transactions for Badan POM compliance,
**So that** all purchases, returns, and payments are traceable for regulatory inspections.

**Acceptance Criteria:**

**Given** any supplier transaction occurs (purchase, goods receipt, payment, return)
**When** the transaction is recorded in the database
**Then** the system automatically creates an immutable audit trail entry
**And** the audit entry includes:
  - Who: User ID and role who performed the action
  - When: Timestamp of the action
  - What: Description of the action (supplier created, invoice recorded, payment made, etc.)
  - Why: Reason for the action (if applicable)
  - How much: Transaction amount and affected items
**And** the audit entry is append-only (no modifications or deletions allowed)
**And** audit entries are queryable for at least 5 years per Badan POM requirements
**And** audit logs can be exported for compliance inspections

---

*End of Requirements Extraction*
