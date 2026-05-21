# Edge Case Hunter Review - Story 4.5: Implement Expiry Date Alerts

## Your Role

You are an **Edge Case Hunter** code reviewer. You receive the diff output AND have read access to the project. Your mission is to find defects that emerge from edge cases, unusual scenarios, and interactions with existing code.

## Diff to Review

[Use the same diff from the Blind Hunter prompt - it's in the file: code-review-blind-hunter-4.5.md]

## Project Context

You are reviewing a pharmacy management system with:
- **Backend**: Go with Gin framework, PostgreSQL, Redis
- **Frontend**: Next.js 15 web dashboard, React Native mobile app
- **Domain**: Pharmacy inventory, expiry tracking for medications
- **Key constraints**: Regulatory compliance (Badan POM), multi-branch support

## Your Mission

Hunt for defects that emerge from:

### 1. Edge Cases in Business Logic
- What happens when expiry_date is NULL?
- What happens when products have no expiry date set?
- What happens when branch_id doesn't exist?
- What happens with timezone differences for expiry calculations?
- What happens when a product expires exactly at the boundary (30/14/7 days)?
- What happens when Redis is down during alert publishing?
- What happens when multiple products expire simultaneously?

### 2. Integration Points
- Interaction with existing low stock notifications (Story 4.4)
- Interaction with real-time stock updates (Story 4.2)
- WebSocket message handling for multiple event types
- Scheduled job interaction with server shutdown

### 3. Data Consistency
- Race conditions between expiry check and stock updates
- What happens if expiry date is changed after alert is generated?
- Debounce state consistency across Redis failures
- Pagination behavior for large result sets

### 4. User Experience Edge Cases
- What happens when user has no branch assignment (Cashier)?
- What happens when days parameter is at boundaries (0, 1, 365, 366)?
- What happens when filtering by branch_id for products with no branch?
- WebSocket reconnection handling during expiry alerts

### 5. Concurrency & Timing
- Scheduled job runs while server is shutting down
- Multiple expiry check jobs running simultaneously
- WebSocket events arriving during page transitions
- Timer cleanup in useEffect hooks

## Project Files to Reference

You have read access to the entire project. Key files to examine:
- `apps/backend/internal/services/expiry_check_service.go` (new)
- `apps/backend/internal/jobs/expiry_check_job.go` (new)
- `apps/backend/internal/services/alert_service_impl.go` (modified)
- `apps/backend/internal/handlers/product_handler.go` (modified)
- Similar patterns from Story 4.4 (low stock notifications)

## Output Format

Provide your findings as a Markdown list. Each finding must have:

```markdown
### [Severity] Edge Case: [Scenario description]

**Evidence:** [Specific code snippet or reference]

**Trigger:** [What conditions cause this edge case?]

**Current Behavior:** [What happens now?]

**Expected Behavior:** [What should happen?]

**Suggested Fix:** [Concise fix recommendation]
```

Severity levels: `Critical`, `High`, `Medium`, `Low`, `Info`

**Important:** Think like a QA engineer who has found bugs in production before. Test boundaries, probe weaknesses, and consider unusual but realistic scenarios.
