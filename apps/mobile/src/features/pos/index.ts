/**
 * Public API exports for POS feature
 * Exports screens, components, context, and navigation
 */

// Screens
export { POSScreen } from './screens/POSScreen';

// Components
export { ProductCard } from './components/ProductCard';
export { CartSummary } from './components/CartSummary';
export { TopControlBar } from './components/TopControlBar';
export { ProductList } from './components/ProductList';
export { ActionButtons } from './components/ActionButtons';
export { ScannerFeedback } from './components/ScannerFeedback';

// Context
export { CartProvider, useCartContext } from './context/CartContext';

// Hooks
export { useBarcodeScanner } from './hooks/useBarcodeScanner';

// Navigation
export { POSNavigator } from './navigation/POSNavigator';

// Types
export type { Product } from './types/product.types';
export type { CartItem, CartState, CartContextType } from './types/cart.types';
export type { POSStackParamList, POSStackScreenName } from './types/navigation.types';
export type {
  ScannerState,
  ScannerConfig,
  ScannerInputState,
  ScannerResult,
  ScannerErrorType,
  ScannerError,
  ScannerCallbacks,
} from './types/scanner.types';
