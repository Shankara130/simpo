package repositories

// Repository is the main repository container
// AC5: Dependency injection pattern - all repositories accessible through this container
type Repository struct {
	Branch              BranchRepository
	Product             ProductRepository
	Transaction         TransactionRepository
	TransactionItem     TransactionItemRepository
	User                UserRepository
	Supplier            SupplierRepository
	PurchaseInvoice     PurchaseInvoiceRepository
	GoodsReceipt        GoodsReceiptRepository
	SupplierPayment     SupplierPaymentRepository
}

// NewRepositories creates a new repository container with all repositories
// AC5: Factory function for dependency injection
// Story 10.3: Added goodsReceiptRepo parameter
// Story 10.4: Added supplierPaymentRepo parameter
func NewRepositories(
	branchRepo BranchRepository,
	productRepo ProductRepository,
	transactionRepo TransactionRepository,
	transactionItemRepo TransactionItemRepository,
	userRepo UserRepository,
	supplierRepo SupplierRepository,
	purchaseInvoiceRepo PurchaseInvoiceRepository,
	goodsReceiptRepo GoodsReceiptRepository,
	supplierPaymentRepo SupplierPaymentRepository,
) *Repository {
	return &Repository{
		Branch:              branchRepo,
		Product:             productRepo,
		Transaction:         transactionRepo,
		TransactionItem:     transactionItemRepo,
		User:                userRepo,
		Supplier:            supplierRepo,
		PurchaseInvoice:     purchaseInvoiceRepo,
		GoodsReceipt:        goodsReceiptRepo,
		SupplierPayment:     supplierPaymentRepo,
	}
}

// This file provides factory functions for all repositories
// AC5: Each concrete implementation will have its own New*Repository function
// The implementations are in separate files:
// - branch_repository.go (implementation)
// - product_repository.go (implementation)
// - transaction_repository.go (implementation)
// - transaction_item_repository.go (implementation)
// - user_repository.go (implementation)
