package permissions

import (
	"testing"
)

// TestGetRolePermissions_SystemAdmin tests SYSTEM_ADMIN role configuration
// Story 1.6, AC7: SYSTEM_ADMIN has all permissions
func TestGetRolePermissions_SystemAdmin(t *testing.T) {
	perms := GetRolePermissions(RoleSystemAdmin)

	if perms.Role != RoleSystemAdmin {
		t.Errorf("Expected role %s, got %s", RoleSystemAdmin, perms.Role)
	}

	// Check: Should have all permissions
	expectedPerms := 4
	if len(perms.Permissions) != expectedPerms {
		t.Errorf("Expected %d permissions, got %d", expectedPerms, len(perms.Permissions))
	}

	// Check: Should have wildcard endpoint access
	if len(perms.AllowedEndpoints) != 1 || perms.AllowedEndpoints[0] != "*" {
		t.Errorf("Expected wildcard endpoint access, got %v", perms.AllowedEndpoints)
	}

	// Check: Should have all-branch access
	if !perms.AllBranchesAccess {
		t.Error("Expected AllBranchesAccess to be true for SYSTEM_ADMIN")
	}
}

// TestGetRolePermissions_Owner tests OWNER role configuration
// Story 1.6, AC7: OWNER has Reports, Inventory, Users (full branch access)
func TestGetRolePermissions_Owner(t *testing.T) {
	perms := GetRolePermissions(RoleOwner)

	if perms.Role != RoleOwner {
		t.Errorf("Expected role %s, got %s", RoleOwner, perms.Role)
	}

	// Check: Should have READ and WRITE permissions
	if len(perms.Permissions) != 2 {
		t.Errorf("Expected 2 permissions for OWNER, got %d", len(perms.Permissions))
	}

	// Check: Should have specific endpoint access
	expectedEndpoints := 6
	if len(perms.AllowedEndpoints) != expectedEndpoints {
		t.Errorf("Expected %d allowed endpoints for OWNER, got %d", expectedEndpoints, len(perms.AllowedEndpoints))
	}

	// Check: Should include reports endpoint
	hasReports := false
	for _, endpoint := range perms.AllowedEndpoints {
		if endpoint == "/api/v1/reports" {
			hasReports = true
			break
		}
	}
	if !hasReports {
		t.Error("Expected OWNER to have access to /api/v1/reports endpoint")
	}

	// Check: Should have all-branch access
	if !perms.AllBranchesAccess {
		t.Error("Expected AllBranchesAccess to be true for OWNER")
	}
}

// TestGetRolePermissions_Cashier tests CASHIER role configuration
// Story 1.6, AC7: CASHIER has POS only (assigned branch only)
func TestGetRolePermissions_Cashier(t *testing.T) {
	perms := GetRolePermissions(RoleCashier)

	if perms.Role != RoleCashier {
		t.Errorf("Expected role %s, got %s", RoleCashier, perms.Role)
	}

	// Check: Should have READ and WRITE permissions
	if len(perms.Permissions) != 2 {
		t.Errorf("Expected 2 permissions for CASHIER, got %d", len(perms.Permissions))
	}

	// Check: Should have POS endpoints only
	expectedEndpoints := 2
	if len(perms.AllowedEndpoints) != expectedEndpoints {
		t.Errorf("Expected %d allowed endpoints for CASHIER, got %d", expectedEndpoints, len(perms.AllowedEndpoints))
	}

	// Check: Should include transactions endpoint
	hasTransactions := false
	for _, endpoint := range perms.AllowedEndpoints {
		if endpoint == "/api/v1/transactions" {
			hasTransactions = true
			break
		}
	}
	if !hasTransactions {
		t.Error("Expected CASHIER to have access to /api/v1/transactions endpoint")
	}

	// Check: Should NOT have all-branch access
	if perms.AllBranchesAccess {
		t.Error("Expected AllBranchesAccess to be false for CASHIER")
	}
}

// TestGetRolePermissions_UnknownRole tests default deny for unknown roles
func TestGetRolePermissions_UnknownRole(t *testing.T) {
	perms := GetRolePermissions("unknown_role")

	if perms.Role != "unknown_role" {
		t.Errorf("Expected role unknown_role, got %s", perms.Role)
	}

	// Check: Should have no permissions
	if len(perms.Permissions) != 0 {
		t.Errorf("Expected 0 permissions for unknown role, got %d", len(perms.Permissions))
	}

	// Check: Should have no endpoint access
	if len(perms.AllowedEndpoints) != 0 {
		t.Errorf("Expected 0 allowed endpoints for unknown role, got %d", len(perms.AllowedEndpoints))
	}

	// Check: Should NOT have all-branch access
	if perms.AllBranchesAccess {
		t.Error("Expected AllBranchesAccess to be false for unknown role")
	}
}

// TestHasPermission_AdminPermission tests ADMIN permission grants all access
func TestHasPermission_AdminPermission(t *testing.T) {
	// SYSTEM_ADMIN has ADMIN permission
	if !HasPermission(RoleSystemAdmin, PermAdmin) {
		t.Error("Expected SYSTEM_ADMIN to have ADMIN permission")
	}

	// ADMIN permission should grant all other permissions
	if !HasPermission(RoleSystemAdmin, PermRead) {
		t.Error("Expected ADMIN permission to grant READ access")
	}
	if !HasPermission(RoleSystemAdmin, PermWrite) {
		t.Error("Expected ADMIN permission to grant WRITE access")
	}
	if !HasPermission(RoleSystemAdmin, PermDelete) {
		t.Error("Expected ADMIN permission to grant DELETE access")
	}
}

// TestHasPermission_OwnerPermissions tests OWNER role permissions
func TestHasPermission_OwnerPermissions(t *testing.T) {
	// OWNER should have READ permission
	if !HasPermission(RoleOwner, PermRead) {
		t.Error("Expected OWNER to have READ permission")
	}

	// OWNER should have WRITE permission
	if !HasPermission(RoleOwner, PermWrite) {
		t.Error("Expected OWNER to have WRITE permission")
	}

	// OWNER should NOT have DELETE permission
	if HasPermission(RoleOwner, PermDelete) {
		t.Error("Expected OWNER to NOT have DELETE permission")
	}

	// OWNER should NOT have ADMIN permission
	if HasPermission(RoleOwner, PermAdmin) {
		t.Error("Expected OWNER to NOT have ADMIN permission")
	}
}

// TestHasPermission_CashierPermissions tests CASHIER role permissions
func TestHasPermission_CashierPermissions(t *testing.T) {
	// CASHIER should have READ permission
	if !HasPermission(RoleCashier, PermRead) {
		t.Error("Expected CASHIER to have READ permission")
	}

	// CASHIER should have WRITE permission
	if !HasPermission(RoleCashier, PermWrite) {
		t.Error("Expected CASHIER to have WRITE permission")
	}

	// CASHIER should NOT have DELETE permission
	if HasPermission(RoleCashier, PermDelete) {
		t.Error("Expected CASHIER to NOT have DELETE permission")
	}

	// CASHIER should NOT have ADMIN permission
	if HasPermission(RoleCashier, PermAdmin) {
		t.Error("Expected CASHIER to NOT have ADMIN permission")
	}
}

// TestCanAccessEndpoint_SystemAdmin tests SYSTEM_ADMIN can access all endpoints
// Story 1.6, AC3: SYSTEM_ADMIN role has access to all endpoints
func TestCanAccessEndpoint_SystemAdmin(t *testing.T) {
	testCases := []struct {
		endpoint  string
		allowed   bool
	}{
		{"/api/v1/users", true},
		{"/api/v1/products", true},
		{"/api/v1/reports/daily", true},
		{"/api/v1/transactions", true},
		{"/api/v1/admin/settings", true},
		{"/any/random/endpoint", true},
	}

	for _, tc := range testCases {
		result := CanAccessEndpoint(RoleSystemAdmin, tc.endpoint)
		if result != tc.allowed {
			t.Errorf("Expected endpoint %s access to be %v for SYSTEM_ADMIN, got %v", tc.endpoint, tc.allowed, result)
		}
	}
}

// TestCanAccessEndpoint_Owner tests OWNER endpoint access
// Story 1.6, AC3: OWNER has access to business oversight endpoints
func TestCanAccessEndpoint_Owner(t *testing.T) {
	testCases := []struct {
		endpoint  string
		allowed   bool
	}{
		{"/api/v1/users", true},
		{"/api/v1/users/1", true},          // Prefix match
		{"/api/v1/products", true},
		{"/api/v1/products/123", true},      // Prefix match
		{"/api/v1/transactions", true},
		{"/api/v1/reports", true},
		{"/api/v1/reports/daily", true},     // Prefix match
		{"/api/v1/inventory", true},
		{"/api/v1/branches", true},
		{"/api/v1/admin/settings", false},   // Not in allowed list
		{"/api/v1/auth/login", false},       // Public endpoint, but not in owner whitelist
	}

	for _, tc := range testCases {
		result := CanAccessEndpoint(RoleOwner, tc.endpoint)
		if result != tc.allowed {
			t.Errorf("Expected endpoint %s access to be %v for OWNER, got %v", tc.endpoint, tc.allowed, result)
		}
	}
}

// TestCanAccessEndpoint_Cashier tests CASHIER endpoint access
// Story 1.6, AC3: CASHIER has access to POS endpoints only
func TestCanAccessEndpoint_Cashier(t *testing.T) {
	testCases := []struct {
		endpoint  string
		allowed   bool
	}{
		{"/api/v1/transactions", true},
		{"/api/v1/transactions/123", true}, // Prefix match
		{"/api/v1/products", true},          // For stock checking
		{"/api/v1/products/123", true},      // Prefix match
		{"/api/v1/reports", false},           // Not allowed for cashier
		{"/api/v1/users", false},             // Not allowed for cashier
		{"/api/v1/inventory", false},         // Not allowed for cashier
		{"/api/v1/admin/settings", false},    // Not allowed for cashier
	}

	for _, tc := range testCases {
		result := CanAccessEndpoint(RoleCashier, tc.endpoint)
		if result != tc.allowed {
			t.Errorf("Expected endpoint %s access to be %v for CASHIER, got %v", tc.endpoint, tc.allowed, result)
		}
	}
}

// TestCanAccessAllBranches tests branch access permissions
// Story 1.6, AC4: Branch-level data isolation
func TestCanAccessAllBranches(t *testing.T) {
	testCases := []struct {
		role              string
		allBranchesAccess bool
	}{
		{RoleSystemAdmin, true},  // Admin: all branches
		{RoleOwner, true},         // Owner: all branches
		{RoleCashier, false},      // Cashier: assigned branch only
		{"unknown_role", false},         // Unknown: no access
	}

	for _, tc := range testCases {
		result := CanAccessAllBranches(tc.role)
		if result != tc.allBranchesAccess {
			t.Errorf("Expected all-branch access for role %s to be %v, got %v", tc.role, tc.allBranchesAccess, result)
		}
	}
}

// TestIsValidRole tests role validation
func TestIsValidRole(t *testing.T) {
	testCases := []struct {
		role   string
		valid  bool
	}{
		{RoleSystemAdmin, true},
		{RoleOwner, true},
		{RoleCashier, true},
		{RoleAdmin, true},        // Legacy GRAB role
		{RoleUser, true},         // Legacy GRAB role
		{"invalid_role", false},
		{"", false},
	}

	for _, tc := range testCases {
		result := IsValidRole(tc.role)
		if result != tc.valid {
			t.Errorf("Expected role %s validity to be %v, got %v", tc.role, tc.valid, result)
		}
	}
}

// TestOwnerCannotDelete tests that OWNER cannot delete resources
// This is a security-critical test to ensure OWNER has limited permissions
func TestOwnerCannotDelete(t *testing.T) {
	// Story 1.6, AC3: OWNER should not have DELETE permission
	if HasPermission(RoleOwner, PermDelete) {
		t.Error("Security: OWNER should NOT have DELETE permission")
	}
}

// TestCashierCannotAccessReports tests that CASHIER cannot access reports
// This is a security-critical test to ensure role separation
func TestCashierCannotAccessReports(t *testing.T) {
	// Story 1.6, AC3: CASHIER should not access reports
	if CanAccessEndpoint(RoleCashier, "/api/v1/reports/daily") {
		t.Error("Security: CASHIER should NOT access report endpoints")
	}

	if CanAccessEndpoint(RoleCashier, "/api/v1/reports") {
		t.Error("Security: CASHIER should NOT access report endpoints (prefix)")
	}
}

// TestCashierCannotAccessAllBranches tests branch isolation for CASHIER
// Story 1.6, AC4: CASHIER can only access assigned branch
func TestCashierCannotAccessAllBranches(t *testing.T) {
	if CanAccessAllBranches(RoleCashier) {
		t.Error("Security: CASHIER should NOT have all-branch access")
	}
}

// TestSystemAdminHasFullAccess tests that SYSTEM_ADMIN has full access
// Story 1.6, AC3: SYSTEM_ADMIN has access to all endpoints
func TestSystemAdminHasFullAccess(t *testing.T) {
	// Should have all permissions
	allPerms := []Permission{PermRead, PermWrite, PermDelete, PermAdmin}
	for _, perm := range allPerms {
		if !HasPermission(RoleSystemAdmin, perm) {
			t.Errorf("Expected SYSTEM_ADMIN to have %s permission", perm)
		}
	}

	// Should access all endpoints
	testEndpoints := []string{
		"/api/v1/users",
		"/api/v1/products",
		"/api/v1/reports",
		"/api/v1/transactions",
		"/api/v1/admin/settings",
		"/any/random/endpoint",
	}
	for _, endpoint := range testEndpoints {
		if !CanAccessEndpoint(RoleSystemAdmin, endpoint) {
			t.Errorf("Expected SYSTEM_ADMIN to access endpoint %s", endpoint)
		}
	}

	// Should have all-branch access
	if !CanAccessAllBranches(RoleSystemAdmin) {
		t.Error("Expected SYSTEM_ADMIN to have all-branch access")
	}
}
