/**
 * Cart Context for POS feature
 * Manages transaction cart state with React Context + useReducer pattern
 * Provides cart actions: add, remove, update quantity, clear
 */

import React, { createContext, useContext, useReducer, ReactNode, useEffect } from 'react';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { Product } from '../types/product.types';
import { CartItem, CartState, CartContextType } from '../types/cart.types';

// Cart storage key for persistence
const CART_STORAGE_KEY = '@simpo_cart';

// Initial cart state
const initialCartState: CartState = {
  items: [],
  total: '0.00',
  itemCount: 0,
};

// Cart reducer actions
type CartAction =
  | { type: 'ADD_ITEM'; payload: CartItem & { stockQty?: number } }
  | { type: 'REMOVE_ITEM'; payload: number }
  | { type: 'UPDATE_QUANTITY'; payload: { productId: number; quantity: number } }
  | { type: 'CLEAR_CART' }
  | { type: 'LOAD_CART'; payload: CartState }
  | { type: 'SET_ITEMS'; payload: CartItem[] };

// Helper function to calculate subtotal with validation
const calculateSubtotal = (price: string, quantity: number): string => {
  // Validate price input
  if (!price || typeof price !== 'string') {
    return '0.00';
  }

  const priceNum = parseFloat(price);

  // Validate parsed price is valid number and positive
  if (isNaN(priceNum) || priceNum <= 0) {
    console.warn('Invalid price:', price);
    return '0.00';
  }

  // Validate quantity is positive
  if (!Number.isFinite(quantity) || quantity < 0) {
    console.warn('Invalid quantity:', quantity);
    return '0.00';
  }

  const subtotal = priceNum * quantity;
  return subtotal.toFixed(2);
};

// Helper function to calculate total from items with validation
const calculateTotal = (items: CartItem[]): string => {
  if (!Array.isArray(items)) {
    return '0.00';
  }

  const total = items.reduce((sum, item) => {
    const itemSubtotal = parseFloat(item.subtotal);
    return sum + (isNaN(itemSubtotal) ? 0 : itemSubtotal);
  }, 0);

  return total.toFixed(2);
};

// Cart reducer
const cartReducer = (state: CartState, action: CartAction): CartState => {
  switch (action.type) {
    case 'ADD_ITEM': {
      const MAX_CART_ITEMS = 100;

      // Check cart limit - prevent adding if would exceed maximum
      if (state.itemCount + action.payload.quantity > MAX_CART_ITEMS) {
        console.warn(`Cannot add: would exceed maximum cart size (${MAX_CART_ITEMS})`);
        // In production: could dispatch error action
        return state;
      }

      // Stock validation - check if product is in stock
      const productStock = action.payload.stockQty ?? 0;
      if (productStock <= 0) {
        console.warn('Cannot add: Product is out of stock');
        return state;
      }

      // Check if product already exists in cart
      const existingItemIndex = state.items.findIndex(
        item => item.productId === action.payload.productId
      );

      let newItems: CartItem[];

      if (existingItemIndex >= 0) {
        // Product exists - update quantity
        const existingItem = state.items[existingItemIndex];
        const newQuantity = existingItem.quantity + action.payload.quantity;

        // Validate stock availability
        if (newQuantity > productStock) {
          console.warn(`Cannot add: Only ${productStock} items in stock`);
          return state;
        }

        // Validate new quantity won't exceed cart limit
        if (state.itemCount - existingItem.quantity + newQuantity > MAX_CART_ITEMS) {
          console.warn('Cannot add: would exceed maximum cart size');
          return state;
        }

        const newSubtotal = calculateSubtotal(existingItem.price, newQuantity);

        newItems = [...state.items];
        newItems[existingItemIndex] = {
          ...existingItem,
          quantity: newQuantity,
          subtotal: newSubtotal,
        };
      } else {
        // New item - add to cart
        // Validate cart limit
        if (state.itemCount + action.payload.quantity > MAX_CART_ITEMS) {
          console.warn('Cannot add: would exceed maximum cart size');
          return state;
        }

        newItems = [...state.items, action.payload];
      }

      return {
        items: newItems,
        total: calculateTotal(newItems),
        itemCount: newItems.reduce((sum, item) => sum + item.quantity, 0),
      };
    }

    case 'REMOVE_ITEM': {
      const newItems = state.items.filter(item => item.productId !== action.payload);
      return {
        items: newItems,
        total: calculateTotal(newItems),
        itemCount: newItems.reduce((sum, item) => sum + item.quantity, 0),
      };
    }

    case 'UPDATE_QUANTITY': {
      const { productId, quantity } = action.payload;

      // Validate quantity is a valid number
      if (!Number.isFinite(quantity) || quantity < 0) {
        console.warn('Invalid quantity:', quantity);
        return state;
      }

      // Check if quantity would exceed cart limit
      const existingItem = state.items.find(item => item.productId === productId);
      if (existingItem && quantity > 0) {
        const currentItemCount = state.itemCount - existingItem.quantity;
        if (currentItemCount + quantity > 100) {
          console.warn('Cannot update: would exceed maximum cart size (100)');
          return state;
        }
      }

      if (quantity === 0) {
        // Remove item if quantity is 0
        const newItems = state.items.filter(item => item.productId !== productId);
        return {
          items: newItems,
          total: calculateTotal(newItems),
          itemCount: newItems.reduce((sum, item) => sum + item.quantity, 0),
        };
      }

      const newItems = state.items.map(item => {
        if (item.productId === productId) {
          const newSubtotal = calculateSubtotal(item.price, quantity);
          return { ...item, quantity, subtotal: newSubtotal };
        }
        return item;
      });

      return {
        items: newItems,
        total: calculateTotal(newItems),
        itemCount: newItems.reduce((sum, item) => sum + item.quantity, 0),
      };
    }

    case 'CLEAR_CART':
      return initialCartState;

    case 'LOAD_CART':
      return action.payload;

    case 'SET_ITEMS':
      return {
        items: action.payload,
        total: calculateTotal(action.payload),
        itemCount: action.payload.reduce((sum, item) => sum + item.quantity, 0),
      };

    default:
      return state;
  }
};

// Create context
export const CartContext = createContext<CartContextType | undefined>(undefined);

// Cart Provider component
interface CartProviderProps {
  children: ReactNode;
}

export const CartProvider = ({ children }: CartProviderProps) => {
  const [state, dispatch] = useReducer(cartReducer, initialCartState);
  const saveTimeoutRef = React.useRef<NodeJS.Timeout | null>(null);

  // Load cart from AsyncStorage on mount
  useEffect(() => {
    let isMounted = true;

    const loadCartFromStorage = async () => {
      try {
        const cartJson = await AsyncStorage.getItem(CART_STORAGE_KEY);
        if (cartJson && isMounted) {
          const savedCart: CartState = JSON.parse(cartJson);
          dispatch({ type: 'LOAD_CART', payload: savedCart });
        }
      } catch (error) {
        console.error('Failed to load cart from storage:', error);
        // Show user feedback about storage failure
        // In production: dispatch({ type: 'STORAGE_ERROR', payload: error.message });
      }
    };

    loadCartFromStorage();

    // Cleanup function
    return () => {
      isMounted = false;
      if (saveTimeoutRef.current) {
        clearTimeout(saveTimeoutRef.current);
      }
    };
  }, []);

  // Save cart to AsyncStorage whenever state changes (with debouncing)
  useEffect(() => {
    // Clear previous timeout
    if (saveTimeoutRef.current) {
      clearTimeout(saveTimeoutRef.current);
    }

    // Debounce save operation
    saveTimeoutRef.current = setTimeout(async () => {
      try {
        await AsyncStorage.setItem(CART_STORAGE_KEY, JSON.stringify(state));
        // Save successful - could dispatch success action if needed
      } catch (error) {
        console.error('Failed to save cart to storage:', error);
        // Notify user of potential data loss
        // In production: dispatch({ type: 'STORAGE_ERROR', payload: error.message });
      }
    }, 300); // 300ms debounce

    // Cleanup on unmount
    return () => {
      if (saveTimeoutRef.current) {
        clearTimeout(saveTimeoutRef.current);
      }
    };
  }, [state]);

  // Helper function to convert Product to CartItem with stock info
  const productToCartItem = (product: Product, quantity: number = 1): CartItem & { stockQty: number } => {
    return {
      productId: product.id,
      sku: product.sku,
      name: product.name,
      price: product.price,
      quantity,
      subtotal: calculateSubtotal(product.price, quantity),
      stockQty: product.stockQty, // Include stock for validation
    };
  };

  // Cart actions
  const actions = {
    addItem: (product: Product) => {
      const cartItem = productToCartItem(product);
      dispatch({ type: 'ADD_ITEM', payload: cartItem });
    },

    removeItem: (productId: number) => {
      dispatch({ type: 'REMOVE_ITEM', payload: productId });
    },

    updateQuantity: (productId: number, quantity: number) => {
      dispatch({ type: 'UPDATE_QUANTITY', payload: { productId, quantity } });
    },

    clearCart: () => {
      dispatch({ type: 'CLEAR_CART' });
    },
  };

  const contextValue: CartContextType = {
    state,
    actions,
  };

  return (
    <CartContext.Provider value={contextValue}>
      {children}
    </CartContext.Provider>
);
};

// Custom hook to use cart context
export const useCartContext = (): CartContextType => {
  const context = useContext(CartContext);
  if (context === undefined) {
    throw new Error('useCartContext must be used within CartProvider');
  }
  return context;
};
