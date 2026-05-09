# Story 1.2: Initialize Mobile POS App with React Native CLI

Status: ready-for-dev

**Epic:** 1 - Authentication & User Management
**Priority:** Foundation (Second Story)
**Story Type:** Project Initialization

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

---

## Story

**As a** Development Team,
**I want to** create the mobile POS application using React Native CLI with TypeScript template,
**So that** cashiers have a modern React Native foundation for the point-of-sale interface with full native control.

---

## Acceptance Criteria

1. **AC1:** Project is initialized with React Native CLI and TypeScript
   - React Native CLI is used (NOT Expo)
   - TypeScript is properly configured with strict mode
   - Type definitions are available for all dependencies
   - tsconfig.json is configured for React Native development

2. **AC2:** Project structure includes features/ directory for feature-based organization
   - features/ directory exists at root level
   - Feature subdirectories are prepared: auth/, pos/, inventory/, reports/, alerts/
   - Each feature directory has placeholder structure (components/, screens/, services/, types/)

3. **AC3:** Android configuration is set up for local builds
   - android/ directory exists with Gradle configuration
   - package name is configured (com.simpo.app or similar)
   - Android minimum SDK version is configured appropriately (API 21+)
   - App version and build number are set to 1.0.0 and 1 respectively

4. **AC4:** Local build environment is configured
   - Android Studio can open the project
   - Gradle build works: `cd android && ./gradlew build`
   - APK can be generated locally
   - Debug signing is configured for development

5. **AC5:** Development environment is fully operational
   - Metro bundler starts: `npm start`
   - App can run on Android emulator/device: `npm run android`
   - Hot reload works for rapid development
   - TypeScript compilation succeeds without errors

6. **AC6:** Project configuration follows simpo standards
   - .gitignore includes node_modules, .env, dist/, build/
   - package.json includes appropriate scripts (start, android, test, lint)
   - ESLint is configured for React Native + TypeScript
   - Project is ready for monorepo structure (simpo-mobile/ directory)

7. **AC7:** React Native and dependencies versions are verified
   - React Native version is latest stable (0.73+)
   - React version is 18+ compatible
   - Node.js version is 18+ (React Native requirement)
   - All dependencies install successfully

---

## Tasks / Subtasks

- [ ] **Task 1: Create React Native Project with TypeScript** (AC: 1, 5, 7)
  - [ ] Run `npx react-native@latest init SimpoMobile --template react-native-template-typescript`
  - [ ] Verify TypeScript is configured in tsconfig.json
  - [ ] Verify package.json includes TypeScript dependencies
  - [ ] Verify React Native version is 0.73+ in package.json
  - [ ] Verify React version compatibility

- [ ] **Task 2: Configure Feature-Based Directory Structure** (AC: 2)
  - [ ] Create features/ directory at root level
  - [ ] Create feature subdirectories: auth/, pos/, inventory/, reports/, alerts/
  - [ ] Create placeholder structure in each feature: components/, screens/, services/, types/
  - [ ] Create index.ts files for exports from each feature
  - [ ] Document the feature-based structure in README

- [ ] **Task 3: Configure Android Build Settings** (AC: 3, 4)
  - [ ] Update android/app/build.gradle with package name (com.simpo.app)
  - [ ] Set Android minSdkVersion to 21 (or appropriate version)
  - [ ] Configure app version and versionCode in android/app/build.gradle
  - [ ] Verify debug signing configuration in android/app/build.gradle
  - [ ] Test Gradle build: `cd android && ./gradlew build`

- [ ] **Task 4: Set Up Metro Bundler Configuration** (AC: 5)
  - [ ] Verify metro.config.js exists and is properly configured
  - [ ] Verify Metro can start: `npm start`
  - [ ] Test Metro bundler with cache reset: `npm start -- --reset-cache`
  - [ ] Configure port if needed (default 8081)

- [ ] **Task 5: Configure Development Tools** (AC: 5, 6)
  - [ ] Verify app runs on Android emulator: `npm run android`
  - [ ] Test hot reload by modifying App.tsx
  - [ ] Configure ESLint for React Native + TypeScript
  - [ ] Set up .gitignore for node_modules, .env, dist/, build/
  - [ ] Add npm scripts: start, android, test, lint

- [ ] **Task 6: Set Up State Management Foundation** (AC: 2)
  - [ ] Create src/context/ directory for React Context providers
  - [ ] Create AuthContext placeholder for future authentication
  - [ ] Create AppProvider component to wrap all contexts
  - [ ] Document state management approach (React Context + useReducer)

- [ ] **Task 7: Verify Monorepo Integration** (AC: 6)
  - [ ] Confirm SimpoMobile/ directory structure matches monorepo pattern
  - [ ] Update root .gitignore if needed for monorepo
  - [ ] Document the monorepo structure in project README
  - [ ] Verify no conflicts with backend/ directory (from Story 1.1)

- [ ] **Task 8: Create Initial Documentation**
  - [ ] Update README.md with SimpoMobile specific setup instructions
  - [ ] Document React Native CLI commands for development
  - [ ] Document Android Studio setup for local builds
  - [ ] Add troubleshooting section for common React Native issues

- [ ] **Task 9: Install Additional Dependencies** (AC: 7)
  - [ ] Install React Navigation: `npm install @react-navigation/native @react-navigation/stack`
  - [ ] Install dependencies for Android: `npm install react-native-screens react-native-safe-area-context`
  - [ ] Link native dependencies if needed: `npx pod-install` (for iOS future)
  - [ ] Install secure storage: `npm install @react-native-async-storage/async-storage`
  - [ ] Verify all packages install successfully

---

## Dev Notes

### Context & Purpose

This is the **second foundational story** for the simpo mobile POS application. All subsequent mobile stories will build upon this React Native CLI foundation. 

**IMPORTANT:** We are using **React Native CLI directly, NOT Expo**. This decision provides:
- Full control over native code
- Direct access to native modules for hardware integration
- Local build process with Android Studio
- No dependency on Expo platform
- More complex setup but greater flexibility

### Architecture Alignment

**[Source: _bmad-output/planning-artifacts/architecture.md]**

**Technical Stack Decisions (REVISED for RN CLI):**
- **Mobile Framework:** React Native CLI (latest stable, 0.73+)
- **Language:** TypeScript (strict mode)
- **Runtime:** Metro bundler, React 18+ compatible
- **State Management:** React Context + useReducer (simple, built into React)
- **Navigation:** React Navigation (standard for RN CLI projects)
- **Build System:** Local Gradle builds (Android Studio or command line)
- **Signing:** Debug signing for development, release signing for production

**Why React Native CLI over Expo?**
- Full native control for hardware integration (thermal printers, barcode scanners)
- No OTA update complexity for MVP
- Local build process gives full control
- Direct access to native modules when needed
- Team has native development capability

### React Native CLI Initialization Specifics

**Initialization Commands:**
```bash
# Create React Native project with TypeScript template
npx react-native@latest init SimpoMobile --template react-native-template-typescript

# Navigate to project
cd SimpoMobile

# Install dependencies
npm install

# Start Metro bundler
npm start

# Run on Android (emulator or device)
npm run android
```

**What React Native CLI Provides:**
- ✅ TypeScript configured with strict mode
- ✅ Metro bundler for fast refresh
- ✅ React 18+ compatibility
- ✅ Hot reload enabled by default
- ✅ Native Android and iOS projects
- ✅ Gradle build system
- ✅ Direct native module access
- ❌ No OTA updates (manual or CodePush needed)
- ❌ No Expo Go for instant testing

### Project Structure Requirements

**[Source: _bmad-output/planning-artifacts/architecture.md - Component Architecture]**

**Monorepo Organization (REVISED):**
```
simpo/
├── backend/              # Story 1.1 (GRAB boilerplate)
├── SimpoMobile/          # ← Story 1.2 (React Native CLI) - this story
│   ├── src/              # Source code (React Native template uses src/)
│   │   ├── App.tsx       # Root component
│   │   ├── features/     # Feature-based organization
│   │   │   ├── auth/     # Login, session management
│   │   │   │   ├── components/
│   │   │   │   ├── screens/
│   │   │   │   ├── services/
│   │   │   │   ├── types/
│   │   │   │   └── index.ts
│   │   │   ├── pos/      # Point of Sale
│   │   │   ├── inventory/
│   │   │   ├── reports/
│   │   │   └── alerts/
│   │   ├── context/      # React Context providers
│   │   │   ├── AuthContext.tsx
│   │   │   └── AppProvider.tsx
│   │   ├── components/   # Shared UI components
│   │   ├── hooks/        # Custom React hooks
│   │   ├── utils/        # Helper functions
│   │   └── types/        # Shared TypeScript types
│   ├── android/          # Native Android project (Gradle)
│   ├── ios/              # Native iOS project (CocoaPods)
│   ├── index.js          # Entry point
│   ├── metro.config.js   # Metro bundler config
│   ├── package.json      # Dependencies and scripts
│   ├── tsconfig.json     # TypeScript configuration
│   ├── .gitignore        # Git ignore patterns
│   └── babel.config.js   # Babel configuration
├── simpo-admin/          # Story 1.3 (Next.js)
└── docker-compose.yml    # Local development infrastructure
```

### State Management Requirements

**[Source: _bmad-output/planning-artifacts/architecture.md - Decision 12]**

**Decision: React Context + useReducer**
- **Rationale:** Simple, built into React, sufficient for MVP scope, no additional dependencies
- **Affects:** Mobile app architecture, data flow patterns

**Implementation:**
```
src/context/
├── AuthContext.tsx       # Authentication state and actions
├── AppProvider.tsx       # Wraps all context providers
└── types/
    └── AuthContext.types.ts
```

**Why No Redux/MobX?**
- Rule of Three before abstraction
- Boring technology for stability
- MVP scope doesn't warrant complex state management
- Can add Redux later if complexity grows

### Feature-Based Organization

**[Source: _bmad-output/planning-artifacts/architecture.md - Decision 14]**

**Decision: Feature-based organization**
- **Rationale:** Maps to PRD capabilities (POS, Inventory, Reports), easier for AI agents to understand business logic, better for phased development
- **Affects:** Code organization, team workflow

**Feature Structure Template:**
```
src/features/{feature}/
├── components/     # Feature-specific components
├── screens/         # Screen components for navigation
├── services/        # API calls and business logic
├── types/           # TypeScript interfaces and types
└── index.ts         # Public exports
```

### Android Build Configuration

**Local Build Setup:**
```bash
# Build debug APK
cd android
./gradlew assembleDebug

# Build release APK
./gradlew assembleRelease

# Install debug APK to device
./gradlew installDebug

# Clean build
./gradlew clean
```

**Android Studio:**
- Open SimpoMobile/android/ in Android Studio
- Sync Gradle files
- Run app using Android Studio's Run button
- Debug native code if needed

### Navigation Setup

**React Navigation Installation:**
```bash
# Install core navigation
npm install @react-navigation/native @react-navigation/stack

# Install dependencies for Android
npm install react-native-screens react-native-safe-area-context

# Link dependencies (if needed for RN version)
npx pod-install  # Only for iOS, can skip for Android MVP
```

### Naming Conventions

**[Source: _bmad-output/planning-artifacts/architecture.md - Naming Patterns]**

**File Naming Conventions:**
- **Components:** PascalCase (e.g., `ProductCard.tsx`)
- **Screens:** PascalCase with "Screen" suffix (e.g., `LoginScreen.tsx`)
- **Services:** PascalCase with "Service" suffix (e.g., `AuthService.ts`)
- **Types:** PascalCase (e.g., `User.ts`, `Product.ts`)
- **Hooks:** camelCase with "use" prefix (e.g., `useAuth.ts`)

### Testing Requirements

**[Source: _bmad-output/planning-artifacts/architecture.md - Testing Standards]**

**Testing Framework for React Native CLI:**
- Jest + React Native Testing Library
- React Native template includes test setup

**Verification for This Story:**
- Run `npm test` to verify test setup works
- Ensure TypeScript compilation succeeds
- Verify Metro bundler starts without errors
- Verify app runs on Android emulator

### Dependencies to Verify

**Critical React Native Packages:**
- `react-native` 0.73+ (core framework)
- `react` 18+ (UI library)
- `typescript` (type checking)
- `@types/react` (React type definitions)
- `@react-navigation/native` (navigation)
- `@react-navigation/stack` (stack navigator)
- `react-native-screens` (navigation support)
- `@react-native-async-storage/async-storage` (secure storage)

**Version Verification:**
- React Native: 0.73+ (latest stable)
- React: 18+ (compatible with RN 0.73)
- Node.js: 18+ (React Native requirement)
- npm or yarn: Latest stable version

### Common Pitfalls to Avoid

1. **Do NOT skip features/ directory structure** - Feature-based organization is critical
2. **Do NOT ignore TypeScript strict mode** - Type safety prevents runtime errors
3. **Do NOT skip Android Studio setup** - Local builds require proper Gradle configuration
4. **Do NOT ignore native module linking** - Some packages need manual linking
5. **Do NOT add unnecessary dependencies** - Keep it minimal (Rule of Three)
6. **Do NOT ignore state management foundation** - Set up Context providers
7. **Do NOT hardcode values** - Use environment variables for configuration

### Verification Checklist

Before marking this story complete, verify:

- [ ] `npm start` starts Metro bundler without errors
- [ ] `npm run android` runs app on emulator/device
- [ ] TypeScript compilation succeeds: `npx tsc --noEmit`
- [ ] Hot reload works when modifying App.tsx
- [ ] features/ directory exists with all feature subdirectories
- [ ] android/app/build.gradle has correct package name
- [ ] Gradle build works: `cd android && ./gradlew build`
- [ ] ESLint is configured and runs without errors
- [ ] .gitignore includes node_modules, .env, dist/, build/
- [ ] package.json includes appropriate scripts
- [ ] README.md documents setup and development process
- [ ] React Navigation is installed and configured
- [ ] Project structure follows monorepo pattern (SimpoMobile/ directory)

### Previous Story Intelligence

**From Story 1.1 (Initialize Backend Project with GRAB Boilerplate):**

**Learnings to Apply:**
- Backend is set up in `backend/` directory using GRAB boilerplate
- GRAB provides JWT authentication and RBAC that mobile app will integrate with
- Backend API will follow REST pattern: `/api/v1/{resource}`
- Backend uses Swagger for API documentation at `/api/docs`
- Backend expects JWT tokens in Authorization header: `Bearer {token}`

**Integration Points:**
- Mobile app will call backend API endpoints for authentication
- JWT tokens from backend will be stored using AsyncStorage
- API base URL should be configurable via environment variable
- Use axios or fetch for API calls (choose one and stick to it)

### Hardware Integration Readiness

**Why React Native CLI helps for hardware:**
- **Thermal Printers:** Direct access to native ESC/POS libraries
- **Barcode Scanners:** USB HID and Bluetooth APIs directly accessible
- **Cash Drawers:** Native control via printer kick commands
- **No Expo Go limitation:** Can test on real hardware immediately

### Project Context Reference

**[Source: _bmad-output/planning-artifacts/prd.md]**

**Business Context:**
- simpo is a cost-effective pharmacy management system for Indonesian SME pharmacies
- Mobile POS app is used by cashiers for daily transaction processing
- Android-first deployment for MVP (iOS future consideration)
- Cashiers need fast, reliable POS interface for sub-30-second transactions

**Technical Constraints:**
- Must work offline for unreliable Indonesian internet connectivity
- Must integrate with thermal printers (ESC/POS protocol)
- Must integrate with barcode scanners (USB HID and Bluetooth)
- Must support cash drawer control via printer kick command

### References

- [React Native Documentation](https://reactnative.dev/docs/getting-started)
- [React Navigation](https://reactnavigation.org/docs/getting-started)
- [Android Studio Setup](https://developer.android.com/studio)
- [Source: _bmad-output/planning-artifacts/architecture.md - Mobile State Management]
- [Source: _bmad-output/planning-artifacts/epics.md - Story 1.2]
- [Source: _bmad-output/planning-artifacts/prd.md - Executive Summary]

---

## Dev Agent Record

### Agent Model Used

_Generated by create-story workflow_

### Debug Log References

_Story not yet implemented - no debug log_

### Completion Notes List

_Story not yet implemented - awaiting dev-story workflow_

### File List

_Files to be created/modified during implementation:_

**NEW Files:**
- `SimpoMobile/` - React Native project created here
- `SimpoMobile/src/` - Source code directory
- `SimpoMobile/src/App.tsx` - Root component
- `SimpoMobile/src/features/` - Feature-based directory structure
- `SimpoMobile/src/features/auth/` - Authentication feature
- `SimpoMobile/src/features/pos/` - Point of Sale feature
- `SimpoMobile/src/features/inventory/` - Inventory feature
- `SimpoMobile/src/features/reports/` - Reports feature
- `SimpoMobile/src/features/alerts/` - Alerts feature
- `SimpoMobile/src/context/` - React Context providers
- `SimpoMobile/android/` - Native Android project (from RN CLI)
- `SimpoMobile/package.json` - Dependencies and scripts (from RN CLI)
- `SimpoMobile/tsconfig.json` - TypeScript configuration (from RN CLI)
- `SimpoMobile/metro.config.js` - Metro bundler config (from RN CLI)
- `SimpoMobile/.gitignore` - Git ignore patterns (from RN CLI)
- `SimpoMobile/README.md` - Project documentation

**MODIFIED Files:**
- `.gitignore` - Ensure SimpoMobile/ is tracked (root monorepo .gitignore)

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2026-05-09 | Story created via create-story workflow (React Native CLI) | BMad System |
