---
stepsCompleted: ["step-01-init", "step-02-discovery", "step-02b-vision", "step-02c-executive-summary", "step-03-success", "step-04-journeys", "step-05-domain", "step-06-innovation (skipped - no innovation signals detected)", "step-07-project-type", "step-08-scoping", "step-09-functional", "step-10-nonfunctional", "step-11-polish"]
inputDocuments: ["product-brief-simpo.md", "product-brief-simpo-distillate.md"]
workflowType: 'prd'
briefCount: 2
researchCount: 0
brainstormingCount: 0
projectDocsCount: 0
projectType: 'greenfield'
releaseMode: 'phased'
classification:
  projectType: 'SaaS B2B + Mobile App + Web App'
  domain: 'Healthcare (Pharmacy)'
  complexity: 'High'
  projectContext: 'Greenfield'
---

# Product Requirements Document - simpo

**Author:** Shankara
**Date:** 2026-05-08

---

## Executive Summary

**simpo** is a cost-effective pharmacy management system designed for Indonesian small to medium-sized pharmacies operating 2-5 branches. It addresses a critical market gap: existing solutions like Farmacare provide complete functionality but at subscription costs (IDR 3-10M/month) that strain thin pharmacy margins.

The system delivers essential operational capabilities—point-of-sale, real-time inventory management, financial reporting, and purchase invoicing—through a modern technology stack (Golang backend + React Native frontend). The immediate use case is internal deployment at the owner's pharmacy, with a clear path to commercialize as an affordable SaaS offering (<IDR 1M/month) once validated.

**Target Users:** Pharmacy owners with 2-5 branches who have adopted digital systems but are experiencing margin pressure from high subscription fees. Secondary users include cashiers requiring fast, reliable POS and system administrators managing multi-location operations.

**Core Problem:** Pharmacy owners face a paradox—digital tools increase operational efficiency, but subscription costs erode the margins they're meant to protect. For a 3-branch pharmacy paying IDR 5M/month, software fees total IDR 60M annually (equivalent to 1-2 staff salaries).

**Solution Approach:** Deliver 80% of Farmacare's functionality at 20% of the cost through self-hosted deployment and modern tech stack efficiency. Built and validated on a real pharmacy before commercialization, ensuring genuine customer proximity and rapid iteration.

### Differentiation

simpo competes on value, not features. While competitors focus on enterprise complexity or basic POS tools, simpo targets the mid-market with a pragmatic feature set focused on daily-used operations.

**Core Value Proposition:** Pharmacy owners don't want fewer features—they want the *same* features at a price that respects their margins. A self-hosted or affordably-priced SaaS alternative, built with modern technology, can deliver equivalent functionality at significantly lower cost.

**Competitive Advantages:**
- **Real-world validation** — Built for and tested on an actual pharmacy before commercializing
- **Tech stack efficiency** — Golang's performance enables lower infrastructure costs than legacy stacks
- **No investor pressure** — Bootstrapped approach means sustainable pricing without growth-at-all-costs pressure
- **Local expertise** — Deep understanding of Indonesian pharmacy workflows and Badan POM regulatory requirements

**Honest Assessment:** The moat isn't unique technology—pharmacy management is a well-understood domain. The advantage is execution speed, lower cost structure, and genuine customer proximity. Success comes from reliability, responsiveness, and price.

---

## Project Classification

**Project Type:** SaaS B2B + Mobile App + Web App
- Multi-tenant pharmacy platform with mobile POS (Android-first) and web admin dashboard
- Self-hosted deployment with Docker containerization

**Domain:** Healthcare (Pharmacy)
- Regulated industry requiring Badan POM compliance
- Patient data privacy and medication safety considerations
- Audit trail and reporting requirements

**Complexity:** High
- Regulatory compliance (Badan POM, future BPJS integration)
- Multi-branch data synchronization
- Offline mode architecture for unreliable connectivity
- Hardware integration (thermal printers, barcode scanners, cash drawers)
- Real-time inventory management across locations

**Project Context:** Greenfield
- New product development from scratch
- Replacing existing competitor solution (Farmacare)
- First deployment at owner's pharmacy as proof-of-concept

---

## Success Criteria

### User Success

**Primary Success Moment:** Pharmacy owner realizes operational cost reduction without compromising functionality—the moment they review monthly reports and see 100% software cost savings while staff efficiently process sales and inventory is accurately tracked.

**User Success Metrics:**
- Staff adoption rate >90% within 3 months of internal deployment (cashiers using system daily)
- Transaction processing time <30 seconds (parity with Farmacare)
- Stock reconciliation accuracy >99% (reduced stockouts and overstock situations)
- Low stock and expiry alerts received and acted upon before business impact
- Financial reports generated without manual intervention for accountant meetings

**Emotional Success Indicators:**
- Relief from eliminating monthly subscription fees
- Confidence in multi-branch inventory visibility
- Control over business operations through real-time data

### Business Success

**Internal Deployment Success (First 6 Months):**
- Successfully replace Farmacare without operational disruption
- All 11 core features functional and used daily across branches
- Monthly subscription savings: 100% (zero external fees)
- Zero critical incidents requiring rollback to previous system

**SaaS Commercialization Success (Year 2 onwards):**
- 10 paying pharmacies within 12 months of launch
- Monthly churn rate <5%
- Average revenue per pharmacy: IDR 500K-1M/month
- Customer acquisition cost (CAC) <3 months of revenue
- Net Promoter Score (NPS) >40

**Leading Indicators:**
- System uptime >99.5%
- Support response time <4 hours during business hours
- Time to process sale: <30 seconds
- Stock reconciliation accuracy: >99%

### Technical Success

**Performance Requirements:**
- POS transaction processing: <30 seconds end-to-end
- Real-time stock synchronization across branches: <5 seconds latency
- Report generation: <10 seconds for monthly reports
- System uptime: >99.5% (excluding planned maintenance)

**Reliability & Data Integrity:**
- Zero data loss incidents
- Automatic daily backups with 30-day retention
- Audit trail for all transactions and inventory changes
- Failed transaction rate <0.1%

**Compliance & Security:**
- Badan POM regulatory compliance for pharmacy operations
- Role-based access control (Admin, Owner, Cashier) properly enforced
- Secure authentication with session management
- Data encryption at rest and in transit
- Offline mode capability for unreliable internet connectivity

### Measurable Outcomes

**Quantitative Targets:**
| Metric | Target | Timeline |
|--------|--------|----------|
| Staff Adoption Rate | >90% | 3 months post-deployment |
| Transaction Time | <30 seconds | Day 1 of operations |
| Stock Accuracy | >99% | Ongoing |
| System Uptime | >99.5% | Ongoing |
| Support Response | <4 hours | Business hours |
| Paying Customers (SaaS) | 10 pharmacies | 12 months post-launch |
| Monthly Churn | <5% | Ongoing |
| ARPU | IDR 500K-1M | Year 2 onwards |

**Qualitative Outcomes:**
- Pharmacy owners report "I can't believe we were paying so much for this"
- Staff report "this is faster/easier than our old system"
- Zero requests to revert to previous system after 90 days
- Word-of-mouth referrals from early adopters

---

Product scope defines what will be delivered across the product lifecycle.

## Product Scope

### MVP - Minimum Viable Product

**Core Capabilities (Must-Have for Internal Deployment):**

1. **Authentication & Access Control** — Login system with secure authentication, multi-user roles (System Admin, Owner, Cashier), staff registration via admin approval or whitelist

2. **Point-of-Sale (Kasir)** — Transaction processing with receipt printing, payment method handling (cash, transfer, e-wallet), thermal printer integration

3. **Inventory Management (Cek Stok)** — Real-time stock visibility across branches, stock adjustment capabilities, low stock notifications with reorder triggers

4. **Financial Reporting (Laporan Keuangan)** — Daily, weekly, monthly profit/loss reports, sales summaries by branch and period, export capabilities for accountant use

5. **Purchase Invoicing (Faktur Pembelian)** — Supplier management, purchase order creation and tracking, invoice recording and payment status

6. **Smart Alerts** — Low stock notifications with reorder points, expiry date alerts (30-day, 14-day, 7-day advance warning)

7. **Multi-Branch Support** — Branch setup and configuration, cross-branch stock visibility, consolidated reporting

**Technical Infrastructure:** Golang backend (monolith), React Native mobile app (Android-first), web admin dashboard, PostgreSQL database, self-hosted Docker deployment, offline mode for unreliable connectivity

### Growth Features (Post-MVP)

**Competitive Enhancements:** iOS mobile app support, multi-branch inventory transfer, advanced analytics and forecasting, API integrations (Jurnal, Accurate), automated backup and disaster recovery, mobile customer app, enhanced reporting with custom report builder

**Go-to-Market Readiness:** Managed hosting option, onboarding automation and setup wizard, customer support ticketing system, usage analytics and product telemetry

### Vision Features (Year 3+)

**Advanced Capabilities:** E-prescription integration (BPJS, government systems), supplier marketplace and direct ordering, AI-powered demand forecasting, automated regulatory compliance reporting, franchise pharmacy management, regional expansion beyond Indonesia

**Strategic Positioning:** Default choice for independent Indonesian pharmacies, 100+ pharmacy customers, regional product with local expertise advantage, sustainable profitable business without VC pressure

---

Understanding success metrics provides the foundation for measuring product value and business viability.

---

## User Journeys

### Journey 1: Budi (Pharmacy Owner) — "The Liberation Journey"

**Opening Scene:**
Budi owns 3 pharmacies in Jakarta. Every month, he transfers IDR 15M to Farmacare for software subscriptions—that's IDR 180M annually. He's feeling trapped. The system works, sure, but he's watching his margins shrink. His accountant asks for financial reports, and he has to navigate multiple screens to export them. Yesterday, a customer complained about a medicine being out of stock—again.

**Rising Action:**
Budi decides to try simpo. His IT guy sets it up on the local server (self-hosted!). The first day, Budi logs into the web dashboard from his home. He sees consolidated sales across all 3 branches in real-time. He notices Branch 2 is running low on a popular antibiotic—he gets a notification. He calls Branch 2 manager: "Order 50 units tomorrow morning."

**Climax:**
End of month arrives. Budi's accountant asks for reports. Instead of the usual 2-hour export nightmare, Budi clicks "Monthly Report"—30 seconds later, it's ready. He hands it over. The accountant looks up: "This is cleaner than usual."

But the real moment: Budi checks his bank account. No IDR 15M deduction. Zero software fees. For the first time in years, he feels like his margins are his own.

**Resolution:**
Six months later, Budi can't imagine going back. His staff adopted the system within weeks—turns out, the interface is simpler than Farmacare. Stockouts dropped 70% because of the alerts. He's finally seeing the profits his hard work generates.

### Journey 2: Siti (Cashier) — "The Flow State Journey"

**Opening Scene:**
Siti's been a cashier at Apotek Sehat for 3 years. Peak hour (5-7 PM) is chaos—customers lining up, phones ringing, everyone's impatient. The old POS system? Laggy. Sometimes it freezes mid-transaction. Customers sigh. Siti sweats. She hates peak hour.

**Rising Action:**
New system day. Siti's skeptical—change always means learning curves. She logs in: clean interface. First customer approaches. Siti scans items... the system responds instantly. She hits print—receipt comes out smooth. Next customer. Same flow. No lag. No freeze.

**Climax:**
Third week in, peak hour hits. Siti's in the zone. Scan, scan, pay, print. Scan, scan, pay, print. She's processing customers 30% faster than before. A customer remarks: "You're fast today!" Siti smiles: "New system helps."

**Resolution:**
Siti doesn't think about the system anymore. It's just... there. Working. Fast. Reliably. That night, she tells her husband: "Work's easier now." That's all that matters.

### Journey 3: Dian (System Admin) — "The Control Tower Journey"

**Opening Scene:**
Dian manages IT for 5 pharmacy branches. Her boss (Budi) wants to switch from Farmacare to save money. Dian's worried: self-hosted? Who handles updates? What if something breaks? She's seen failed migrations before.

**Rising Action:**
Migration day. Dian has prepared rollback plans. She watches the data sync from Farmacare to simpo. Expected 4 hours. Actually takes 3. She verifies: products, customers, transaction history—all there. Not a single record lost.

She sets up user accounts: Budi (Owner), 3 branch managers, 12 cashiers. Role-based access—each role sees only what they need. Clean.

**Climax:**
Week 2, Branch 3's internet goes down. Dian holds her breath... but transactions keep processing. Offline mode kicks in automatically. When internet returns, data syncs seamlessly. Dian breathes. The system handled it.

**Resolution:**
Month 1 ends. Dian checks system health dashboard: 99.8% uptime. Zero critical incidents. Support response? She emailed about a minor question—reply came in 2 hours. Dian's no longer worried. The system just... works.

### Journey Requirements Summary

**From Budi's Journey (Pharmacy Owner):**
- Multi-branch consolidated dashboard with real-time sales visibility
- Stock level notifications with reorder triggers
- One-click financial report generation (daily/weekly/monthly)
- Mobile-responsive web dashboard for remote access
- Cross-branch stock visibility and management

**From Siti's Journey (Cashier):**
- Fast, lag-free POS interface with sub-30-second transaction processing
- Instant receipt printing with thermal printer integration
- Intuitive interface minimizing training requirements
- Reliable performance during peak hours with no system freezes

- **From Dian's Journey (System Admin):**
  - Data migration from existing systems (Farmacare) with validation
  - Role-based access control configuration (Admin, Owner, Cashier)
  - System health monitoring dashboard with uptime metrics
  - Offline mode with automatic data synchronization
  - Responsive support channel with <4 hour response time

---

The healthcare pharmacy domain imposes specific regulatory, compliance, and technical requirements beyond typical business applications.

## Domain-Specific Requirements

### Compliance & Regulatory

**Badan POM Compliance Requirements:**
- Immutable audit trail for all inventory transactions (purchase, sale, return, adjustment, disposal)
- Complete expiry date tracking with 30/14/7-day advance alerts
- Controlled substances: unbroken chain-of-custody from purchase to sale
- Digital documentation for expired item disposal (witness signature, photo evidence)
- Data retention: minimum 5 years for transaction history
- Tamper-evident backup storage with recovery capability

**Audit Trail Requirements:**
- Every inventory change must log: who, when, what, why (4 W's)
- Append-only log structure (no deletion, no modification after storage)
- User authentication mandatory for all system actions
- Mandatory reason field for all manual adjustments

**Report Generation for Audits:**
- Monthly sales reports by product and category
- Purchase invoices with supplier documentation
- Stock adjustment reports with before/after values and reasons
- Expiry reports showing items expiring within period
- Physical count reconciliation reports (before/after, discrepancies)
- Export formats: PDF, Excel, CSV

**Tax Audit Support:**
- Transaction-level tax calculation records
- Discount, promotion, and return documentation
- Multi-branch consolidated financial reports
- Reconciliation between physical stock and system stock

### Technical Constraints

**Data Privacy & Security:**
- Patient information encryption (if prescription data captured)
- Role-based access control enforcing least privilege
- Secure authentication with session management
- Audit logging for all data access

**Medication Safety:**
- Drug interaction checking (future phase consideration)
- Expiry date enforcement (block sale of expired medications)
- Dosage validation rules (future phase consideration)

### Integration Requirements

**Hardware Integration:**
- Thermal printer support via ESC/POS protocol
- Barcode scanner integration (USB, Bluetooth)
- Cash drawer interface control
- Label printer for barcode/sticker generation

**External Systems (Future):**
- Accounting software: Jurnal, Accurate, Zoho Books API integration
- Payment gateway: e-wallet, QRIS, debit card processing
- Supplier systems: PBF ordering platforms
- BPJS integration: e-prescription format standards (Year 3+)

### Risk Mitigations

**Regulatory Audit Failure Prevention:**
- All manual stock adjustments require documented reason + supervisor approval
- Discrepancy reports automatically generated when physical ≠ system stock
- Expiry disposal workflow with required witness signatures and photo upload
- Read-only audit mode for compliance verification without data alteration
- Hash-based data integrity verification (tamper detection)
- Annual compliance review checklist with built-in reminders

**Medication Error Liability:**
- Expiry date blocking (prevent sale of expired medications)
- Stock validation at transaction time
- Dispensing workflow validation (future: prescription verification)

**Data Integrity Assurance:**
- Append-only audit trail prevents historical record modification
- Cryptographic hashing for tamper detection
- Automated daily backups with 30-day retention
- Disaster recovery testing quarterly

**System Availability Risk:**
- Offline mode ensures business continuity during internet outages
- Automatic data synchronization when connectivity restored
- <4 hour support response commitment during business hours
- 99.5% uptime SLA with penalty provisions (future SaaS)

---

SaaS B2B + Mobile + Web App hybrid architecture introduces specific tenant, permission, integration, and platform requirements.

## Project-Type Specific Requirements

### Project-Type Overview

simpo is a hybrid product combining **SaaS B2B** (multi-branch pharmacy management), **Mobile App** (Android POS for cashiers), and **Web App** (admin dashboard for owners/admins). The product operates as a single-tenant self-hosted deployment for MVP, with potential future SaaS commercialization offering managed hosting options.

### Tenant Model

**Architecture Approach: Single-Tenant with Hybrid Future Support**
- Initial deployment: Single-tenant self-hosted per pharmacy (MVP)
- Customer manages own server/environment (on-premise or cloud infrastructure)
- Data isolation guaranteed by architecture (no shared customer data)
- Future SaaS phase: Evaluate multi-tenant cloud option for managed service tier

**Deployment Options:**
- **Self-hosted Docker deployment** (primary option for MVP)
- Customer provides server infrastructure (specifications: 4GB RAM, 2 CPU cores, 50GB storage minimum)
- Automated backup and restore procedures with 30-day retention
- System update mechanism: Manual for MVP with scripts, automated for future versions
- Rollback capability for failed updates

**Multi-Branch Architecture:**
- Single tenant supports multiple branches under one pharmacy entity
- Branch-level data segregation within single database
- Cross-branch visibility for owners and system admins
- Branch-specific access control for cashiers and managers

### Permission Matrix (RBAC)

**Role Definitions:**

| Role | Description | Permissions | Branch Access |
|------|-------------|-------------|---------------|
| **System Admin** | IT/Operations manager | Full system configuration, user management, branch setup, system health monitoring | All branches |
| **Owner** | Pharmacy owner/executive | Business oversight, all reports, financial data, staff performance | All branches |
| **Cashier** | Frontline staff | POS operations, sales processing, receipt printing, cash handling | Assigned branch only |
| **Manager** (Future) | Branch supervisor | Branch management, staff supervision, local reports | Assigned branch only |
| **Warehouse** (Future) | Inventory staff | Stock management, purchase orders, receiving | All branches |

**Permission Granularity:**
- Branch-level access control for cashiers and managers (isolated to assigned branch)
- Global cross-branch access for owners and system admins
- Audit trail for all permission changes and role assignments
- Role assignment via admin approval or whitelist (email domain validation)
- Session management with automatic timeout after 8 hours of inactivity

**Permission Enforcement:**
- All API calls validate user role and branch access
- UI hides features/functions based on role permissions
- Critical actions (stock adjustment, price changes) require supervisor approval for non-admin roles

### Integration List

**Core Integrations (MVP):**

*Hardware Integration:*
- **Thermal Printers:** ESC/POS protocol support (58mm, 80mm receipt printers)
- **Barcode Scanners:** USB HID and Bluetooth barcode scanner compatibility
- **Cash Drawers:** RJ-12 interface control via printer kick command
- **Label Printers:** Barcode/sticker label generation for internal inventory marking

*Data Export:*
- **Financial Reports:** PDF and Excel export for accountant use
- **Inventory Reports:** CSV export for spreadsheet analysis
- **Sales Data:** Export for tax preparation and business analysis

**Future Integrations (Growth Phase):**

*Accounting Software:*
- **Jurnal:** API integration for automated sales and purchase journal entries
- **Accurate:** Similar integration capabilities
- **Zoho Books:** International market expansion support

*Payment Processing:*
- **QRIS:** Indonesian QR code payment standard
- **E-wallets:** GoPay, OVO, Dana, ShopeePay integration
- **Debit Cards:** EDC machine integration for card payments

*Supplier Systems:*
- **PBF (Pharmaceutical Wholesaler) Platforms:** Direct ordering integration
- **Supplier Catalogs:** Product information and pricing synchronization

*Government Systems (Vision Phase):*
- **BPJS:** E-prescription format standards and claim submission
- **Badan POM:** Regulatory reporting interfaces

### Platform Requirements

**Mobile App (Android POS):**
- **Minimum SDK:** Android 8.0 (API 26) for broad device compatibility
- **Target SDK:** Android 14 (API 34) for latest security standards
- **Form Factor:** Phone-first design, tablet-compatible
- **Orientation:** Portrait mode for cashier ergonomics
- **Network:** Offline-first architecture with background sync when online

**Web App (Admin Dashboard):**
- **Browser Support:** Chrome 90+, Edge 90+, Firefox 88+, Safari 15+
- **Responsive Design:** Desktop (1920px+), tablet (768-1024px), mobile (<768px)
- **Performance:** Initial page load <3 seconds, subsequent interactions <500ms
- **Accessibility:** WCAG 2.1 Level A compliance (minimum), keyboard navigation support

**Self-Hosted Deployment:**
- **Containerization:** Docker Compose for simplified deployment
- **Database:** PostgreSQL 14+ with connection pooling
- **Caching:** Redis for session management and real-time updates
- **Reverse Proxy:** Nginx for SSL termination and static asset serving
- **Monitoring:** Health check endpoint for uptime monitoring

### Offline Mode Strategy

**Offline Capabilities:**
- **Transaction Processing:** Full POS functionality available without internet
- **Inventory Lookup:** Real-time stock visibility from last sync state
- **Customer Data:** Customer information cached locally
- **Receipt Printing:** Local print queue with server sync when online

**Data Synchronization:**
- **Bidirectional Sync:** When connectivity restored, push offline transactions to server
- **Conflict Resolution:** Last-write-wins with manual override for disputed records
- **Sync Indicators:** Visual UI indicators showing sync status (synced, pending, failed)
- **Background Sync:** Automatic retry with exponential backoff for failed syncs

**Offline Limitations:**
- **Multi-Branch Visibility:** Branch stock levels not available offline
- **Reporting:** Real-time reports unavailable, use cached data
- **User Management:** New user registration requires online connectivity

### Subscription Tiers (Future SaaS)

**Tier Structure (Year 2+):**

| Tier | Price (IDR/month) | Features | Deployment |
|------|-------------------|----------|-------------|
| **Basic** | 500K | Core POS, inventory, reports, 2 branches | Self-hosted |
| **Professional** | 750K | All Basic + 5 branches, priority support, automated updates | Self-hosted |
| **Enterprise** | 1M | All Professional + unlimited branches, managed hosting, custom integrations | Cloud-hosted |

**Tier Differentiation:**
- Branch count limits per tier
- Support SLA: <8 hours (Basic), <4 hours (Pro), <2 hours (Enterprise)
- Update frequency: Monthly (Basic), Bi-weekly (Pro), Weekly (Enterprise)
- Managed hosting option for Enterprise tier

---

Phased development balances risk reduction with speed-to-market by delivering value incrementally.

## Project Scoping & Phased Development

### MVP Strategy & Philosophy

**MVP Approach:** Problem-Solving MVP — Validate core value proposition (cost reduction + operational efficiency) through focused single-branch implementation before scaling to multi-branch complexity.

**Resource Requirements:**
- **Team Minimum:** 2-3 developers (1 backend Golang, 1 frontend React Native, 1 part-time web dashboard)
- **Timeline Phase 1:** 3-4 months for single-branch MVP
- **Validation Strategy:** Internal deployment at owner's pharmacy as proof-of-concept before external commercialization

**Rationale:** Fased approach de-risks high-complexity domain (healthcare + multi-branch + offline sync) by validating core functionality in simpler single-branch context first.

### MVP Feature Set (Phase 1) - Single Branch

**Focus:** Replace Farmacare for 1 pharmacy location with validated core value.

**Core User Journeys Supported:**
- **Siti (Cashier):** Fast POS transaction processing, receipt printing, flow state during peak hours
- **Budi (Owner):** Daily sales reports, stock visibility, cost savings realization
- **Basic Dian (Admin):** User management, system configuration, health monitoring

**Must-Have Capabilities (9 Features):**

1. **Login & Authentication**
   - Secure user authentication with session management
   - Role-based access control: System Admin, Owner, Cashier
   - Staff registration via admin approval or whitelist

2. **Single-Branch POS/Kasir**
   - Transaction processing with receipt printing
   - Payment method handling: cash, transfer, e-wallet
   - Thermal printer integration (ESC/POS protocol)
   - Sub-30 second transaction processing time

3. **Stock Management (Single Location)**
   - Real-time stock visibility and tracking
   - Stock adjustment capabilities with reason logging
   - Low stock notifications with reorder triggers

4. **Daily Financial Reports**
   - End-of-day sales summary
   - Basic profit/loss report (daily view)
   - Export to PDF/Excel for accountant use

5. **Low Stock Alerts**
   - Automatic notifications when stock falls below reorder point
   - Dashboard alerts for at-risk items
   - Configurable reorder thresholds per product

6. **Expiry Date Alerts**
   - 30-day, 14-day, 7-day advance warnings for expiring medications
   - Dashboard view of upcoming expirations
   - Prevention of sales for expired items

7. **Staff Registration**
   - Admin creates cashier accounts
   - Whitelist option for email domain validation
   - Role assignment with permission enforcement

8. **Basic Print Capabilities**
   - Receipt printing for transactions
   - Simple daily report printing
   - Thermal printer and A4 printer support

9. **Offline Mode (Basic)**
   - Transaction processing without internet
   - Stock lookup from last sync state
   - Automatic data synchronization when connectivity restored

**Out of Scope for Phase 1:**
- ❌ Multi-branch support and cross-branch features
- ❌ Purchase invoicing system (manual process)
- ❌ iReport custom format reports
- ❌ Advanced financial reports (monthly, quarterly, annual)
- ❌ Branch comparison and consolidated analytics
- ❌ Multi-branch inventory transfers

**Validation Criteria:**
- Staff adoption rate >90% within 3 months
- Transaction time <30 seconds (parity with Farmacare)
- Zero requests to revert to Farmacare after 90 days
- Owner reports tangible operational improvement
- Successfully replace Farmacare without business disruption

### Post-MVP Features (Phase 2) - Multi-Branch Expansion

**Focus:** Scale to full 2-5 branch capability with competitive parity to Farmacare.

**New Capabilities:**

1. **Multi-Branch Support**
   - Branch setup and configuration for multiple locations
   - Cross-branch stock visibility and monitoring
   - Branch-specific user assignment and access control

2. **Cross-Branch Features**
   - Consolidated reporting across all branches
   - Branch performance comparison analytics
   - Owner-level dashboard with multi-branch overview

3. **Purchase Invoicing System**
   - Supplier management and master data
   - Purchase order creation and tracking
   - Invoice recording and payment status tracking
   - Supplier performance analytics

4. **Advanced Financial Reports**
   - Monthly, quarterly, and annual financial statements
   - Multi-branch profit/loss analysis
   - Tax-ready export formats
   - Custom date range reporting

5. **iReport (Custom Reports)**
   - User-defined report formats and templates
   - Badan POM compliance report formats
   - Audit trail reports for regulatory inspections
   - Scheduled automated report generation

6. **Enhanced Offline Sync**
   - Multi-branch conflict resolution for offline transactions
   - Advanced sync queue management and retry logic
   - Sync status indicators and error handling

7. **iOS Mobile App**
   - iPhone and iPad support for iOS ecosystem
   - App Store submission and compliance

**Validation Criteria:**
- Successfully operating 2-5 branches with synchronized data
- All Phase 1 features working seamlessly across branches
- Owner reports "can't imagine going back to old system"
- System ready for SaaS commercialization to external pharmacies

### Growth Features (Phase 3) - SaaS Commercialization

**Focus:** Launch to 10 external pharmacies and validate business model.

**New Capabilities:**

1. **Managed Hosting Option**
   - Cloud-hosted alternative to self-hosted deployment
   - Automated backups and disaster recovery
   - Managed updates and maintenance

2. **Onboarding Automation**
   - Setup wizard for new pharmacy configuration
   - Data migration automation from competitor systems
   - Training materials and in-app guidance

3. **Customer Support System**
   - Ticket-based support management
   - Knowledge base and FAQ system
   - Support analytics and response time tracking

4. **Accounting Software Integrations**
   - Jurnal API integration for automated journal entries
   - Accurate accounting software compatibility
   - Zoho Books integration for international expansion

5. **Usage Analytics & Telemetry**
   - Product usage metrics and feature adoption tracking
   - Performance monitoring and error reporting
   - Business intelligence for product improvement

6. **Subscription Management**
   - Billing and payment processing for SaaS subscriptions
   - Tier management (Basic, Professional, Enterprise)
   - Customer self-service portal for account management

7. **Advanced Analytics**
   - Demand forecasting and inventory optimization
   - Sales trend analysis and seasonality detection
   - Customer behavior insights

**Validation Criteria:**
- 10 paying pharmacies within 12 months of launch
- Monthly churn rate <5%
- ARPU: IDR 500K-1M/month
- Customer acquisition cost <3 months of revenue
- Net Promoter Score >40

### Vision Features (Phase 4) - Future Expansion

**Focus:** Advanced integrations, AI capabilities, and regional market expansion.

**New Capabilities:**

1. **BPJS E-Prescription Integration**
   - Electronic prescription format compliance
   - Insurance claim submission automation
   - Government health system data exchange

2. **Supplier Marketplace**
   - Direct ordering integration with PBF systems
   - Supplier catalog synchronization
   - Automated procurement workflows

3. **AI-Powered Features**
   - Intelligent demand forecasting for inventory optimization
   - Automated stock replenishment recommendations
   - Price optimization based on demand elasticity

4. **Automated Compliance**
   - Regulatory report automation for Badan POM
   - Audit trail generation for compliance inspections
   - Tax compliance reporting automation

5. **Franchise Management**
   - Franchise pharmacy operational oversight
   - Centralized inventory management for franchise networks
   - Performance benchmarking across franchise locations

6. **Regional Expansion**
   - Localization for Southeast Asian markets
   - Multi-currency and multi-language support
   - Regional regulatory compliance adaptations

**Strategic Positioning:**
- Default choice for independent Indonesian pharmacies
- 100+ pharmacy customers by Year 3
- Regional product with local expertise advantage
- Sustainable, profitable business without VC pressure

### Risk Mitigation Strategy

**Technical Risks:**

*Offline Sync Complexity:*
- **Mitigation:** Phase 1 single-branch approach eliminates cross-branch sync complexity. Validate offline mode in simpler context before adding multi-branch conflict resolution in Phase 2.

*Data Migration from Farmacare:*
- **Mitigation:** Phase 1 focuses on single location with manageable data volume. Develop robust migration scripts with validation and rollback plans before scaling to multi-branch.

*Hardware Compatibility:*
- **Mitigation:** Phase 1 validates core hardware integration (thermal printers, barcode scanners) in real-world environment. Establish hardware compatibility lab before adding more devices in later phases.

**Market Risks:**

*Value Proposition Not Validated:*
- **Mitigation:** Phase 1 internal deployment serves as real-world validation. Owner's pharmacy acts as proof-of-concept before external SaaS launch. Staff adoption rate >90% validates value proposition.

*Competitor Price Response:*
- **Mitigation:** Phase 1 speed-to-market (3-4 months) beats typical enterprise development cycles. First-mover advantage in cost-sensitive mid-market segment before competitors respond.

*Customer Switching Resistance:*
- **Mitigation:** Phase 1 proves switching is possible with minimal business disruption. Zero-downtime migration approach with parallel running period reduces perceived risk.

**Resource Risks:**

*Limited Team Size:*
- **Mitigation:** Phase 1 scope designed for 2-3 developers. Single-branch focus reduces architectural complexity. Monolithic backend initially (microservices when scaled).

*Feature Creep During Phase 1:*
- **Mitigation:** Strict scope gate with only 9 must-have features. Any additional requests deferred to Phase 2. Clear phase boundaries prevent scope expansion.

*Extended Timeline:*
- **Mitigation:** Phase 1 has clear 3-4 month target with simplified scope. Single-branch requirement reduces testing complexity. Regular milestone checkpoints to catch delays early.

---

Functional requirements define specific system capabilities that deliver the product scope.

## Functional Requirements

### Authentication & User Management

- FR1: System Administrator can create new user accounts with role assignment (Admin, Owner, Cashier)
- FR2: System Administrator can deactivate existing user accounts
- FR3: Users can authenticate into the system using username and password credentials
- FR4: System can enforce role-based access control, restricting users to actions permitted for their assigned role
- FR5: System can terminate user sessions after 8 hours of inactivity for security

### Point of Sale (POS)

- FR6: Cashier can process sales transactions by scanning product barcodes or manual item entry
- FR7: Cashier can apply payment methods including cash, bank transfer, and e-wallet for each transaction
- FR8: Cashier can generate and print transaction receipts using thermal printer
- FR9: System can calculate transaction totals including item prices, discounts, and taxes
- FR10: System can complete end-to-end sales transactions within 30 seconds

### Inventory Management

- FR11: Owner can view current stock levels for all products in real-time
- FR12: Cashier can check product availability during sales transactions
- FR13: System Administrator can manually adjust stock quantities with required reason logging
- FR14: System can automatically deduct sold items from inventory during sales transactions
- FR15: System can maintain stock reconciliation accuracy greater than 99%

### Alerts & Notifications

- FR16: System can automatically generate low stock notifications when product quantity falls below configurable reorder threshold
- FR17: System can automatically generate expiry date alerts at 30-day, 14-day, and 7-day intervals before expiration
- FR18: Owner can view aggregated alerts and notifications in a centralized dashboard
- FR19: System can prevent sale of expired medications

### Financial Reporting

- FR20: Owner can generate daily sales summary reports showing total sales, payment methods, and top-selling products
- FR21: Owner can generate basic profit/loss reports showing revenue, cost of goods sold, and gross profit
- FR22: Owner can export financial reports in PDF and Excel formats
- FR23: System can maintain complete audit trail of all financial transactions for compliance purposes

### System Administration

- FR24: System Administrator can configure system settings including business name, address, and contact information
- FR25: System Administrator can monitor system health and uptime through admin dashboard
- FR26: System can maintain complete append-only audit trail logging all system changes with user, timestamp, and reason
- FR27: System can automatically perform daily backups of all data with 30-day retention

### Hardware Integration

- FR28: System can interface with thermal printers supporting ESC/POS protocol for receipt printing
- FR29: System can interface with USB and Bluetooth barcode scanners for product scanning
- FR30: System can control cash drawers via printer kick command
- FR31: System can generate barcode/sticker labels using label printer

### Offline Mode & Synchronization

- FR32: Cashier can process sales transactions without internet connectivity
- FR33: System can queue offline transactions for synchronization when connectivity is restored
- FR34: System can automatically synchronize data when internet connectivity resumes
- FR35: System can provide visual indicators showing synchronization status (synced, pending, failed)

---

Non-functional requirements define quality attributes that ensure system reliability, performance, security, and scalability.

## Non-Functional Requirements

### Performance

**Transaction Processing:**
- NFR-PERF-001: System shall complete end-to-end sales transactions within 30 seconds from product scan to receipt print
- NFR-PERF-002: System shall process product barcode scans and return product information within 1 second
- NFR-PERF-003: System shall generate daily sales summary reports within 10 seconds
- NFR-PERF-004: System shall synchronize offline transactions to server within 5 seconds of connectivity restoration

**User Interface Responsiveness:**
- NFR-PERF-005: System shall load admin dashboard within 3 seconds on standard broadband connection
- NFR-PERF-006: System shall respond to user interactions (button clicks, form submissions) within 500 milliseconds
- NFR-PERF-007: System shall support concurrent usage by 5 cashiers across multiple terminals with <2 second response degradation

**Data Processing:**
- NFR-PERF-008: System shall maintain stock reconciliation accuracy greater than 99% during ongoing operations

### Security

**Authentication & Access Control:**
- NFR-SEC-001: System shall enforce role-based access control with three roles: System Admin, Owner, Cashier
- NFR-SEC-002: System shall terminate user sessions after 8 hours of inactivity
- NFR-SEC-003: System shall require strong password storage using industry-standard hashing algorithms
- NFR-SEC-004: System shall maintain append-only audit trail for all system changes with user identification, timestamp, and reason

**Data Protection:**
- NFR-SEC-005: System shall encrypt all data at rest using AES-256 encryption or equivalent
- NFR-SEC-006: System shall encrypt all data in transit using TLS 1.3 or higher
- NFR-SEC-007: System shall support daily automated backups with 30-day retention
- NFR-SEC-008: System shall prevent data deletion for audit compliance (append-only logs)

**Compliance:**
- NFR-SEC-009: System shall maintain complete audit trail for all inventory transactions (purchase, sale, adjustment, disposal) for minimum 5 years
- NFR-SEC-010: System shall provide read-only audit mode for compliance verification without data alteration capabilities
- NFR-SEC-011: System shall prevent sale of expired medications through blocking logic at transaction time

### Reliability

**System Availability:**
- NFR-REL-001: System shall achieve 99.5% uptime excluding planned maintenance windows
- NFR-REL-002: System shall provide automatic data backup every 24 hours with 30-day retention
- NFR-REL-003: System shall maintain failed transaction rate less than 0.1% of total transactions
- NFR-REL-004: System shall provide offline mode capability for continued POS operations during internet outages

**Business Continuity:**
- NFR-REL-005: System shall automatically queue offline transactions when internet connectivity is lost
- NFR-REL-006: System shall synchronize offline data when connectivity resumes with conflict resolution (last-write-wins with manual override)
- NFR-REL-007: System shall provide visual indicators for synchronization status (synced, pending, failed)

### Integration

**Hardware Integration:**
- NFR-INT-001: System shall support thermal printers using ESC/POS protocol for receipt printing
- NFR-INT-002: System shall support USB HID barcode scanners for product scanning
- NFR-INT-003: System shall support Bluetooth barcode scanners for wireless product scanning
- NFR-INT-004: System shall control cash drawers via printer kick command (RJ-12 interface)

**Future Integration Readiness:**
- NFR-INT-005: System architecture shall support future API integrations with accounting software (Jurnal, Accurate, Zoho Books)
- NFR-INT-006: System architecture shall support future payment gateway integrations (QRIS, e-wallets, debit cards)

### Scalability

**Phase 1 Scope (Single Pharmacy):**
- NFR-SCAL-001: System shall support 5 concurrent cashier users across multiple POS terminals with <2 second response degradation
- NFR-SCAL-002: System shall support up to 10,000 product SKUs in inventory database
- NFR-SCAL-003: System shall support up to 50,000 transaction records per month

**Growth Path:**
- NFR-SCAL-004: System architecture shall support horizontal scaling to 5 pharmacy branches in Phase 2
- NFR-SCAL-005: Database architecture shall support growth to 100 pharmacy customers by Year 3

---

