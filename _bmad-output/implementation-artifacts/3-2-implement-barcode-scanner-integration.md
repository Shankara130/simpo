# Story 3.2: Implement Barcode Scanner Integration

**Status:** done

**Epic:** 3 - Point of Sale (Mobile)
**Priority:** Foundation (Second Story of Epic 3)
**Story Type:** Mobile Hardware Integration
**Story ID:** 3.2
**Story Key:** 3-2-implement-barcode-scanner-integration

---

## Story

**As a** Cashier,
**I want** to scan product barcodes using USB or Bluetooth scanners to quickly add items to transactions,
**So that** I can process customers faster without manual product lookup.

---

## Acceptance Criteria

1. **AC1: USB HID Barcode Scanner Input**
   - Mobile app receives barcode input from USB HID scanners as keyboard events
   - Input is debounced to prevent duplicate scans (same barcode within 500ms)
   - Barcode input is validated and trimmed before processing
   - Scan triggers automatic product lookup via backend API
   - Product is added to cart if found, or displays "product not found" error

2. **AC2: Bluetooth Barcode Scanner Support**
   - App pairs with Bluetooth scanners via device Bluetooth settings
   - Scanner input is received and processed identically to USB scanners
   - Connection status is maintained during POS operations
   - Connection failures are handled gracefully with user notification

3. **AC3: Product Lookup by SKU/Barcode**
   - Backend API is queried with scanned barcode as SKU
   - Product details (name, price, stock) are displayed in cart upon successful lookup
   - "Product not found" error is shown for unknown SKUs with option to add manually
   - API response is validated before adding product to cart

4. **AC4: User Feedback**
   - Successful scan plays confirmation sound or vibration
   - Product found displays product info (name, SKU, price, stock) in cart
   - Product not found shows clear error message: "Produk dengan barcode {barcode} tidak ditemukan"
   - Loading state is shown during API lookup

5. **AC5: Integration with Existing Components**
   - Scanner input integrates with TopControlBar search input
   - Scanned products use existing CartContext actions for adding to cart
   - Stock validation from CartContext is enforced (no out-of-stock items)
   - Scanner works seamlessly without disrupting cart management

---

## Tasks / Subtasks

- [x] **Task 1: Create Barcode Scanner Hook (AC: 1, 2, 4)**
  - [x] Create `apps/mobile/src/features/pos/hooks/useBarcodeScanner.ts`
  - [x] Implement debouncing logic (500ms cooldown between identical scans)
  - [x] Add barcode validation (trim, length check, character validation)
  - [x] Implement sound/vibration feedback on successful scan
  - [x] Add error state management for scanner failures
  - [x] Create TypeScript types for scanner state and events

- [x] **Task 2: Integrate Scanner with TopControlBar (AC: 1, 5)**
  - [x] Modify `apps/mobile/src/features/pos/components/TopControlBar.tsx`
  - [x] Replace TextInput with scanner-aware input component
  - [x] Add focus management for scanner input (auto-focus on POS mount)
  - [x] Handle scanner input without showing on-screen keyboard
  - [x] Add visual indicator for scanner-ready state

- [x] **Task 3: Implement Barcode Product Lookup (AC: 3)**
  - [x] Extend `ProductService.ts` with barcode-specific lookup method
  - [x] Add `getProductByBarcode` method (alias for `getProductBySKU`)
  - [x] Implement API response validation before cart addition
  - [x] Add "product not found" error handling with Indonesian messages
  - [x] Integrate with existing CartContext `addItem` action

- [x] **Task 4: Create Scanner Feedback Components (AC: 4)**
  - [x] Create `apps/mobile/src/features/pos/components/ScannerFeedback.tsx`
  - [x] Implement success indicator (green flash, checkmark)
  - [x] Implement error indicator (red flash, error message)
  - [x] Add loading state during product lookup
  - [x] Add haptic feedback using React Native Haptics

- [x] **Task 5: Add Scanner Settings and Configuration (AC: 2)**
  - [x] Create `apps/mobile/src/features/pos/types/scanner.types.ts`
  - [x] Define scanner configuration interface (debounce time, sound enabled, etc.)
  - [x] Add AsyncStorage key for scanner preferences
  - [x] Implement Bluetooth scanner pairing instructions UI (future story placeholder)

- [x] **Task 6: Create Tests (AC: All)**
  - [x] Create `apps/mobile/src/features/pos/hooks/useBarcodeScanner.test.ts`
  - [x] Test debouncing logic (prevent duplicate scans)
  - [x] Test barcode validation (empty, invalid characters)
  - [x] Test successful scan flow (lookup → add to cart)
  - [x] Test product not found error handling
  - [x] Test scanner failure scenarios

---

## Dev Notes

### Context & Purpose

This is the **second story of Epic 3 (Point of Sale - Mobile)**. Story 3.1 established the POS screen layout and navigation. This story adds barcode scanner integration to enable fast product scanning, a critical feature for achieving sub-30-second transaction times (NFR-PERF-001).

**Business Context:**
- Cashiers scan 50-100+ products per hour during peak times
- Manual product lookup adds 5-10 seconds per item → 30+ item transactions take 5+ minutes
- Barcode scanning reduces item entry to <1 second each → 30+ item transactions in <30 seconds
- USB HID scanners are the industry standard (plug-and-play, no drivers needed)
- Bluetooth scanners provide mobility for warehouse/stock-taking scenarios

**Technical Context:**
- USB HID barcode scanners appear as keyboard devices to Android
- Scanner input comes through TextInput `onChangeText` events
- Challenge: Distinguish scanner input from manual keyboard input
- Solution: Scanner input typically ends with Enter key and has consistent timing
- Bluetooth scanners pair via Android Bluetooth settings (standard OS pairing)

### Architecture Alignment

**[Source: _bmad-output/planning-artifacts/architecture.md#Hardware Integration]**

**Hardware Integration Requirements:**
- **NFR-INT-002:** USB HID barcode scanner support for product scanning
- **NFR-INT-003:** Bluetooth barcode scanner support for wireless scanning
- Scanner input must be validated and debounced
- <1 second response time for scan-to-product lookup (NFR-PERF-002)

**Mobile Project Structure:**
```
apps/mobile/src/features/pos/
├── hooks/
│   ├── useCart.ts                    # Already exists
│   └── useBarcodeScanner.ts          # NEW - Scanner logic
├── components/
│   ├── TopControlBar.tsx             # MODIFY - Add scanner support
│   ├── ScannerFeedback.tsx           # NEW - Visual/haptic feedback
│   └── ...                           # Existing components unchanged
├── services/
│   └── ProductService.ts             # MODIFY - Add barcode lookup
├── types/
│   └── scanner.types.ts              # NEW - Scanner configuration types
└── context/
    └── CartContext.tsx               # REFERENCE ONLY - Use existing actions
```

**API Integration Points:**
- `GET /api/v1/products?sku={barcode}` - Already exists from Story 3.1
- Product response structure: `{ id, sku, name, price, stock_qty, branch_id }`
- Use existing `ProductService.getProductBySKU()` method (already implemented)

### Previous Story Intelligence

**From Story 3.1 (Design POS Screen Layout and Navigation):**

**✅ Completed Infrastructure:**
- POS screen layout: TopControlBar (15%), ProductList (55%), CartSummary (15%), ActionButtons (15%)
- CartContext with state management: ADD_ITEM, REMOVE_ITEM, UPDATE_QUANTITY, CLEAR_CART
- ProductService with getProductBySKU, searchProducts, getProducts methods
- Navigation stack: POSNavigator with POSScreen as initial route
- TypeScript types: Product, CartItem, CartState, CartContextType

**📋 Review Findings Applied (Story 3.1):**
- Stock validation enforced in CartContext (ADD_ITEM checks stockQty)
- Cart limit of 100 items enforced
- SafeAreaView added to POSScreen for device boundaries
- Null checks added for price calculations and API responses
- Error boundaries implemented in App.tsx

**🔧 Code Patterns Established:**
```typescript
// Product API call pattern (from ProductService.ts)
const response = await axios.get(`${FULL_API_URL}/products`, {
  params: { sku: barcode.trim() },
  timeout: 10000,
});

// Cart action pattern (from CartContext.tsx)
dispatch({ 
  type: 'ADD_ITEM', 
  payload: { 
    productId, 
    sku, 
    name, 
    price, 
    quantity, 
    subtotal,
    stockQty // For validation
  } 
});
```

**📁 Files Created in Story 3.1:**
- `apps/mobile/src/features/pos/screens/POSScreen.tsx` - Main POS screen
- `apps/mobile/src/features/pos/components/TopControlBar.tsx` - Top controls (has search input)
- `apps/mobile/src/features/pos/context/CartContext.tsx` - Cart state with stock validation
- `apps/mobile/src/features/pos/services/ProductService.ts` - Product API calls
- `apps/mobile/src/features/pos/types/product.types.ts` - Product type definitions

**🚫 Deferred from Story 3.1:**
- Barcode scanning functionality explicitly deferred to Story 3.2
- Current TopControlBar has placeholder TextInput for search/barcode input
- Scanner input handling not yet implemented

### Current State Analysis

**Existing TopControlBar.tsx:**
- Has TextInput with `placeholder="Search products or scan barcode..."`
- Uses `onChangeText` prop for search query handling
- Currently only handles manual text input (no scanner logic)
- Missing: debouncing, scanner detection, automatic lookup, feedback

**What Needs to Change:**
1. Add scanner input detection (distinguish from manual typing)
2. Implement debouncing (prevent duplicate scans within 500ms)
3. Trigger automatic product lookup on scan complete
4. Add visual/haptic feedback for scan success/failure
5. Integrate with existing CartContext for adding scanned products

### Technical Requirements

**Barcode Scanner Detection:**

USB HID scanners send input as keyboard events with these characteristics:
- Fast typing speed (5-20ms between characters)
- Consistent timing (human typing varies more)
- Ends with Enter key (`\n`)
- No focus changes during input

**Detection Strategy:**
```typescript
// Track input timing to detect scanner vs keyboard
interface ScannerInputState {
  inputBuffer: string;
  firstCharTime: number | null;
  lastCharTime: number | null;
  isScanning: boolean;
}

// If total input time < 100ms and ends with Enter → likely scanner
// If total input time > 300ms or manual edits → likely keyboard
```

**Debouncing Logic:**
```typescript
// Prevent duplicate scans of same barcode
const DEBOUNCE_MS = 500;
let lastBarcodeTime = 0;
let lastBarcode = '';

function canProcessBarcode(barcode: string): boolean {
  const now = Date.now();
  if (barcode === lastBarcode && (now - lastBarcodeTime) < DEBOUNCE_MS) {
    return false; // Skip duplicate scan
  }
  lastBarcode = barcode;
  lastBarcodeTime = now;
  return true;
}
```

**React Native Haptics for Feedback:**
```typescript
import { Platform } from 'react-native';
import * as Haptics from 'expo-haptics'; // If available, or React Native built-in

// Success feedback
Haptics.notificationAsync(
  Haptics.NotificationFeedbackType.Success
);

// Error feedback
Haptics.notificationAsync(
  Haptics.NotificationFeedbackType.Error
);
```

**Dependencies Check:**
Current `package.json` does NOT include `expo-haptics`. Options:
1. Use React Native built-in `Vibration` API (already available)
2. Add `expo-haptics` as dependency (better experience, requires install)

**Recommended:** Use built-in `Vibration` API for MVP to avoid new dependencies.

### Project Structure Notes

**Files to CREATE in this story:**

1. `apps/mobile/src/features/pos/hooks/useBarcodeScanner.ts` - Scanner logic hook
2. `apps/mobile/src/features/pos/components/ScannerFeedback.tsx` - Visual feedback
3. `apps/mobile/src/features/pos/types/scanner.types.ts` - Scanner configuration types
4. `apps/mobile/src/features/pos/hooks/useBarcodeScanner.test.ts` - Tests

**Files to MODIFY in this story:**

1. `apps/mobile/src/features/pos/components/TopControlBar.tsx` - Integrate scanner
2. `apps/mobile/src/features/pos/services/ProductService.ts` - Add barcode lookup alias

**Files to REFERENCE (do NOT modify):**

- `apps/mobile/src/features/pos/context/CartContext.tsx` - Use existing ADD_ITEM action
- `apps/mobile/src/features/pos/types/product.types.ts` - Use existing Product type
- `apps/mobile/src/features/pos/screens/POSScreen.tsx` - Scanner integrated via TopControlBar

**Naming Conventions (from Story 3.1):**
- Hooks: camelCase with "use" prefix (e.g., `useBarcodeScanner`)
- Components: PascalCase (e.g., `ScannerFeedback`)
- Types: PascalCase (e.g., `ScannerConfig`)
- Services: PascalCase with "Service" suffix
- Test files: Same name with `.test.ts` suffix

### Testing Requirements

**Test Framework (from Story 3.1):**
- Jest + React Native Testing Library
- Test files co-located with source
- Mock `ProductService` for API calls
- Mock `Vibration` API for feedback
- Test coverage goal: >80% for scanner hook

**Test Scenarios:**

1. **Scanner Detection:**
   - Fast input with Enter → detected as scanner
   - Slow input → treated as manual typing
   - Input without Enter → treated as manual typing

2. **Debouncing:**
   - Same barcode scanned twice within 500ms → second scan ignored
   - Different barcodes scanned rapidly → both processed
   - Same barcode after 500ms → both processed

3. **Product Lookup:**
   - Valid barcode → product added to cart
   - Invalid barcode → error message shown
   - API error → error message shown
   - Network timeout → error message shown

4. **Feedback:**
   - Success → vibration called, success indicator shown
   - Error → vibration called, error indicator shown
   - Loading → loading indicator shown during lookup

### Implementation Gotchas

**⚠️ CRITICAL: Scanner vs Keyboard Input**

Don't break manual search functionality! Cashiers need both:
- Scanner for fast product lookup
- Manual typing for search by product name

Solution: Detect scanner input automatically, don't require separate mode.

**⚠️ CRITICAL: CartContext Stock Validation**

CartContext ADD_ITEM already enforces stock validation:
```typescript
// From CartContext.tsx line 83-88
const productStock = action.payload.stockQty ?? 0;
if (productStock <= 0) {
  console.warn('Cannot add: Product is out of stock');
  return state;
}
```

**Must pass stockQty when calling addItem from scanner:**
```typescript
// WRONG - will skip stock validation
dispatch({ 
  type: 'ADD_ITEM', 
  payload: { productId, sku, name, price, quantity, subtotal } 
});

// CORRECT - includes stock validation
dispatch({ 
  type: 'ADD_ITEM', 
  payload: { productId, sku, name, price, quantity, subtotal, stockQty } 
});
```

**⚠️ CRITICAL: API Response Validation**

ProductService already validates response structure, but scanner hook must handle errors:
```typescript
// Product service throws ProductServiceError
try {
  const product = await ProductService.getProductBySKU(barcode);
  // Add to cart
} catch (error) {
  if (error instanceof ProductServiceError) {
    if (error.statusCode === 404) {
      // Product not found - show error
    } else {
      // API error - show generic error
    }
  }
}
```

**⚠️ TextInput Focus Management**

Scanner input requires focused TextInput:
- Auto-focus TopControlBar input when POS screen mounts
- Keep focus during scanning (don't lose focus to other elements)
- Allow manual focus loss for manual product selection
- Re-focus after manual interaction completes

### Performance Requirements

**[Source: _bmad-output/planning-artifacts/prd.md#Non-Functional Requirements]**

- **NFR-PERF-002:** System shall process product barcode scans and return product information within 1 second
- **NFR-PERF-001:** Complete end-to-end sales transactions within 30 seconds (scanner is critical for this)

**Scanner Performance Targets:**
- Scan-to-lookup: <500ms (scanner input → API request → product found)
- Add-to-cart: <100ms (product found → cart update)
- Total scan-to-cart: <1 second
- Debouncing overhead: <10ms (negligible)

### Security Considerations

**Input Validation:**
- Barcode input must be trimmed before API call
- Barcode length validation (typically 8-13 characters)
- Character validation (alphanumeric, no special characters except hyphen)
- SQL injection protection (use parameterized queries in backend)

**Data Privacy:**
- No sensitive data in barcode input
- No logging of full barcode values (use partial masking in logs)

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic 3] - Epic 3 requirements and AC
- [Source: _bmad-output/planning-artifacts/architecture.md#Hardware Integration] - NFR-INT-002, NFR-INT-003
- [Source: _bmad-output/planning-artifacts/prd.md#Non-Functional Requirements] - NFR-PERF-001, NFR-PERF-002
- [Source: _bmad-output/implementation-artifacts/3-1-design-pos-screen-layout-and-navigation.md] - Previous story implementation

---

## Dev Agent Record

### Agent Model Used

Claude 4.6 Opus (bmad-create-story workflow)

### Completion Notes List

- Story context created with exhaustive analysis of all artifacts
- Previous story intelligence extracted from Story 3.1
- Architecture patterns documented for consistency
- Anti-pattern prevention: Reinvention of existing CartContext, ProductService
- Regression prevention: Stock validation, API error handling patterns documented
- Testing requirements aligned with Story 3.1 patterns
- Critical gotchas documented to prevent common mistakes

**Task 4 Completed - ScannerFeedback Component:**
- Created ScannerFeedback.tsx with visual feedback for scanner states
- Implemented state-based color coding (green=success, red=error, blue=scanning, orange=loading)
- Added accessibility support (aria-label, live region announcements)
- Used React Native's built-in Vibration API for haptic feedback (no new dependencies)
- All 21 tests passing for ScannerFeedback component
- Component exported from index.ts for public API

**Task 5 Completed - Scanner Settings and Configuration:**
- Created ScannerConfigService.ts with AsyncStorage persistence
- Implemented load/save/reset methods for scanner configuration
- Config merges with defaults when loading partial/invalid data
- All 12 tests passing for ScannerConfigService
- Bluetooth pairing UI deferred to future story per requirements

**Task 6 Completed - Comprehensive Testing:**
- Fixed timing issues in useBarcodeScanner tests (all 19 tests passing)
- Properly configured Jest fake timers for async operations
- Fixed test timing to stay under 100ms maxScanTimeMs threshold
- All POS tests passing: 98 tests total (19 scanner + 21 feedback + 12 config + others)
- Full test suite validates scanner detection, debouncing, validation, error handling

### File List

**To Create:**
- apps/mobile/src/features/pos/hooks/useBarcodeScanner.ts ✓
- apps/mobile/src/features/pos/hooks/useBarcodeScanner.test.ts ✓
- apps/mobile/src/features/pos/components/ScannerFeedback.tsx ✓
- apps/mobile/src/features/pos/components/ScannerFeedback.test.tsx ✓
- apps/mobile/src/features/pos/types/scanner.types.ts ✓
- apps/mobile/src/features/pos/services/ScannerConfigService.ts ✓
- apps/mobile/src/features/pos/services/ScannerConfigService.test.ts ✓

**To Modify:**
- apps/mobile/src/features/pos/components/TopControlBar.tsx (add scanner support) ✓
- apps/mobile/src/features/pos/services/ProductService.ts (add barcode lookup alias) ✓

**To Reference (Read Only):**
- apps/mobile/src/features/pos/context/CartContext.tsx (use existing actions)
- apps/mobile/src/features/pos/types/product.types.ts (use existing types)
- apps/mobile/src/features/pos/screens/POSScreen.tsx (integration point)
