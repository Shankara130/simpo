# Simpo Mobile POS

Mobile Point of Sale application for simpo Pharmacy Management System.

## Overview

Simpo Mobile is a React Native CLI application for Indonesian SME pharmacies. Built with TypeScript and organized using a feature-based architecture for maintainability and scalability.

## Tech Stack

- **Framework:** React Native CLI 0.85.3
- **Language:** TypeScript (strict mode)
- **State Management:** React Context + useReducer
- **Navigation:** React Navigation
- **Storage:** AsyncStorage

## Project Structure

```
apps/mobile/
├── src/
│   ├── features/           # Feature-based organization
│   │   ├── auth/          # Authentication (login, session)
│   │   ├── pos/           # Point of Sale functionality
│   │   ├── inventory/     # Inventory management
│   │   ├── reports/       # Financial reports
│   │   └── alerts/        # System alerts & notifications
│   ├── context/           # React Context providers
│   │   ├── AuthContext.tsx
│   │   └── AppProvider.tsx
│   ├── App.tsx            # Root component
│   └── index.js           # Entry point
├── android/               # Native Android project
├── ios/                   # Native iOS project
├── package.json
├── tsconfig.json
└── metro.config.js
```

## Development Setup

### Prerequisites

- Node.js 22+ (current: 24.1.0)
- Android Studio (for Android builds)
- JDK 17+
- Android SDK

### Installation

```bash
cd apps/mobile
npm install
```

### Running the App

**Start Metro Bundler:**
```bash
npm start
```

**Run on Android Emulator:**
```bash
npm run android
```

**Run on iOS Simulator:**
```bash
npm run ios
```

### Building APK

**Debug APK:**
```bash
cd android
./gradlew assembleDebug
```

**Release APK:**
```bash
cd android
./gradlew assembleRelease
```

## Available Scripts

| Command | Description |
|---------|-------------|
| `npm start` | Start Metro bundler |
| `npm run android` | Run on Android |
| `npm run ios` | Run on iOS |
| `npm test` | Run Jest tests |
| `npm run lint` | Run ESLint |

## Architecture

### Feature-Based Organization

Each feature is self-contained with its own components, screens, services, and types:

```
src/features/{feature}/
├── components/     # Feature-specific UI components
├── screens/        # Screen components for navigation
├── services/       # API calls and business logic
├── types/          # TypeScript interfaces
└── index.ts        # Public exports
```

### State Management

We use React Context + useReducer for simple, built-in state management:

- **AuthContext:** Authentication state, user session, JWT tokens
- **AppProvider:** Root provider wrapping all contexts

### Naming Conventions

- **Components:** PascalCase (e.g., `ProductCard.tsx`)
- **Screens:** PascalCase with "Screen" suffix (e.g., `LoginScreen.tsx`)
- **Services:** PascalCase with "Service" suffix (e.g., `AuthService.ts`)
- **Hooks:** camelCase with "use" prefix (e.g., `useAuth.ts`)

## Monorepo Integration

This is part of the simpo monorepo:

- `backend/` — Go backend (GRAB boilerplate)
- `apps/mobile/` — React Native CLI mobile app (this directory)
- `apps/admin/` — Next.js admin dashboard (future)

## Troubleshooting

### Metro bundler won't start

```bash
npm start -- --reset-cache
```

### Android build fails

```bash
cd android
./gradlew clean
cd ..
npm run android
```

### Type errors

```bash
npx tsc --noEmit
```

### Clear all caches

```bash
watchman watch-del-all
rm -rf node_modules
npm install
cd android && ./gradlew clean && cd ..
npm start -- --reset-cache
```

## Hardware Integration

### USB Barcode Scanner Support (Story 7.2)

The mobile POS supports USB HID barcode scanners for fast product scanning.

**Supported Scanner Types:**
- USB HID barcode scanners (plug-and-play)
- Common barcode formats: EAN-8, EAN-13, Code 128
- Tested brands: Zebra DS2200, Honeywell Eclipse 5145, Datalogic QuickScan

**Configuration:**
- Access scanner settings via gear icon in POS screen header
- Adjustable debounce interval (100ms - 2000ms)
- Configurable barcode length validation
- Feedback (vibration) can be toggled on/off

**Usage:**
1. Connect USB barcode scanner to mobile device
2. Scanner is automatically detected as keyboard input
3. Scan product barcode to add to cart automatically
4. Visual feedback shows scan success/failure

**Troubleshooting:**
- Scanner not working: Verify USB HID model (not serial/USB-COM)
- Duplicate scans: Increase debounce interval in settings
- Typing interference: Scanner input vs manual typing distinguished by timing

### Bluetooth Barcode Scanner Support (Story 7.3)

The mobile POS supports Bluetooth barcode scanners for wireless scanning flexibility.

**Supported Scanner Types:**
- BLE (Bluetooth Low Energy) barcode scanners
- Classic Bluetooth barcode scanners
- Common barcode formats: EAN-8, EAN-13, Code 128, UPC-A
- Tested brands: Zebra DS2200-BT, Honeywell Voyager 1202g-BF, Datalogic Gryphon I GD4500

**Configuration:**
- Access scanner settings via gear icon in POS screen header
- Bluetooth scanner section shows paired devices
- Auto-reconnect to last-used scanner (configurable)
- Connection status indicator in POS screen header

**Usage:**
1. Ensure Bluetooth is enabled on mobile device
2. Grant required permissions (Bluetooth, Location for Android)
3. Open Scanner Settings → Bluetooth Scanner section
4. Tap "Cari Perangkat Baru" to discover scanners
5. Select scanner to pair and connect
6. Scanner connection status shows in POS header
7. Scan product barcode to add to cart automatically

**Android Permissions:**
- `BLUETOOTH_SCAN` - Required for device discovery
- `BLUETOOTH_CONNECT` - Required for connection management
- `ACCESS_FINE_LOCATION` - Required for BLE scanning (system requirement)

**iOS Permissions:**
- Core Bluetooth framework properly configured
- Bluetooth Always permission required for background scanning

**Features:**
- Automatic reconnection with exponential backoff (1s, 2s, 4s, 8s delays)
- Connection state monitoring in background
- Manual connection/disconnection controls
- Device pairing and unpairing support
- Connection error handling with user notifications

**Troubleshooting:**
- Scanner not discovered: Ensure scanner is in pairing mode (check manual)
- Connection fails: Check if scanner is already paired to another device
- Frequent disconnections: Check battery level and signal strength
- Permissions denied: Clear app data and grant permissions again
- Scan not detected: Verify scanner is in "HID keyboard mode" (not inventory mode)

**Tested Models:**
- Zebra DS2200-BT (BLE)
- Zebra DS2278 (BLE)
- Honeywell Voyager 1202g-BF (Classic)
- Datalogic Gryphon I GD4500 (BLE)

### Thermal Printer Support (Story 7.1)

ESC/POS protocol support for 58mm and 80mm thermal printers.

### Cash Drawer Support (Story 7.4)

The mobile POS supports automatic cash drawer opening via thermal printer kick command for efficient cash payment processing.

**Supported Drawer Types:**
- RJ-12 interface cash drawers (connects to thermal printer)
- Standard and heavy-duty cash drawers
- Single and dual drawer configurations
- Tested brands: APG (VB series), Star Micronics (SCD/ECD series), custom/unbranded models

**How It Works:**
Cash drawers connect to thermal printers via RJ-12 cable (6-pin connector). When a CASH payment is processed, the printer sends an electrical pulse to specific pins that triggers the drawer solenoid to open. This happens automatically during receipt printing.

**ESC/POS Drawer Command:**
```
ESC p 0 m t1 t2
```
- **ESC** = 0x1B (escape character)
- **p** = 0x70 (printer command)
- **0** = drawer number (0 = drawer 1, 1 = drawer 2)
- **m** = mode (0x00 = pulse mode, 0x01 = steady mode)
- **t1** = pulse on time (in 2ms units)
- **t2** = pulse off time (in 2ms units)

**Example for 100ms pulse:**
```
0x1B 0x70 0x00 0x00 0x32 0x32
(t1 = 0x32 = 50 units × 2ms = 100ms, t2 = same)
```

**Common Pulse Timings by Drawer Model:**
- **Standard cash drawers:** 100-150ms
- **Heavy-duty drawers:** 150-200ms
- **Light-duty drawers:** 50-100ms
- **Custom/unbranded:** 100ms (default, configurable)

**RJ-12 Connector Pinout:**
- **Pin 1:** +24V (power)
- **Pin 2:** Drawer 1 trigger (most common)
- **Pin 3:** Ground
- **Pin 4:** +24V (power alternative)
- **Pin 5:** Drawer 2 trigger (dual drawer systems)
- **Pin 6:** Ground

**Configuration:**
- Access drawer settings via Printer Settings screen
- **Enable/Disable** automatic drawer opening toggle
- **Pulse Timing Slider:** 50ms - 500ms (default: 100ms)
- **Pin Selection:** Pin 2 (drawer 1) or Pin 5 (drawer 2)
- **Test Drawer Button:** Opens drawer without transaction (for testing)

**Usage:**
1. Connect cash drawer to thermal printer via RJ-12 cable
2. Connect thermal printer to mobile device via USB
3. Open Printer Settings → Cash Drawer Configuration section
4. Enable "Buka Laci Uang Otomatis" toggle
5. Adjust pulse timing if needed (start with 100ms)
6. Select correct drawer pin (Pin 2 for most drawers)
7. Tap "🧪 Tes Buka Laci" to test drawer opening
8. Complete CASH payment transaction to trigger automatic drawer opening

**When Drawer Opens:**
- **Payment Method:** CASH only (non-cash payments don't trigger drawer)
- **Timing:** After payment confirmation, before receipt completes
- **Automatic:** No manual intervention required during transaction
- **Audit Logged:** All drawer openings recorded for compliance

**Drawer Status Indicator (POSScreen Header):**
- **🟢 Green:** "Laci Uang: Terhubung" - Drawer ready (printer connected)
- **⚪ Gray:** "Laci Uang: Terputus" - Drawer disconnected (printer offline)
- **🟠 Orange:** "Laci Uang: Membuka..." - Drawer opening in progress
- **🔴 Red:** "Laci Uang: Gagal" - Drawer failed to open (see error)

**Error Handling:**
If drawer fails to open:
1. Red status indicator appears in POSScreen header
2. Toast notification: "Laci uang gagal dibuka - silakan buka manual"
3. **Transaction continues** - payment not blocked
4. Error logged to audit trail for investigation
5. Cashier can open drawer manually

**Audit Trail Logging:**
All drawer openings are logged for Badan POM compliance:

```json
{
  "eventType": "cash_drawer_open",
  "transactionId": "TRX-20240508-0001",
  "userId": "123",
  "timestamp": "2026-05-28T10:30:00Z",
  "metadata": {
    "drawerStatus": "opened",
    "pulseTiming": 100,
    "pinNumber": 2
  }
}
```

**Audit Log Behavior:**
- **Success:** Logged as `cash_drawer_open` event
- **Failure:** Logged as `cash_drawer_failed` event with error message
- **Offline Support:** Audit logs queued for sync when connectivity restored
- **Append-Only:** Logs never modified or deleted (compliance requirement)
- **5-Year Retention:** Minimum storage period for compliance
- **User Tracking:** Cashier ID logged for accountability

**Supported Cash Drawer Models (Indonesian Market):**

| Brand | Series | Type | Pulse Timing | Notes |
|-------|--------|------|--------------|-------|
| APG | VB130, VB320 | Standard | 100-150ms | Most common, reliable |
| APG | VB400 | Heavy-duty | 150-200ms | For high-volume stores |
| Star Micronics | SCD-211 | Standard | 100-150ms | Compact design |
| Star Micronics | ECD-201 | Standard | 100-150ms | Economy model |
| Custom/Unbranded | Generic | Light-duty | 50-100ms | Budget option |
| Dual Systems | Custom | Dual drawer | 100ms each | Two drawers support |

**Troubleshooting Common Issues:**

**1. Drawer Not Opening**
- **Check:** Printer connected and powered on (USB indicator light)
- **Check:** RJ-12 cable securely connected to both printer and drawer
- **Check:** Drawer configuration enabled (toggle ON)
- **Check:** Correct pin selected (Pin 2 for most drawers, Pin 5 for dual drawer)
- **Try:** Increase pulse timing to 150-200ms (weak solenoid needs longer pulse)
- **Try:** Test with different drawer to verify solenoid functionality
- **Verify:** Drawer solenoid not burned out (listen for click when testing)

**2. Drawer Opens But Doesn't Stay Open**
- **Cause:** Pulse timing too short
- **Solution:** Increase pulse timing in 25ms increments (100ms → 125ms → 150ms)
- **Check:** Drawer mechanical mechanism for obstruction
- **Check:** Drawer spring weak or broken (mechanical issue)

**3. Drawer Opens for Non-Cash Payments**
- **Check:** Payment method logic (only CASH should trigger)
- **Verify:** Configuration auto-open setting
- **Review:** Transaction payment method classification
- **Expected:** Transfer, E-Wallet, QRIS should NOT open drawer

**4. Drawer Opens Randomly Without Payment**
- **Cause:** Electrical interference or printer issue
- **Check:** RJ-12 cable for damage or loose connection
- **Check:** Printer firmware (may need update)
- **Check:** Grounding issues (power supply)
- **Solution:** Replace RJ-12 cable if damaged

**5. Audit Log Not Recording**
- **Check:** API connectivity (backend online)
- **Check:** Offline sync queue (settings → offline logs)
- **Verify:** AuditLogService initialized (check app logs)
- **Check:** Network permissions for audit endpoint
- **Expected:** Offline logs sync when connectivity restored

**6. Configuration Not Persisting**
- **Check:** AsyncStorage permissions granted
- **Verify:** PrinterConfigService.saveDrawerConfig called
- **Check:** App data cleared (settings reset on data clear)
- **Solution:** Reconfigure drawer settings after data clear

**7. Test Drawer Button Disabled**
- **Cause:** Printer not connected
- **Solution:** Connect printer via USB first
- **Expected:** Button enabled only when printer online

**8. Drawer Status Always Shows "Disconnected"**
- **Cause:** Printer connection issue (drawer follows printer)
- **Check:** Printer USB connection
- **Check:** Printer powered on
- **Solution:** Fix printer connection first, drawer follows printer status

**Performance Impact:**
- **Drawer opening adds:** ~300ms to transaction time (mechanical delay + pulse)
- **Total transaction time:** Still under 30 seconds (NFR requirement)
- **UI feedback:** Instant (<100ms) for drawer status updates
- **Audit logging:** Async (doesn't block transaction)

**Android Permissions:**
- Inherited from printer USB permissions (no new permissions required)
- Drawer uses printer USB connection for kick command

**iOS Limitations:**
- USB printer support limited on iOS (platform constraint)
- Bluetooth printer support available (future enhancement)

**Best Practices:**
1. **Test drawer** before first use using Test Drawer button
2. **Adjust pulse timing** for each new drawer model
3. **Check drawer solenoid** if drawer never opens (mechanical failure)
4. **Keep RJ-12 cable** away from power cables (interference prevention)
5. **Verify pin selection** matches drawer connection (Pin 2 vs Pin 5)
6. **Monitor audit logs** regularly for compliance verification
7. **Train cashiers** to manually open drawer if automatic fails (business continuity)

**Integration Notes:**
- Drawer control via ESC/POS command sent through PrinterManager
- Audit logging via AuditLogService with offline queue support
- Configuration persistence via PrinterConfigService (AsyncStorage)
- UI feedback via CashDrawerStatus component in POSScreen
- Error handling doesn't block transaction (business continuity priority)

**See Also:**
- Thermal Printer Support (Story 7.1)
- USB Barcode Scanner Support (Story 7.2)
- Bluetooth Barcode Scanner Support (Story 7.3)

## License

Copyright © 2025 simpo. All rights reserved.
