/**
 * Cart type definitions for POS feature
 * Defines cart state, cart items, and cart context interfaces
 */

export interface CartItem {
  productId: number;
  sku: string;
  name: string;
  price: string; // Decimal as string for precision
  quantity: number;
  subtotal: string;
  stockQty: number; // Required: Current stock level for validation
}

export interface CartState {
  items: CartItem[];
  total: string; // Calculated from items
  itemCount: number;
}

export interface CartContextType {
  state: CartState;
  actions: {
    addItem: (product: Product) => void;
    removeItem: (productId: number) => void;
    updateQuantity: (productId: number, quantity: number) => void;
    clearCart: () => void;
  };
}
