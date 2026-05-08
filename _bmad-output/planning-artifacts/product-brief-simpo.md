---
title: "Product Brief: simpo"
status: "complete"
created: "2026-05-08T22:00:00+07:00"
updated: "2026-05-08T22:30:00+07:00"
inputs: ["User interview", "Web research (Farmacare, SwipeRx, competitors)", "Technical analysis (Golang architecture)"]
project: "simpo"
type: "pharmacy-management-system"
---

# Product Brief: simpo

## Executive Summary

**simpo** is a cost-effective pharmacy management system designed for small to medium-sized pharmacies operating 2-5 branches. It addresses a critical pain point in the Indonesian market: existing solutions like Farmacare provide complete functionality but at subscription costs that strain thin pharmacy margins.

By building a modern, feature-complete alternative using Golang (backend) and React Native (frontend), simpo delivers the essential operational capabilities that pharmacies use daily—point-of-sale, inventory management, financial reporting, and purchase invoicing—at a fraction of the cost. The immediate use case is internal deployment at the owner's pharmacy, with a clear path to commercialize as an affordable SaaS offering once validated.

The timing is right. Indonesian pharmacies are increasingly digitalizing operations, but current solutions either overserve (enterprise complexity) or underperform (basic POS tools). There's a gap for a pragmatic, mid-market solution that balances capability with affordability.

## The Problem

**Who feels it:** Pharmacy owners with 2-5 branches who've adopted digital systems but are bleeding from monthly subscription fees.

**What they're coping with today:**
- Paying IDR 3-10+ million per month for pharmacy management software (Farmacare and competitors)
- Systems that work functionally but create ongoing cash flow pressure
- The paradox: digital tools increase operational efficiency, but subscription costs erode the margins they're meant to protect

**The cost of status quo:**
- For a 3-branch pharmacy paying IDR 5M/month: IDR 60M annually in software fees
- That's 1-2 staff salaries, or inventory that could generate revenue
- Owners feel trapped—switching means migration pain and learning curve, staying means perpetual cost

**The insight:** Pharmacy owners don't want fewer features. They want the *same* features at a price that respects their margins. A self-hosted or affordably-priced SaaS alternative, built with modern tech, can deliver 80% of the functionality at 20% of the cost.

## The Solution

**simpo** is a web + mobile pharmacy management system that covers the daily operational workflow:

**Core Capabilities:**
- **Point-of-Sale (Kasir):** Fast, reliable transaction processing with receipt printing
- **Inventory Management (Cek Stok):** Real-time stock visibility across branches
- **Financial Reporting (Laporan Keuangan):** Daily, weekly, monthly profit/loss and sales reports
- **Purchase Invoicing (Faktur Pembelian):** Supplier management and purchase order tracking

**Security & Access:**
- Multi-user roles: System Admin, Owner, Cashier—each with appropriate permissions
- Staff registration via admin approval or whitelist
- Secure login authentication

**Smart Alerts:**
- Low stock notifications (reorder triggers)
- Expiry date alerts (reduce shrinkage from expired medications)

**Reporting & Operations:**
- iReport: Custom, specific-format reports for regulatory and business needs
- Print capabilities for receipts, invoices, and reports

The experience is straightforward: staff log in, process sales, check stock, print receipts. Owners review reports, receive alerts, manage inventory. Admins oversee users and system configuration. No complexity they won't use.

## What Makes This Different

| Competitor Approach | simpo Approach |
|---------------------|----------------|
| Subscription-heavy pricing (IDR 3-10M+/month) | Self-hosted or affordable SaaS (target: <IDR 1M/month when commercialized) |
| Legacy tech stacks, slower iteration | Modern Golang backend + React Native frontend—fast, scalable, maintainable |
| One-size-fits-all feature bloat | Pragmatic feature set focused on daily-used operations |
| Enterprise-focused or overly basic | Purpose-built for 2-5 branch mid-market |

**The unfair advantage:**
- **Real-world validation:** Built for and tested on an actual pharmacy before commercializing
- **Tech stack efficiency:** Golang's performance enables lower infrastructure costs than legacy stacks
- **No investor pressure:** Bootstrapped approach means pricing can remain affordable without growth-at-all-costs pressure

**Honest assessment:** The moat isn't unique technology—pharmacy management is a well-understood domain. The advantage is execution speed, lower cost structure, and genuine customer proximity. The first 2-3 years will be about winning customers through reliability, responsiveness, and price.

## Who This Serves

**Primary User: Pharmacy Owners (2-5 branches)**

They care about:
- Margins—every million rupiah saved goes to inventory or staff
- Control—knowing stock, sales, and cash flow across branches
- Compliance—reports that satisfy tax and regulatory requirements
- Reliability—systems that don't break during peak hours

Success looks like: Spending less on software while maintaining operational excellence. Having stock visibility that prevents stockouts. Getting alerts before meds expire. Printing reports for the accountant without hassle.

**Secondary Users:**

*Cashiers* need fast, reliable POS. They shouldn't think about the system—it just works.

*System Admins* need to manage users, configure branches, and oversee system health. They want clear dashboards and straightforward controls.

## Success Criteria

**For the internal deployment (first 6 months):**
- Successfully replaces Farmacare without operational disruption
- All 10 core features functional and used daily
- Staff adoption rate >90% (cashiers actually using it)
- Monthly subscription savings: 100% (zero external fees)

**For future SaaS commercialization:**
- 10 paying pharmacies within 12 months of launch
- Churn rate <5% monthly
- Average revenue per pharmacy: IDR 500K-1M/month
- Customer acquisition cost <3 months of revenue
- Net Promoter Score >40

**Leading indicators:**
- Time to process a sale: <30 seconds (parity with Farmacare)
- Stock reconciliation accuracy: >99%
- Uptime: >99.5%
- Support response time: <4 hours during business hours

## Scope

**IN for MVP:**

| Feature | Priority |
|---------|----------|
| Login & Authentication | Critical |
| Multi-user roles (Admin, Owner, Cashier) | Critical |
| POS/Kasir with receipt printing | Critical |
| Stock management with real-time updates | Critical |
| Financial reports (P&L, sales summaries) | Critical |
| Purchase invoicing | Critical |
| Low stock alerts | High |
| Expiry date alerts | High |
| Staff registration (admin/whitelist) | High |
| iReport (custom reports) | Medium |
| Print (receipts, invoices, reports) | High |

**OUT for MVP (future phases):**

- E-prescription integration (BPJS, government systems)
- Supplier marketplace/direct ordering
- Mobile customer app (loyalty, refill reminders)
- Multi-branch inventory transfers
- Advanced analytics and forecasting
- API integrations (accounting software, e-commerce)

**Technical scope:**
- Golang backend (monolith initially, microservices-ready)
- React Native mobile app (Android first, iOS if feasible)
- Web admin dashboard (React/Next.js)
- PostgreSQL database
- Self-hosted deployment (Docker)

## Vision

**Year 1:** Internal deployment. Prove the system works in a real pharmacy. Fix bugs. Refine workflows. Document everything. Build quietly.

**Year 2:** Soft launch to 5-10 friendly pharmacies. Word-of-mouth marketing. Pricing at IDR 500K/month to validate willingness-to-pay. Incorporate feedback rapidly. Reliability and responsiveness are the brand.

**Year 3:** Broader launch across Indonesia. Target 100 pharmacies. IDR 750K-1M/month pricing. Add requested features (multi-branch transfers, maybe e-prescription integration). Small team, high margins.

**If successful:** simpo becomes the default choice for independent Indonesian pharmacies who want capability without enterprise pricing. A beloved, regional product that grows through customer love, not VC sales teams. The goal isn't unicorn status—it's a profitable, sustainable business that serves an underserved market well.

---

*Prepared by: BMad Product Brief Workflow*
*Date: 2026-05-08*
