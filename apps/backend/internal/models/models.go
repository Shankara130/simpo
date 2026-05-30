package models

// Models package exports all database models for the simpo application.
// These models use GORM for ORM operations and follow the architecture:
// - Branch: Pharmacy branch locations
// - Product: Inventory items with stock and pricing
// - Transaction: Sales transactions
// - TransactionItem: Line items within transactions
// - Supplier: Supplier master data for purchase management
// - PurchaseInvoice: Purchase invoices from suppliers
// - PurchaseInvoiceItem: Line items within purchase invoices
// - GoodsReceipt: Goods receipts from suppliers (Story 10.3)

// All models follow these conventions:
// - JSON serialization uses camelCase
// - Database columns use snake_case
// - Price fields are strings for decimal precision
// - Soft delete support with DeletedAt
// - Audit trail with CreatedBy, UpdatedBy, Version
