package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/permissions"
)

// BranchAccessInfo contains branch access information for the current user
// Story 1.6, AC4: Branch-Level Data Isolation
type BranchAccessInfo struct {
	UserRole        string  // User's role (SYSTEM_ADMIN, OWNER, CASHIER)
	AssignedBranch  *uint   // Branch ID assigned to user (nil for admin/owner)
	CanAccessAll    bool    // Whether user can access all branches
}

// GetBranchAccessInfo extracts branch access information from request context
// Story 1.6, AC4: CASHIER can only access assigned branch, OWNER/SYSTEM_ADMIN can access all branches
//
// Usage:
//   branchAccess := GetBranchAccessInfo(c)
//   if !branchAccess.CanAccessAll {
//       // Filter by branchAccess.AssignedBranch
//   }
func GetBranchAccessInfo(c *gin.Context) *BranchAccessInfo {
	userRole := GetUserRole(c)
	if userRole == "" {
		return &BranchAccessInfo{
			UserRole:       "",
			AssignedBranch: nil,
			CanAccessAll:   false,
		}
	}

	branchID := GetBranchID(c)

	// Check if role can access all branches
	// Story 1.6, AC4: OWNER and SYSTEM_ADMIN can access all branches
	canAccessAll := permissions.CanAccessAllBranches(userRole)

	return &BranchAccessInfo{
		UserRole:       userRole,
		AssignedBranch: branchID,
		CanAccessAll:   canAccessAll,
	}
}

// GetBranchFilter returns the branch ID to use for filtering repository queries
// Returns nil if user can access all branches (no filter needed)
// Returns specific branch ID if user can only access assigned branch
//
// Usage in repository:
//   branchFilter := GetBranchFilter(c)
//   query := db.Where("branch_id = ?", *branchFilter) // if branchFilter != nil
//   query := db // if branchFilter == nil (no filter)
func GetBranchFilter(c *gin.Context) *uint {
	branchAccess := GetBranchAccessInfo(c)
	if branchAccess.CanAccessAll {
		// User can access all branches - no filter needed
		return nil
	}
	// User can only access assigned branch
	return branchAccess.AssignedBranch
}

// ValidateBranchAccess checks if a user can access a specific branch
// Returns true if user can access the branch, false otherwise
//
// Usage:
//   if !ValidateBranchAccess(c, requestedBranchID) {
//       return 403 Forbidden
//   }
func ValidateBranchAccess(c *gin.Context, requestedBranchID uint) bool {
	branchAccess := GetBranchAccessInfo(c)

	// Admin and Owner can access all branches
	if branchAccess.CanAccessAll {
		return true
	}

	// Cashier: only assigned branch
	if branchAccess.AssignedBranch != nil && *branchAccess.AssignedBranch == requestedBranchID {
		return true
	}

	return false
}
