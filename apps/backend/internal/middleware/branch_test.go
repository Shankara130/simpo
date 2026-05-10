package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestGetBranchAccessInfo_SystemAdmin tests SYSTEM_ADMIN has all-branch access
// Story 1.6, AC4: SYSTEM_ADMIN role can access data from all branches
func TestGetBranchAccessInfo_SystemAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("GET", "/api/v1/products", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Set user context with SYSTEM_ADMIN role (no branch assigned)
	testUserCtx := &UserContext{
		UserID:   1,
		Username: "admin",
		Email:    "admin@simpo.com",
		Role:     RoleSystemAdmin,
		BranchID: nil,
	}
	c.Set(UserContextKey, testUserCtx)

	branchAccess := GetBranchAccessInfo(c)

	if branchAccess.UserRole != RoleSystemAdmin {
		t.Errorf("Expected role %s, got %s", RoleSystemAdmin, branchAccess.UserRole)
	}

	if !branchAccess.CanAccessAll {
		t.Error("Expected SYSTEM_ADMIN to have CanAccessAll = true")
	}

	if branchAccess.AssignedBranch != nil {
		t.Error("Expected SYSTEM_ADMIN to have nil AssignedBranch")
	}
}

// TestGetBranchAccessInfo_Owner tests OWNER has all-branch access
// Story 1.6, AC4: OWNER role can access data from all branches
func TestGetBranchAccessInfo_Owner(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("GET", "/api/v1/products", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Set user context with OWNER role (no branch assigned)
	testUserCtx := &UserContext{
		UserID:   2,
		Username: "owner",
		Email:    "owner@simpo.com",
		Role:     RoleOwner,
		BranchID: nil,
	}
	c.Set(UserContextKey, testUserCtx)

	branchAccess := GetBranchAccessInfo(c)

	if branchAccess.UserRole != RoleOwner {
		t.Errorf("Expected role %s, got %s", RoleOwner, branchAccess.UserRole)
	}

	if !branchAccess.CanAccessAll {
		t.Error("Expected OWNER to have CanAccessAll = true")
	}

	if branchAccess.AssignedBranch != nil {
		t.Error("Expected OWNER to have nil AssignedBranch")
	}
}

// TestGetBranchAccessInfo_Cashier tests CASHIER has assigned branch only
// Story 1.6, AC4: CASHIER role can only access data from assigned branch
func TestGetBranchAccessInfo_Cashier(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("GET", "/api/v1/products", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	testBranchID := uint(5)
	// Set user context with CASHIER role (branch assigned)
	testUserCtx := &UserContext{
		UserID:   3,
		Username: "cashier1",
		Email:    "cashier1@simpo.com",
		Role:     RoleCashier,
		BranchID: &testBranchID,
	}
	c.Set(UserContextKey, testUserCtx)

	branchAccess := GetBranchAccessInfo(c)

	if branchAccess.UserRole != RoleCashier {
		t.Errorf("Expected role %s, got %s", RoleCashier, branchAccess.UserRole)
	}

	if branchAccess.CanAccessAll {
		t.Error("Expected CASHIER to have CanAccessAll = false")
	}

	if branchAccess.AssignedBranch == nil {
		t.Error("Expected CASHIER to have non-nil AssignedBranch")
	} else if *branchAccess.AssignedBranch != testBranchID {
		t.Errorf("Expected AssignedBranch %d, got %d", testBranchID, *branchAccess.AssignedBranch)
	}
}

// TestGetBranchFilter_Admin tests no filter for SYSTEM_ADMIN
func TestGetBranchFilter_Admin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("GET", "/api/v1/products", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	testUserCtx := &UserContext{
		UserID:   1,
		Username: "admin",
		Email:    "admin@simpo.com",
		Role:     RoleSystemAdmin,
		BranchID: nil,
	}
	c.Set(UserContextKey, testUserCtx)

	branchFilter := GetBranchFilter(c)

	if branchFilter != nil {
		t.Error("Expected nil branch filter for SYSTEM_ADMIN (all branches access)")
	}
}

// TestGetBranchFilter_Owner tests no filter for OWNER
func TestGetBranchFilter_Owner(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("GET", "/api/v1/products", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	testUserCtx := &UserContext{
		UserID:   2,
		Username: "owner",
		Email:    "owner@simpo.com",
		Role:     RoleOwner,
		BranchID: nil,
	}
	c.Set(UserContextKey, testUserCtx)

	branchFilter := GetBranchFilter(c)

	if branchFilter != nil {
		t.Error("Expected nil branch filter for OWNER (all branches access)")
	}
}

// TestGetBranchFilter_Cashier tests branch filter for CASHIER
func TestGetBranchFilter_Cashier(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("GET", "/api/v1/products", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	testBranchID := uint(3)
	testUserCtx := &UserContext{
		UserID:   3,
		Username: "cashier1",
		Email:    "cashier1@simpo.com",
		Role:     RoleCashier,
		BranchID: &testBranchID,
	}
	c.Set(UserContextKey, testUserCtx)

	branchFilter := GetBranchFilter(c)

	if branchFilter == nil {
		t.Error("Expected non-nil branch filter for CASHIER")
	} else if *branchFilter != testBranchID {
		t.Errorf("Expected branch filter %d, got %d", testBranchID, *branchFilter)
	}
}

// TestValidateBranchAccess_Admin tests admin can access any branch
func TestValidateBranchAccess_Admin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("GET", "/api/v1/products", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	testUserCtx := &UserContext{
		UserID:   1,
		Username: "admin",
		Email:    "admin@simpo.com",
		Role:     RoleSystemAdmin,
		BranchID: nil,
	}
	c.Set(UserContextKey, testUserCtx)

	// Admin should be able to access any branch
	testBranches := []uint{1, 2, 3, 4, 5}
	for _, branchID := range testBranches {
		if !ValidateBranchAccess(c, branchID) {
			t.Errorf("Expected SYSTEM_ADMIN to have access to branch %d", branchID)
		}
	}
}

// TestValidateBranchAccess_Owner tests owner can access any branch
func TestValidateBranchAccess_Owner(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("GET", "/api/v1/products", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	testUserCtx := &UserContext{
		UserID:   2,
		Username: "owner",
		Email:    "owner@simpo.com",
		Role:     RoleOwner,
		BranchID: nil,
	}
	c.Set(UserContextKey, testUserCtx)

	// Owner should be able to access any branch
	testBranches := []uint{1, 2, 3, 4, 5}
	for _, branchID := range testBranches {
		if !ValidateBranchAccess(c, branchID) {
			t.Errorf("Expected OWNER to have access to branch %d", branchID)
		}
	}
}

// TestValidateBranchAccess_Cashier tests cashier can only access assigned branch
func TestValidateBranchAccess_Cashier(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("GET", "/api/v1/products", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	assignedBranchID := uint(5)
	testUserCtx := &UserContext{
		UserID:   3,
		Username: "cashier1",
		Email:    "cashier1@simpo.com",
		Role:     RoleCashier,
		BranchID: &assignedBranchID,
	}
	c.Set(UserContextKey, testUserCtx)

	// Cashier should only access assigned branch
	if !ValidateBranchAccess(c, assignedBranchID) {
		t.Errorf("Expected CASHIER to have access to assigned branch %d", assignedBranchID)
	}

	// Cashier should NOT access other branches
	otherBranches := []uint{1, 2, 3, 4, 6}
	for _, branchID := range otherBranches {
		if ValidateBranchAccess(c, branchID) {
			t.Errorf("Expected CASHIER to NOT have access to branch %d (only branch %d)", branchID, assignedBranchID)
		}
	}
}

// TestValidateBranchAccess_NoContext tests behavior when no user context exists
func TestValidateBranchAccess_NoContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("GET", "/api/v1/products", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// No user context set
	branchAccess := GetBranchAccessInfo(c)

	if branchAccess.UserRole != "" {
		t.Error("Expected empty role when no context")
	}

	if branchAccess.CanAccessAll {
		t.Error("Expected CanAccessAll = false when no context")
	}

	if branchAccess.AssignedBranch != nil {
		t.Error("Expected nil AssignedBranch when no context")
	}

	// Should not have access to any branch
	if ValidateBranchAccess(c, 1) {
		t.Error("Expected no branch access when no user context")
	}
}

// TestGetBranchAccessInfo_MultipleRoles tests branch access for all roles
// Story 1.6, AC4: Comprehensive branch access testing
func TestGetBranchAccessInfo_MultipleRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name           string
		role           string
		branchID       *uint
		expectedAll    bool
		expectedBranch *uint
	}{
		{
			name:         "SYSTEM_ADMIN - all branches",
			role:         RoleSystemAdmin,
			branchID:     nil,
			expectedAll:  true,
			expectedBranch: nil,
		},
		{
			name:         "OWNER - all branches",
			role:         RoleOwner,
			branchID:     nil,
			expectedAll:  true,
			expectedBranch: nil,
		},
		{
			name:           "CASHIER - assigned branch only",
			role:           RoleCashier,
			branchID:       func() *uint { i := uint(7); return &i }(),
			expectedAll:    false,
			expectedBranch: func() *uint { i := uint(7); return &i }(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/test", nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			testUserCtx := &UserContext{
				UserID:   1,
				Username: "test",
				Email:    "test@simpo.com",
				Role:     tc.role,
				BranchID: tc.branchID,
			}
			c.Set(UserContextKey, testUserCtx)

			branchAccess := GetBranchAccessInfo(c)

			if branchAccess.UserRole != tc.role {
				t.Errorf("Expected role %s, got %s", tc.role, branchAccess.UserRole)
			}

			if branchAccess.CanAccessAll != tc.expectedAll {
				t.Errorf("Expected CanAccessAll %v, got %v", tc.expectedAll, branchAccess.CanAccessAll)
			}

			if tc.expectedBranch == nil {
				if branchAccess.AssignedBranch != nil {
					t.Errorf("Expected nil AssignedBranch, got %d", *branchAccess.AssignedBranch)
				}
			} else {
				if branchAccess.AssignedBranch == nil {
					t.Error("Expected non-nil AssignedBranch, got nil")
				} else if *branchAccess.AssignedBranch != *tc.expectedBranch {
					t.Errorf("Expected AssignedBranch %d, got %d", *tc.expectedBranch, *branchAccess.AssignedBranch)
				}
			}
		})
	}
}
