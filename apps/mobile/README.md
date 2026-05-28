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

### Thermal Printer Support (Story 7.1)

ESC/POS protocol support for 58mm and 80mm thermal printers.

### Future Hardware Support

- **Bluetooth Barcode Scanners:** Wireless scanner support (Story 7.3)
- **Cash Drawers:** Printer kick command control (Story 7.4)

## License

Copyright © 2025 simpo. All rights reserved.
