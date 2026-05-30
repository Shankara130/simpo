package services

import (
	"fmt"
)

// Domain error types for service layer
// AC4: Services return domain errors (not repository errors)
// These errors can be converted to RFC 7807 format by handlers

// InsufficientStockError represents an error when stock is insufficient
type InsufficientStockError struct {
	ProductID    uint
	ProductName  string
	RequestedQty int64
	AvailableQty int64
}

func (e *InsufficientStockError) Error() string {
	return fmt.Sprintf("insufficient stock for product '%s' (ID: %d): requested %d, available %d",
		e.ProductName, e.ProductID, e.RequestedQty, e.AvailableQty)
}

// ProductExpiredError represents an error when product is expired
type ProductExpiredError struct {
	ProductID   uint
	ProductName string
	ProductSKU  string
	ExpiryDate  string
}

// ErrProductExpired is an alias for ProductExpiredError (Story 4.6, Task 3.3)
type ErrProductExpired = ProductExpiredError

func (e *ProductExpiredError) Error() string {
	return fmt.Sprintf("product '%s' (ID: %d) is expired since %s",
		e.ProductName, e.ProductID, e.ExpiryDate)
}

// InvalidInputError represents an error when input validation fails
type InvalidInputError struct {
	Field   string
	Message string
}

func (e *InvalidInputError) Error() string {
	return fmt.Sprintf("invalid input for field '%s': %s", e.Field, e.Message)
}

// ProductNotFoundError represents an error when product is not found
type ProductNotFoundError struct {
	ProductID uint
	SKU       string
}

func (e *ProductNotFoundError) Error() string {
	if e.SKU != "" {
		return fmt.Sprintf("product with SKU '%s' not found", e.SKU)
	}
	return fmt.Sprintf("product with ID %d not found", e.ProductID)
}

// UserNotFoundError represents an error when user is not found
type UserNotFoundError struct {
	UserID   uint
	Username string
}

func (e *UserNotFoundError) Error() string {
	if e.Username != "" {
		return fmt.Sprintf("user with username '%s' not found", e.Username)
	}
	return fmt.Sprintf("user with ID %d not found", e.UserID)
}

// DuplicateSKUError represents an error when SKU already exists
type DuplicateSKUError struct {
	SKU      string
	BranchID uint
}

func (e *DuplicateSKUError) Error() string {
	return fmt.Sprintf("product with SKU '%s' already exists in branch %d", e.SKU, e.BranchID)
}

// DuplicateUsernameError represents an error when username already exists
type DuplicateUsernameError struct {
	Username string
}

func (e *DuplicateUsernameError) Error() string {
	return fmt.Sprintf("user with username '%s' already exists", e.Username)
}

// UnauthorizedError represents an error when action is not authorized
type UnauthorizedError struct {
	Action string
	Reason string
}

func (e *UnauthorizedError) Error() string {
	return fmt.Sprintf("unauthorized action '%s': %s", e.Action, e.Reason)
}

// TransactionError represents a general transaction error
type TransactionError struct {
	Message string
	Details string
}

func (e *TransactionError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("transaction error: %s - %s", e.Message, e.Details)
	}
	return fmt.Sprintf("transaction error: %s", e.Message)
}

// ServiceError is a wrapper for service-level errors
type ServiceError struct {
	Op  string // Operation that failed
	Err error  // Underlying error
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}

// DuplicateInvoiceError represents an error when invoice number already exists
// Story 10.2: Invoice number must be unique across all invoices
type DuplicateInvoiceError struct {
	InvoiceNumber string
}

func (e *DuplicateInvoiceError) Error() string {
	return fmt.Sprintf("purchase invoice with number '%s' already exists", e.InvoiceNumber)
}

// InvoiceNotFoundError represents an error when purchase invoice is not found
// Story 10.2: Invoice lookup by ID returns this error
type InvoiceNotFoundError struct {
	ID uint
}

func (e *InvoiceNotFoundError) Error() string {
	return fmt.Sprintf("purchase invoice with ID %d not found", e.ID)
}
