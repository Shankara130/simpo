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

## Hardware Integration (Future)

- **Thermal Printers:** ESC/POS protocol support
- **Barcode Scanners:** USB HID and Bluetooth
- **Cash Drawers:** Printer kick command control

## License

Copyright © 2025 simpo. All rights reserved.
