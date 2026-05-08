---
title: "Product Brief Distillate: simpo"
type: llm-distillate
source: "product-brief-simpo.md"
created: "2026-05-08T22:30:00+07:00"
purpose: "Token-efficient context for downstream PRD creation"
project: "simpo"
stage: "product-brief-complete"
---

# Product Brief Distillate: simpo

This distillate captures all overflow context from product brief discovery—requirements hints, technical decisions, competitive intelligence, and open questions. Use this as input for PRD creation to avoid re-work and re-proposals.

---

## User Requirements (Explicit)

**Core Features (MVP):**
1. Login & Authentication system
2. Multi-user roles: System Admin, Owner, Cashier
3. POS/Kasir with receipt printing
4. Stock management (cek stok) with real-time updates
5. Financial reports (laporan keuangan)
6. Purchase invoicing (faktur pembelian)
7. Low stock notifications (stok menipis)
8. Expiry date alerts (obat kadaluarsa)
9. Staff registration (by admin or whitelist)
10. iReport (specific format reports)
11. Print capabilities (receipts, invoices, reports)

**User Roles:**
- System Admin: Full access, user management, system configuration
- Owner: Business oversight, reports, financial data access
- Cashier: POS operations, sales processing

**Target:**
- Apotek dengan 2-5 cabang
- Saat ini untuk apotek sendiri (internal use)
- Masa depan: SaaS affordable

**Motivation:**
- Saat ini menggunakan Farmacare
- Ingin ganti untuk mengurangi pengeluaran (cost reduction)
- Target free sekarang → SaaS di masa depan

---

## Technical Context

**Backend:**
- Golang (explicitly requested by user)
- Microservices-ready architecture
- Monolith initially for MVP

**Frontend:**
- React Native (already bootstrapped in project)
- Android first, iOS if feasible
- Web admin dashboard needed

**Infrastructure:**
- Self-hosted deployment (Docker)
- PostgreSQL database
- Redis for caching and real-time updates
- gRPC/REST for inter-service communication

**Architecture Patterns (from research):**
- Microservices: Inventory, POS, Procurement, Reporting, Auth, Expiration Tracking
- Offline mode consideration for unreliable internet areas

**Hardware Integration:**
- Thermal printers (receipt printing)
- Barcode scanners
- Cash drawers
- A4 printers (reports, invoices)

---

## Competitive Intelligence

**Direct Competitors:**

| Name | Type | Pricing | Strengths | Weaknesses |
|------|------|---------|-----------|------------|
| Farmacare | Indonesia-specific | IDR 3-10M/month | Established, local compliance, market familiarity | Expensive, tech stack unknown |
| SwipeRx | Regional (SEA) | Unknown | Largest pharmacy community platform, strong regional presence | May not be Indonesia-specific, potentially complex for small pharmacies |
| PharmaPOS | Indonesia | Unknown | Specifically designed for Indonesian pharmacies | Market penetration unknown |
| GPOS | Multi-sector | Unknown | Versatile (serves pharmacies, clinics, hospitals) | Not pharmacy-specialized |
| G-MEDS | Indonesia | Unknown | Pharmacy management, healthcare integration | Less market presence |
| Trustmedis | Indonesia | Unknown | Medical software expertise, digital solutions | Not pharmacy-specialized |

**Major Chains (Indirect Competitors):**
- Kimia Farma, Apotek K-24, Kalbe Farma
- Use in-house solutions, not available to independents
- Have e-commerce platforms and mobile apps

**Market Gap Identified:**
- Small independent pharmacies (1-3 locations) underserved
- Limited mobile-first solutions
- Few options with good offline support
- Complex/expensive enterprise solutions dominate

---

## Pain Points (From Research & User)

**Operational:**
- Inefficient manual inventory tracking
- Stock management difficulties
- Expiration date tracking complexity
- Manual cashier processes
- Lack of real-time stock visibility

**Financial:**
- Cash flow management
- Profit margin optimization
- Financial reporting complexity
- Cost control difficulties
- High software subscription costs (IDR 3-10M/month)

**Compliance:**
- Badan POM regulatory compliance
- Prescription management requirements
- Audit trail maintenance

**Scalability:**
- Multi-location management
- Business growth support
- Integration with suppliers

**Switching Barriers:**
- Migration pain and learning curve
- Data loss fears
- Business disruption during transition

---

## Differentiation Opportunities

**Validated:**
- Cost-effective alternative to Farmacare
- Modern tech stack (Golang + React Native)
- Real-world validation through internal deployment
- No investor pressure = sustainable pricing
- Bootstrapped = capital efficient

**Identified but Not Committed:**
- Regulatory compliance automation (BPJS, POM)
- Supplier ecosystem play (PBF ordering channel)
- Staff training platform
- Migration insurance proposition
- Data sovereignty as premium feature
- Localization as defense (Indonesian workflows)
- Reliability as premium positioning

**Partnership Opportunities:**
- Accounting software (Jurnal, Accurate, Zoho Books)
- Hardware OEM distributors
- Pharmacy associations (GPFI, regional groups)
- Pharma distributors (PBFs)

---

## Scope Signals

**IN for MVP (User Confirmed):**
- All 11 core features listed above
- Multi-branch support (user has 2-5 branches)
- Self-hosted deployment

**OUT for MVP (Explicitly Deferred):**
- E-prescription integration (BPJS, government systems)
- Supplier marketplace/direct ordering
- Mobile customer app (loyalty, refill reminders)
- Multi-branch inventory transfers
- Advanced analytics and forecasting
- API integrations (accounting software, e-commerce)

**Open Questions (From Review):**
- MVP scope: Fased (POS first) or all features together?
- Multi-branch: v1 (single branch) or v1 with multi-branch?
- Deployment: Self-hosted only or managed hosting option?
- Regulatory compliance: How familiar is user with BPJS/POM requirements?

---

## Risks & Considerations

**Technical Risks:**
- Data migration from Farmacare (high severity)
- Downtime during switching (high severity)
- Hardware compatibility (medium severity)
- Multi-branch sync complexity (medium severity)
- Offline mode architecture for unreliable internet

**Business Risks:**
- Regulatory non-compliance could make system unusable
- Support burden for self-hosted deployments
- Feature creep during internal deployment
- Staff adoption below 90% target
- Competitors could match pricing

**Unvalidated Assumptions:**
- Indonesian 2-5 branch pharmacies find Farmacare expensive but functional
- 80% of functionality can be delivered at 20% of cost
- Golang enables significantly lower infrastructure costs
- Self-hosted deployment is viable for non-technical pharmacy owners
- Staff adoption rate >90% is achievable

---

## Success Criteria (From Brief + Context)

**Internal Deployment (First 6 Months):**
- Successfully replaces Farmacare without operational disruption
- All 10 core features functional and used daily
- Staff adoption rate >90%
- Monthly subscription savings: 100%

**SaaS Commercialization (Future):**
- 10 paying pharmacies within 12 months
- Churn rate <5% monthly
- ARPU: IDR 500K-1M/month
- CAC <3 months of revenue
- NPS >40

**Leading Indicators:**
- Transaction time: <30 seconds (parity with Farmacare)
- Stock reconciliation accuracy: >99%
- Uptime: >99.5%
- Support response: <4 hours during business hours

---

## Vision & Roadmap

**Year 1:** Internal deployment only. Prove system works. Fix bugs. Refine workflows.

**Year 2:** Soft launch to 5-10 friendly pharmacies. Word-of-mouth. IDR 500K/month pricing.

**Year 3:** Broader launch. Target 100 pharmacies. IDR 750K-1M/month pricing.

**Long-term:** Default choice for independent Indonesian pharmacies. Profitable, sustainable business. Not unicorn status—regional beloved product.

---

## Regional Context

**Country:** Indonesia
**Language:** Bahasa Indonesia (for UI and documentation)
**Currency:** IDR
**Regulatory:** Badan POM (Food and Drug Authority)
**Healthcare System:** BPJS (national health insurance)
**Market Maturity:** Early adoption stage for digital pharmacy solutions

**Localization Needs:**
- Date formats (DD/MM/YYYY)
- Number formats (Indonesian conventions)
- Tax rules and reporting
- Report formats for local compliance
- Islamic finance compliance for payment terms (considered)

---

## Rejected or Deferred Ideas

**Rejected:**
- None explicitly rejected by user

**Deferred (User Aware):**
- E-prescription integration (future phase)
- Supplier marketplace (future phase)
- Mobile customer app (future phase)
- Multi-branch transfers (future phase)
- Advanced analytics (future phase)
- Accounting software integration (future phase)

**Reviewer Suggestions Not Yet Committed:**
- Fased MVP approach
- Single-branch first, multi-branch later
- Managed hosting option
- Regulatory compliance as differentiator
- Zero-downtime migration guarantee

---

## Open Questions for PRD Phase

1. What are the specific acceptance criteria for each of the 11 core features?
2. What defines "functional" for the internal deployment success criteria?
3. What are the specific BPJS/POM compliance requirements?
4. What hardware devices will be supported (specific printer models, barcode scanners)?
5. What is the data migration strategy from Farmacare?
6. What is the training plan for staff to achieve >90% adoption?
7. What are the specific report formats required (iReport specifications)?
8. What is the rollback plan if internal deployment fails?
9. What are the offline mode requirements and sync strategy?
10. What is the support model for self-hosted deployments?

---

## Next Steps

**Immediate:** Use this brief and distillate as input for PRD creation.

**When creating PRD:**
- Reference the 11 core features as must-have requirements
- Consider the open questions above as research topics
- Validate assumptions through user interviews (current Farmacare users)
- Define acceptance criteria for each feature
- Specify data migration strategy
- Detail hardware compatibility requirements
- Define compliance requirements (BPJS/POM)

**Before Development:**
- Resolve open questions about MVP scope (fased vs. all features)
- Clarify multi-branch timing (v1 vs. v2)
- Decide on deployment model options
- Validate regulatory compliance requirements
- Create detailed technical architecture

---

*Generated by: BMad Product Brief Workflow*
*Date: 2026-05-08*
*Token-optimized for downstream PRD consumption*
