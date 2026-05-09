# simpo Web Admin Dashboard

Web admin dashboard for the simpo pharmacy management system. Built with Next.js 16, TypeScript, and Tailwind CSS.

## Overview

This dashboard provides business oversight for **Pharmacy Owners** and system configuration for **System Admins**.

### Key Features

- 📊 **Dashboard** - Real-time business metrics and KPIs
- 💊 **Products** - Inventory management with search and filters
- 📈 **Reports** - Financial reports (daily sales, profit/loss)
- 👥 **Users** - User management with role-based access control
- ⚙️ **Settings** - System configuration and health monitoring

## Tech Stack

- **Framework:** Next.js 16.2.6 (App Router with React Server Components)
- **Language:** TypeScript 5 (strict mode)
- **Styling:** Tailwind CSS v4
- **State Management:** React Context + Server Components
- **API Client:** Axios with RFC 7807 error handling
- **Monorepo:** Part of simpo monorepo structure

## Getting Started

### Prerequisites

- Node.js 18+ or later
- npm, yarn, or pnpm

### Installation

```bash
# Install dependencies
npm install
```

### Development

```bash
# Start development server
npm run dev

# Or with Turbopack (faster builds)
npm run dev -- --turbo
```

Open [http://localhost:3000](http://localhost:3000) with your browser.

### Build

```bash
# Production build
npm run build

# Start production server
npm start
```

### Lint

```bash
# Run ESLint
npm run lint
```

## Project Structure

```
apps/web/
├── app/                      # App Router directory
│   ├── (auth)/               # Authenticated route group
│   │   ├── layout.tsx         # Protected layout (sidebar, header)
│   │   ├── page.tsx           # Dashboard
│   │   ├── products/          # Product management
│   │   ├── reports/           # Financial reports
│   │   ├── users/             # User management
│   │   └── settings/          # System settings
│   ├── login/                 # Login page (public)
│   ├── layout.tsx             # Root layout
│   └── globals.css            # Global styles
├── components/               # React components
│   ├── ui/                   # Shadcn/ui components
│   ├── layout/               # Layout components
│   └── features/             # Feature-specific components
├── context/                  # React Context providers
│   ├── AuthContext.tsx       # Authentication state
│   └── AppProvider.tsx       # Root provider wrapper
├── lib/                      # Utilities
│   ├── apiClient.ts          # Backend API client
│   ├── auth.ts               # Authentication utilities
│   └── utils.ts              # Helper functions
├── types/                    # TypeScript types
│   └── api.ts                # API response types
├── public/                   # Static assets
├── next.config.ts            # Next.js configuration
├── tailwind.config.ts        # Tailwind CSS v4 configuration
├── tsconfig.json             # TypeScript configuration
└── package.json              # Dependencies and scripts
```

## Environment Variables

Create a `.env.local` file in the root directory:

```bash
# API Configuration
NEXT_PUBLIC_API_URL=http://localhost:8081/api/v1

# App Configuration
NEXT_PUBLIC_APP_NAME=simpo
NEXT_PUBLIC_APP_VERSION=1.0.0
```

## App Router Structure

### Route Groups

- `(auth)` - Authenticated routes (requires login)
  - All pages under this group share the same layout with sidebar and header
  - Middleware will redirect unauthenticated users to login
- `login` - Public route (no authentication required)

### Layout Hierarchy

```
app/
├── layout.tsx              # Root layout (html, head, body)
├── (auth)/
│   ├── layout.tsx          # Authenticated layout (sidebar, header)
│   ├── page.tsx            # Dashboard
│   └── [feature]/
│       └── page.tsx        # Feature pages
└── login/
    └── page.tsx            # Login page
```

## API Integration

### Backend Connection

The web dashboard connects to the simpo backend API:

- **API Base URL:** `http://localhost:8081/api/v1` (development)
- **Authentication:** JWT tokens via httpOnly cookies
- **Error Handling:** RFC 7807 (Problem Details) format

### API Client Usage

```typescript
import apiClient from '@/lib/apiClient';
import type { User } from '@/types/api';

// Get all products
const products = await apiClient.get<Product[]>('/products');

// Get user by ID
const user = await apiClient.get<User>(`/users/${id}`);

// Create new transaction
const transaction = await apiClient.post('/transactions', {
  cashierId: 1,
  items: [...],
  paymentMethod: 'CASH',
});
```

## State Management

### Server Components (Default)

Next.js 16 uses React Server Components by default for data fetching:

```typescript
// app/(auth)/products/page.tsx
async function getProducts() {
  const res = await fetch('http://localhost:8081/api/v1/products', {
    cache: 'no-store',
  });
  const products = await res.json();
  return products;
}

export default async function ProductsPage() {
  const products = await getProducts();
  return <ProductList products={products} />;
}
```

### Client Components (React Context)

For interactive state, use React Context:

```typescript
'use client';

import { useAuth } from '@/context/AuthContext';

export default function LoginForm() {
  const { login } = useAuth();
  // ... interactive UI
}
```

## Styling with Tailwind CSS v4

This project uses Tailwind CSS v4 with inline theme configuration:

```tsx
<div className="bg-white p-6 rounded-lg border shadow-sm">
  <h2 className="text-2xl font-bold mb-4">Dashboard</h2>
</div>
```

### Custom Theme

Custom theme colors are configured in `app/globals.css`:

```css
@theme inline {
  --color-primary: #0ea5e9;
  --color-secondary: #6366f1;
  /* ... */
}
```

## Monorepo Structure

This web dashboard is part of the simpo monorepo:

```
simpo/
├── backend/          # Go backend (GRAB boilerplate)
├── apps/
│   ├── mobile/       # React Native POS app
│   └── web/          # ← Next.js admin dashboard (this project)
└── docker-compose.yml # Infrastructure services
```

## Development Workflow

### Recommended Workflow

1. **Start development server:**
   ```bash
   npm run dev
   ```

2. **Start backend API** (in separate terminal):
   ```bash
   cd ../backend && air
   ```

3. **Start infrastructure services** (if needed):
   ```bash
   cd ../../ && docker-compose up -d
   ```

### Hot Reload

Next.js provides fast refresh for React components. Changes to code will automatically reflect in the browser.

### TypeScript

TypeScript strict mode is enabled. This ensures type safety throughout the application.

## Troubleshooting

### Port Already in Use

If port 3000 is already in use:

```bash
# Kill process on port 3000 (macOS/Linux)
lsof -ti:3000 | xargs kill -9

# Or use a different port
PORT=3001 npm run dev
```

### Module Not Found Errors

If you encounter module not found errors:

```bash
# Clear Next.js cache
rm -rf .next

# Reinstall dependencies
rm -rf node_modules package-lock.json
npm install
```

### API Connection Issues

If the dashboard can't connect to the backend:

1. Verify backend is running on port 8081
2. Check CORS configuration in backend
3. Verify API_URL in `.env.local`

### Build Errors

If production build fails:

```bash
# Clear cache and rebuild
rm -rf .next
npm run build
```

## Deployment

### Production Build

```bash
npm run build
npm start
```

### Environment Variables

Ensure all environment variables are set in production:

- `NEXT_PUBLIC_API_URL` - Backend API URL
- `NEXT_PUBLIC_APP_NAME` - Application name
- `NEXT_PUBLIC_APP_VERSION` - Application version

## Browser Support

- Chrome 90+
- Edge 90+
- Firefox 88+
- Safari 15+

## Accessibility

This dashboard follows WCAG 2.1 Level A guidelines:
- Keyboard navigation support
- Semantic HTML
- ARIA labels where needed
- Sufficient color contrast

## Performance

- **Initial Page Load:** <3 seconds
- **Subsequent Interactions:** <500ms
- **Image Optimization:** Automatic via Next.js Image component

## Security

- **JWT Tokens:** Stored in httpOnly cookies
- **CORS:** Configured for allowed origins
- **HTTPS:** Required in production
- **Content Security Policy:** Configured via Next.js headers

## Contributing

This is an internal project for simpo pharmacy management system.

## License

Copyright © 2026 simpo. All rights reserved.

## Support

For issues or questions, please contact the development team.

---

**Built with ❤️ for Indonesian pharmacies**
