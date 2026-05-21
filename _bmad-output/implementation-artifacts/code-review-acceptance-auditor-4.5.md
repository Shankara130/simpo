# Acceptance Auditor Review - Story 4.5: Implement Expiry Date Alerts

## Your Role

You are an **Acceptance Auditor**. Your job is to verify that the implementation matches the specification and acceptance criteria. You receive the diff, the full spec file, and access to context docs.

## Diff to Review

[Use the same diff from the Blind Hunter prompt - it's in the file: code-review-blind-hunter-4.5.md]

## Specification to Verify

### User Story
```
As a **Pharmacy Owner**,
I want **to receive advance alerts when products are approaching their expiry dates at 30, 14, and 7 days**,
so that **I can discount or dispose of expiring medications proactively and comply with regulations**.
```

### Acceptance Criteria

**AC1:** Given products have expiry dates recorded in the system, When the current date reaches 30 days before a product's expiry date, Then the system generates the first 30-day expiry alert

**AC2:** Given a product is approaching expiry, When the current date reaches 14 days before expiry, Then the system generates a 14-day alert

**AC3:** Given a product is near expiry, When the current date reaches 7 days before expiry, Then the system generates a 7-day alert

**AC4:** Given an expiry alert is generated, When the event is published, Then it is published to Redis pub/sub with event type "product.expiry"

**AC5:** Given a product.expiry event is published, When subscribed clients receive the event, Then notifications are displayed to owners via mobile app alert banner and web dashboard notifications

**AC6:** Given an expiry alert is generated, When the notification payload is constructed, Then it includes: product SKU, product name, expiry date, days remaining, and branch location

**AC7:** Given a 7-day expiry alert is generated, When the notification is displayed, Then it is marked as urgent with visual highlighting (red background, bold text)

### Key Task Requirements

**Backend Task 1:** Create ExpiryAlertEvent struct with fields:
- EventID: string (UUID)
- EventType: string (constant: "product.expiry")
- Timestamp: time.Time
- Data: ProductExpiryData with ProductID, SKU, ProductName, ExpiryDate, DaysRemaining (30, 14, or 7), AlertLevel ("warning", "critical", "urgent"), BranchID, BranchName

**Backend Task 2:** Implement ExpiryCheckService with:
- CheckExpiringProducts method
- Query products where expiry_date BETWEEN (NOW + 7 days) AND (NOW + 30 days)
- Categorize by alert level: 30 days="warning", 14 days="critical", 7 days="urgent"
- Debounce logic using Redis Sorted Set with key format: `expiry_alerts:{product_id}:{branch_id}`
- Only alert if 24+ hours since last alert for same threshold

**Backend Task 3:** AlertService.PublishExpiryAlert:
- Publish to Redis pub/sub channel: "product.expiry"
- Graceful error handling (log but don't fail)

**Backend Task 4:** Scheduled job running every 6 hours (00:00, 06:00, 12:00, 18:00)

**Backend Task 5:** GET /api/v1/products/expiring?days={30,14,7}&branch_id={id}
- RBAC: Owners see all branches, Cashiers see assigned branch only

**Web Task 7:** ExpiryAlertBanner component with:
- Color coding: 30-day yellow/orange, 14-day orange, 7-day red with bold text
- Product info: SKU, name, expiry date, days remaining, branch
- "View Product" and "Dismiss" buttons
- Auto-dismiss after 60 seconds

**Web Task 8:** Expiring products page with:
- Filters by days (30, 14, 7) and branch
- Table with: Product, SKU, Expiry Date, Days Remaining, Branch, Actions
- Sort by urgency (closest expiry first)

**Web Task 9:** Navigation item "Expiring" with:
- Badge showing 7-day urgent count
- Highlight when critical alerts exist (red background, pulse animation)

**Mobile Task 11:** ExpiryAlertBanner with color coding and swipe-to-dismiss

**Mobile Task 12:** ExpiringProductsScreen with pull-to-refresh and filter buttons

### Architectural Requirements

**Event Naming:** `{domain}.{action}` format → "product.expiry"

**Event Payload Structure:**
```go
type ExpiryAlertEvent struct {
    EventID   string            `json:"eventId"`
    EventType string            `json:"eventType"` // "product.expiry"
    Timestamp string            `json:"timestamp"`
    Data      ProductExpiryData `json:"data"`
}
```

**Role-Based Access Control:**
- Owners and Admins receive expiry alerts
- Cashiers do NOT receive expiry alerts (similar to low stock)

**Performance:** NFR-PERF-006: UI response < 500ms

**Scheduled Job:** Runs every 6 hours using time.Ticker with context cancellation

### Context from Previous Stories

**Story 4.4 (Low Stock) established patterns:**
- AlertService with Redis pub/sub
- Debounce using Redis Sets/Sorted Sets with 24-hour TTL
- WebSocket extension for multiple event types
- Color-coded alert banners

**Story 4.2 (Real-Time Stock) established:**
- WebSocket handler in product_handler.go
- useStockWebSocket hook pattern
- RealTimeStockService for mobile

## Your Mission

Audit the implementation against the spec:

1. **AC Compliance Check**: For each AC1-AC7, verify the implementation satisfies it completely
2. **Task Completion Check**: Verify each task's subtasks are implemented as specified
3. **Architecture Compliance**: Check event naming, payload structure, RBAC rules
4. **Spec Intent vs Implementation**: Identify any deviations from the spec's intent
5. **Missing Implementation**: Find specified behaviors not implemented

## Output Format

Provide your findings as a Markdown list. Each non-compliant finding must have:

```markdown
### [Severity] AC Violation: [AC number] - [Title]

**Requirement:** [Exact text from the AC or task]

**Evidence:** [What the code actually does - reference diff or file path]

**Issue:** [Why this doesn't match the requirement]

**Expected:** [What the spec requires]

**Suggested Fix:** [How to make it comply]
```

For task completion issues:
```markdown
### [Severity] Missing Implementation: [Task X.Y] - [Title]

**Required:** [What the task specifies]

**Evidence:** [What exists (or doesn't exist)]

**Suggested Fix:** [What needs to be added]
```

Severity levels: `Critical`, `High`, `Medium`, `Low`, `Info`

**Important:** Be thorough. A missed acceptance criterion is a failed story. If the spec says "7-day alerts must be bold" and the code uses normal font weight, that's a violation.
