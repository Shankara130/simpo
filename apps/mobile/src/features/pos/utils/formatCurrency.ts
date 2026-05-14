/**
 * Currency formatting utility for Indonesian Rupiah
 * Formats numbers as IDR with "Rp" prefix and dot separator
 */

/**
 * Format a number as Indonesian Rupiah currency string
 * @param amount - The amount to format (number or string)
 * @returns Formatted string like "Rp 150.000"
 */
export function formatCurrency(amount: number | string): string {
  // Convert to number if string input
  const numAmount = typeof amount === 'string' ? parseFloat(amount) : amount;

  // Handle invalid input
  if (isNaN(numAmount)) {
    return 'Rp 0';
  }

  // Round to nearest integer for IDR (no decimals)
  const roundedAmount = Math.round(numAmount);

  // Format with Indonesian locale
  const formatted = new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(roundedAmount);

  // Replace non-breaking space (\u00A0) with regular space for consistency
  return formatted.replace(/\u00A0/g, ' ');
}
