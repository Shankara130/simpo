# Deferred Work Items

This file tracks issues that were identified during reviews but deferred to future stories.

## Deferred from: code review of 1-3-initialize-web-admin-dashboard-with-next-js (2026-05-10)

- **Missing Token Expiration Validation** [lib/apiClient.ts:115-119]
  - Reason: JWT decode logic needed - requires crypto library implementation
  - Impact: User might remain "authenticated" after token expires
  - Future Story: 1.5 (Implement User Authentication with JWT)

- **Missing CSRF Protection** [lib/apiClient.ts:39-59]
  - Reason: Backend CSRF implementation needed first
  - Impact: Application vulnerable to CSRF attacks with cookie-based auth
  - Future Story: 1.5 (Implement User Authentication with JWT)

- **No Error Boundaries** [app/layout.tsx]
  - Reason: Error boundary implementation requires separate task
  - Impact: Unhandled errors crash entire app with blank screen
  - Future Story: 1.6 (Configure Development Infrastructure)

- **Inconsistent API Client Patterns** [lib/auth.ts vs lib/apiClient.ts]
  - Reason: Acceptable for foundation story - will consolidate when implementing auth
  - Impact: Inconsistent error handling patterns
  - Future Story: 1.5 (Implement User Authentication with JWT)

- **Missing React.memo on Layout** [app/(auth)/layout.tsx:4-34]
  - Reason: Performance optimization, not critical for MVP
  - Impact: Minor performance impact from unnecessary DOM reconciliation
  - Future Story: Performance optimization story
