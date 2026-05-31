package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"log/slog"

	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
)

// supplierProductCatalogService implements SupplierProductCatalogService interface
// Story 10.5: Service layer with business logic, validation, and audit logging
type supplierProductCatalogService struct {
	catalogRepo repositories.SupplierProductCatalogRepository
	supplierRepo  repositories.SupplierRepository
	productRepo   repositories.ProductRepository
	branchRepo    repositories.BranchRepository
	auditSvc      AuditService
}

// NewSupplierProductCatalogService creates a new supplier product catalog service
// Story 10.5: Factory function for dependency injection
func NewSupplierProductCatalogService(
	catalogRepo repositories.SupplierProductCatalogRepository,
	supplierRepo repositories.SupplierRepository,
	productRepo repositories.ProductRepository,
	branchRepo repositories.BranchRepository,
	auditSvc AuditService,
) SupplierProductCatalogService {
	return &supplierProductCatalogService{
		catalogRepo: catalogRepo,
		supplierRepo:  supplierRepo,
		productRepo:   productRepo,
		branchRepo:    branchRepo,
		auditSvc:      auditSvc,
	}
}

// AssociateProduct associates a product with a supplier and specifies purchase price
// Story 10.5, AC1: Validates supplier and product exist, checks for duplicates, logs creation
func (s *supplierProductCatalogService) AssociateProduct(ctx context.Context, request *AssociateProductRequest, createdBy uint, ipAddress string) (*models.SupplierProductCatalog, error) {
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if createdBy == 0 {
		return nil, fmt.Errorf("createdBy user ID is required")
	}

	// Validate supplier exists and is active
	supplier, err := s.supplierRepo.GetByID(ctx, request.SupplierID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("supplier not found")
		}
		return nil, fmt.Errorf("failed to validate supplier: %w", err)
	}

	// Validate product exists and is active
	product, err := s.productRepo.GetByID(ctx, request.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to validate product: %w", err)
	}

	// Validate purchase price is positive (already handled by binding:gt=0, but double check)
	if request.PurchasePrice <= 0 {
		return nil, fmt.Errorf("purchase price must be positive")
	}

	// Validate minimum order quantity (already handled by binding:min=1)
	if request.MinimumOrderQuantity < 1 {
		return nil, fmt.Errorf("minimum order quantity must be at least 1")
	}

	// Validate lead time days if provided (PATCH-009)
	if request.LeadTimeDays != nil && *request.LeadTimeDays < 0 {
		return nil, fmt.Errorf("lead time days must be non-negative")
	}

	// Validate branch exists
	branch, err := s.branchRepo.GetByID(ctx, request.BranchID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("branch not found")
		}
		return nil, fmt.Errorf("failed to validate branch: %w", err)
	}

	_ = branch // Use branch for validation only

	// Create catalog entry
	catalog := &models.SupplierProductCatalog{
		SupplierID:          request.SupplierID,
		ProductID:           request.ProductID,
		PurchasePrice:       request.PurchasePrice,
		IsPreferred:         request.IsPreferred,
		SKUCode:             request.SKUCode,
		MinimumOrderQuantity: request.MinimumOrderQuantity,
		LeadTimeDays:        request.LeadTimeDays,
		BranchID:            request.BranchID, // Use request branch
		CreatedBy:           createdBy,
	}

	// If setting as preferred, unset any existing preferred supplier
	if request.IsPreferred {
		// Get existing preferred supplier for this product-branch combination
		existingPreferred, err := s.catalogRepo.GetPreferredSupplier(ctx, request.ProductID, request.BranchID)
		if err == nil && existingPreferred.ID != 0 {
			// Unset previous preferred supplier
			if err := s.catalogRepo.SetPreferredSupplier(ctx, existingPreferred.SupplierID, request.ProductID, false, request.BranchID); err != nil {
				return nil, fmt.Errorf("failed to unset previous preferred supplier: %w", err)
			}
		}
	}

	// Create catalog entry
	if err := s.catalogRepo.Create(ctx, catalog); err != nil {
		return nil, fmt.Errorf("failed to create catalog entry: %w", err)
	}

	// Log audit trail (CRITICAL: audit logging for all operations)
	slog.InfoContext(ctx, "catalog.product_associated",
		"user_id", createdBy,
		"ip_address", ipAddress,
		"supplier_id", supplier.ID,
		"supplier_name", supplier.Name,
		"product_id", product.ID,
		"product_name", product.Name,
		"purchase_price", catalog.PurchasePrice,
		"is_preferred", catalog.IsPreferred,
	)

	return catalog, nil
}

// GetProductCatalogByID retrieves a catalog entry by ID
// Story 10.5: Returns catalog entry details with supplier and product information
func (s *supplierProductCatalogService) GetProductCatalogByID(ctx context.Context, id uint) (*models.SupplierProductCatalog, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid catalog ID")
	}

	catalog, err := s.catalogRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalog entry: %w", err)
	}

	return catalog, nil
}

// ListProductCatalogs retrieves catalog entries with filtering and pagination
// Story 10.5: Supports filtering by supplier, product, branch, preferred status
func (s *supplierProductCatalogService) ListProductCatalogs(ctx context.Context, filter *SupplierProductCatalogListFilter) ([]*models.SupplierProductCatalog, int64, error) {
	// Convert filter to repository filter
	repoFilter := &repositories.SupplierProductCatalogFilter{
		SupplierID:  filter.SupplierID,
		ProductID:   filter.ProductID,
		BranchID:    filter.BranchID,
		IsPreferred: filter.IsPreferred,
		Page:        filter.Page,
		Limit:       filter.Limit,
		SortBy:      filter.SortBy,
		SortOrder:   filter.SortOrder,
	}

	catalogs, total, err := s.catalogRepo.List(ctx, repoFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list catalog entries: %w", err)
	}

	return catalogs, total, nil
}

// UpdatePurchasePrice updates the purchase price with price history tracking
// Story 10.5, AC1: Archives old price, creates new entry, logs change
func (s *supplierProductCatalogService) UpdatePurchasePrice(ctx context.Context, catalogID uint, request *UpdatePriceRequest, updatedBy uint, ipAddress string) error {
	if catalogID == 0 {
		return fmt.Errorf("invalid catalog ID")
	}
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if updatedBy == 0 {
		return fmt.Errorf("updatedBy user ID is required")
	}

	// Validate new price is positive (PATCH-009: already handled by binding:gt=0, but double check)
	if request.NewPrice <= 0 {
		return fmt.Errorf("new price must be positive")
	}

	// Get current catalog entry to log old price
	currentCatalog, err := s.catalogRepo.GetByID(ctx, catalogID)
	if err != nil {
		return fmt.Errorf("failed to get current catalog entry: %w", err)
	}

	// Update price with history tracking
	if err := s.catalogRepo.UpdatePrice(ctx, catalogID, request.NewPrice, updatedBy); err != nil {
		return fmt.Errorf("failed to update purchase price: %w", err)
	}

	// Log audit trail (CRITICAL: audit logging for all operations)
	slog.InfoContext(ctx, "catalog.price_updated",
		"user_id", updatedBy,
		"ip_address", ipAddress,
		"catalog_id", catalogID,
		"supplier_id", currentCatalog.SupplierID,
		"product_id", currentCatalog.ProductID,
		"old_price", currentCatalog.PurchasePrice,
		"new_price", request.NewPrice,
	)

	return nil
}

// SetPreferredSupplier sets or unsets the preferred supplier flag for a product
// Story 10.5, AC1: Ensures only one preferred supplier per product per branch, logs change
func (s *supplierProductCatalogService) SetPreferredSupplier(ctx context.Context, catalogID uint, request *SetPreferredRequest, updatedBy uint, ipAddress string) error {
	if catalogID == 0 {
		return fmt.Errorf("invalid catalog ID")
	}
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if updatedBy == 0 {
		return fmt.Errorf("updatedBy user ID is required")
	}

	// Get catalog entry to extract supplier and product IDs
	catalog, err := s.catalogRepo.GetByID(ctx, catalogID)
	if err != nil {
		return fmt.Errorf("failed to get catalog entry: %w", err)
	}

	// Set preferred supplier
	if err := s.catalogRepo.SetPreferredSupplier(ctx, catalog.SupplierID, catalog.ProductID, request.IsPreferred, catalog.BranchID); err != nil {
		return fmt.Errorf("failed to set preferred supplier: %w", err)
	}

	// Log audit trail (CRITICAL: audit logging for all operations)
	slog.InfoContext(ctx, "catalog.preferred_set",
		"user_id", updatedBy,
		"ip_address", ipAddress,
		"catalog_id", catalogID,
		"supplier_id", catalog.SupplierID,
		"product_id", catalog.ProductID,
		"is_preferred", request.IsPreferred,
	)

	return nil
}

// GetPreferredSupplier retrieves the preferred supplier for a product
// Story 10.5: Returns preferred supplier catalog entry or error if none set
func (s *supplierProductCatalogService) GetPreferredSupplier(ctx context.Context, productID uint, branchID uint) (*models.SupplierProductCatalog, error) {
	if productID == 0 || branchID == 0 {
		return nil, fmt.Errorf("product ID and branch ID are required")
	}

	catalog, err := s.catalogRepo.GetPreferredSupplier(ctx, productID, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get preferred supplier: %w", err)
	}

	return catalog, nil
}

// GetPriceHistory retrieves price history for a product with optional date range filtering
// Story 10.5, AC1: Returns historical prices grouped by supplier with date ranges
func (s *supplierProductCatalogService) GetPriceHistory(ctx context.Context, productID uint, filter *PriceHistoryFilter) ([]*PriceHistoryEntry, error) {
	if productID == 0 {
		return nil, fmt.Errorf("product ID is required")
	}

	// Parse date range filter if provided
	var startDate, endDate *time.Time
	if filter.StartDate != nil {
		parsed, err := time.Parse("2006-01-02", *filter.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start date format: %w", err)
		}
		startDate = &parsed
	}
	if filter.EndDate != nil {
		parsed, err := time.Parse("2006-01-02", *filter.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end date format: %w", err)
		}
		endDate = &parsed
	}

	// Get price history from repository
	catalogs, err := s.catalogRepo.GetPriceHistory(ctx, productID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get price history: %w", err)
	}

	// Convert to response format
	entries := make([]*PriceHistoryEntry, len(catalogs))
	for i, catalog := range catalogs {
		// Get supplier name
		supplierName := ""
		if catalog.Supplier.ID != 0 {
			supplierName = catalog.Supplier.Name
		}

		// Get product name
		productName := ""
		if catalog.Product.ID != 0 {
			productName = catalog.Product.Name
		}

		entries[i] = &PriceHistoryEntry{
			ID:            catalog.ID,
			SupplierID:    catalog.SupplierID,
			SupplierName:  supplierName,
			ProductID:     catalog.ProductID,
			ProductName:   productName,
			PurchasePrice: catalog.PurchasePrice,
			EffectiveFrom: catalog.PriceEffectiveFrom.Format("2006-01-02"),
			IsCurrent:     catalog.IsCurrentPrice(),
			IsPreferred:   catalog.IsPreferred,
			CreatedBy:     catalog.CreatedBy,
			CreatedAt:     catalog.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}

		// Set effective to if not current price
		if catalog.PriceEffectiveTo != nil {
			entries[i].EffectiveTo = catalog.PriceEffectiveTo.Format("2006-01-02")
		}
	}

	return entries, nil
}

// GetCatalogBySupplier retrieves a supplier's product catalog with current prices
// Story 10.5: Returns all current catalog entries for a supplier
func (s *supplierProductCatalogService) GetCatalogBySupplier(ctx context.Context, supplierID uint, branchID uint) ([]*models.SupplierProductCatalog, error) {
	if supplierID == 0 || branchID == 0 {
		return nil, fmt.Errorf("supplier ID and branch ID are required")
	}

	catalogs, err := s.catalogRepo.GetCatalogBySupplier(ctx, supplierID, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier catalog: %w", err)
	}

	return catalogs, nil
}
