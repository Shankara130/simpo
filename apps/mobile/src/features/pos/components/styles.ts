/**
 * Common styles and theming for POS components
 * Centralized style tokens for consistency across the POS feature
 */

import { StyleSheet } from 'react-native';

/**
 * Color Palette
 * Semantic colors for different UI states
 */
export const colors = {
  // Primary colors
  primary: '#2196F3',      // Blue - Primary actions, links
  primaryDark: '#1976D2',  // Dark blue - Pressed states

  // Secondary colors
  secondary: '#4CAF50',    // Green - Success, payment, checkout
  secondaryDark: '#388E3C', // Dark green - Pressed states

  // Accent colors
  accent: '#FF9800',       // Orange - Warning, low stock

  // Error colors
  error: '#F44336',        // Red - Error, out of stock, delete
  errorDark: '#D32F2F',    // Dark red - Pressed states

  // Neutral colors
  background: '#F5F5F5',   // Light gray - Page background
  surface: '#FFFFFF',      // White - Card, panel background
  border: '#E0E0E0',       // Light gray - Borders, dividers

  // Text colors
  text: {
    primary: '#212121',    // Dark gray - Primary text
    secondary: '#757575',  // Medium gray - Secondary text, captions
    disabled: '#BDBDBD',   // Light gray - Disabled text
    inverse: '#FFFFFF',    // White - Text on dark backgrounds
  },

  // Status colors
  success: '#4CAF50',      // Green - Success messages
  warning: '#FF9800',      // Orange - Warning messages
  info: '#2196F3',         // Blue - Info messages
};

/**
 * Spacing Tokens
 * Consistent spacing values for margins and paddings
 */
export const spacing = {
  xs: 4,    // Extra small - Tight spacing
  sm: 8,    // Small - Compact spacing
  md: 12,   // Medium - Comfortable spacing
  lg: 16,   // Large - Section spacing
  xl: 24,   // Extra large - Major sections
  xxl: 32,  // Extra extra large - Page margins
};

/**
 * Typography
 * Font sizes and weights for different text hierarchies
 */
export const typography = {
  // Font sizes
  fontSize: {
    xs: 12,    // Captions, labels
    sm: 14,    // Body text, secondary
    md: 16,    // Default text, inputs
    lg: 18,    // Subheadings
    xl: 20,    // Headings
    xxl: 24,   // Large headings
  },

  // Font weights
  fontWeight: {
    regular: '400' as const,
    medium: '500' as const,
    semibold: '600' as const,
    bold: '700' as const,
  },

  // Line heights
  lineHeight: {
    tight: 1.2,
    normal: 1.5,
    relaxed: 1.8,
  },
};

/**
 * Border Radius
 * Consistent border radius values
 */
export const borderRadius = {
  sm: 4,    // Small - Tags, badges
  md: 8,    // Medium - Cards, buttons
  lg: 12,   // Large - Panels
  xl: 16,   // Extra large - Modals
  full: 9999, // Fully rounded - Pills, circular buttons
};

/**
 * Touch Targets
 * Minimum touch target sizes for accessibility
 */
export const touchTarget = {
  minHeight: 44,  // Minimum - WCAG AAA standard
  recommended: 48, // Recommended - Better usability
};

/**
 * Common Styles
 * Pre-built style combinations for common use cases
 */
export const commonStyles = StyleSheet.create({
  // Container styles
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },

  surface: {
    backgroundColor: colors.surface,
  },

  // Card styles
  card: {
    backgroundColor: colors.surface,
    borderRadius: borderRadius.md,
    padding: spacing.lg,
    marginBottom: spacing.sm,
    borderWidth: 1,
    borderColor: colors.border,
  },

  // Button styles
  button: {
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    borderRadius: borderRadius.md,
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: touchTarget.minHeight,
  },

  buttonPrimary: {
    backgroundColor: colors.primary,
  },

  buttonSecondary: {
    backgroundColor: colors.secondary,
  },

  buttonError: {
    backgroundColor: colors.error,
  },

  buttonText: {
    color: colors.text.inverse,
    fontSize: typography.fontSize.md,
    fontWeight: typography.fontWeight.semibold,
  },

  // Text styles
  textPrimary: {
    color: colors.text.primary,
    fontSize: typography.fontSize.md,
  },

  textSecondary: {
    color: colors.text.secondary,
    fontSize: typography.fontSize.sm,
  },

  textPrice: {
    color: colors.secondary,
    fontSize: typography.fontSize.lg,
    fontWeight: typography.fontWeight.semibold,
  },

  // Input styles
  input: {
    backgroundColor: colors.background,
    borderRadius: borderRadius.md,
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    fontSize: typography.fontSize.md,
    color: colors.text.primary,
    borderWidth: 1,
    borderColor: colors.border,
    minHeight: touchTarget.minHeight,
  },

  // Divider styles
  divider: {
    height: 1,
    backgroundColor: colors.border,
    marginVertical: spacing.md,
  },

  // Spacing utilities
  marginSm: { margin: spacing.sm },
  marginMd: { margin: spacing.md },
  marginLg: { margin: spacing.lg },
  paddingSm: { padding: spacing.sm },
  paddingMd: { padding: spacing.md },
  paddingLg: { padding: spacing.lg },
});
