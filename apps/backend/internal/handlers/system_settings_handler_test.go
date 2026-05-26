package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// MockSystemService is a mock implementation of SystemService for testing
type MockSystemService struct {
	getSettingsFunc     func(ctx context.Context) (map[string]string, error)
	getPharmacySettingsFunc func(ctx context.Context) (*models.PharmacySettings, error)
	updateSettingsFunc  func(ctx context.Context, settings *models.PharmacySettings, adminID uint, adminUsername, ipAddress string) error
	getPublicSettingsFunc func(ctx context.Context) (*models.PublicSettings, error)
	getBusinessNameFunc func(ctx context.Context) (string, error)
	getBusinessAddressFunc func(ctx context.Context) (string, error)
	getBusinessPhoneFunc func(ctx context.Context) (string, error)
	getBusinessEmailFunc func(ctx context.Context) (string, error)
}

func (m *MockSystemService) GetSettings(ctx context.Context) (map[string]string, error) {
	if m.getSettingsFunc != nil {
		return m.getSettingsFunc(ctx)
	}
	return map[string]string{}, nil
}

func (m *MockSystemService) GetPharmacySettings(ctx context.Context) (*models.PharmacySettings, error) {
	if m.getPharmacySettingsFunc != nil {
		return m.getPharmacySettingsFunc(ctx)
	}
	return &models.PharmacySettings{
		Name:   "Simpo Pharmacy",
		Address: "123 Main St",
		Phone:   "+62-21-1234-5678",
		Email:   "admin@simpo.pharmacy",
		LogoURL: "",
	}, nil
}

func (m *MockSystemService) UpdateSettings(ctx context.Context, settings *models.PharmacySettings, adminID uint, adminUsername, ipAddress string) error {
	if m.updateSettingsFunc != nil {
		return m.updateSettingsFunc(ctx, settings, adminID, adminUsername, ipAddress)
	}
	return nil
}

func (m *MockSystemService) GetPublicSettings(ctx context.Context) (*models.PublicSettings, error) {
	if m.getPublicSettingsFunc != nil {
		return m.getPublicSettingsFunc(ctx)
	}
	return &models.PublicSettings{
		BusinessName: "Simpo Pharmacy",
		Address:      "123 Main St",
		Phone:        "+62-21-1234-5678",
		Email:        "admin@simpo.pharmacy",
	}, nil
}

func (m *MockSystemService) GetBusinessName(ctx context.Context) (string, error) {
	if m.getBusinessNameFunc != nil {
		return m.getBusinessNameFunc(ctx)
	}
	return "Simpo Pharmacy", nil
}

func (m *MockSystemService) GetBusinessAddress(ctx context.Context) (string, error) {
	if m.getBusinessAddressFunc != nil {
		return m.getBusinessAddressFunc(ctx)
	}
	return "123 Main St", nil
}

func (m *MockSystemService) GetBusinessPhone(ctx context.Context) (string, error) {
	if m.getBusinessPhoneFunc != nil {
		return m.getBusinessPhoneFunc(ctx)
	}
	return "+62-21-1234-5678", nil
}

func (m *MockSystemService) GetBusinessEmail(ctx context.Context) (string, error) {
	if m.getBusinessEmailFunc != nil {
		return m.getBusinessEmailFunc(ctx)
	}
	return "admin@simpo.pharmacy", nil
}

// setupSettingsTestRouter creates a test router with the handler and context
func setupSettingsTestRouter(handler SystemSettingsHandler, userRole string, userID uint, username string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Middleware to set user context
	router.Use(func(c *gin.Context) {
		c.Set("user_role", userRole)
		c.Set("user_id", userID)
		c.Set("username", username)
		c.Set("branch_id", uint(1))
		c.Next()
	})

	router.GET("/api/v1/settings", handler.GetSettings)
	router.PUT("/api/v1/settings", handler.UpdateSettings)
	router.GET("/api/v1/settings/public", handler.GetPublicSettings)

	return router
}

// TestSystemSettingsHandler_GetSettings_Success tests successful get settings
func TestSystemSettingsHandler_GetSettings_Success(t *testing.T) {
	mockService := &MockSystemService{
		getPharmacySettingsFunc: func(ctx context.Context) (*models.PharmacySettings, error) {
			return &models.PharmacySettings{
				Name:    "Test Pharmacy",
				Address: "456 Test St",
				Phone:   "+62-21-9876-5432",
				Email:   "test@example.com",
				LogoURL: "https://example.com/logo.png",
			}, nil
		},
	}

	handler := NewSystemSettingsHandler(mockService)
	router := setupSettingsTestRouter(handler, user.RoleSystemAdmin, 1, "admin")

	req, _ := http.NewRequest("GET", "/api/v1/settings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "Test Pharmacy", data["businessName"])
	assert.Equal(t, "456 Test St", data["address"])
	assert.Equal(t, "+62-21-9876-5432", data["phone"])
	assert.Equal(t, "test@example.com", data["email"])
	assert.Equal(t, "https://example.com/logo.png", data["logoUrl"])
}

// TestSystemSettingsHandler_GetSettings_RBAC_NotSystemAdmin tests RBAC enforcement
func TestSystemSettingsHandler_GetSettings_RBAC_NotSystemAdmin(t *testing.T) {
	mockService := &MockSystemService{}
	handler := NewSystemSettingsHandler(mockService)
	router := setupSettingsTestRouter(handler, user.RoleCashier, 2, "cashier1")

	req, _ := http.NewRequest("GET", "/api/v1/settings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusOK, w.Code, "Should not return 200 OK for non-admin")
	assert.Contains(t, []int{http.StatusForbidden, http.StatusUnauthorized}, w.Code)
}

// TestSystemSettingsHandler_GetSettings_RBAC_NoRole tests missing user role
func TestSystemSettingsHandler_GetSettings_RBAC_NoRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := &MockSystemService{}
	handler := NewSystemSettingsHandler(mockService)
	router.GET("/api/v1/settings", handler.GetSettings)

	req, _ := http.NewRequest("GET", "/api/v1/settings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusOK, w.Code, "Should not return 200 OK without user role")
	assert.Contains(t, []int{http.StatusForbidden, http.StatusUnauthorized}, w.Code)
}

// TestSystemSettingsHandler_UpdateSettings_Success tests successful update
func TestSystemSettingsHandler_UpdateSettings_Success(t *testing.T) {
	mockService := &MockSystemService{
		updateSettingsFunc: func(ctx context.Context, settings *models.PharmacySettings, adminID uint, adminUsername, ipAddress string) error {
			assert.Equal(t, "Updated Pharmacy", settings.Name)
			assert.Equal(t, "789 Updated St", settings.Address)
			assert.Equal(t, "+62-21-5555-5555", settings.Phone)
			assert.Equal(t, "updated@example.com", settings.Email)
			assert.Equal(t, uint(1), adminID)
			assert.Equal(t, "admin", adminUsername)
			return nil
		},
	}

	handler := NewSystemSettingsHandler(mockService)
	router := setupSettingsTestRouter(handler, user.RoleSystemAdmin, 1, "admin")

	reqBody := `{
		"businessName": "Updated Pharmacy",
		"address": "789 Updated St",
		"phone": "+62-21-5555-5555",
		"email": "updated@example.com"
	}`
	req, _ := http.NewRequest("PUT", "/api/v1/settings", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "Settings updated successfully", data["message"])
	assert.Equal(t, "admin", data["updatedBy"])
}

// TestSystemSettingsHandler_UpdateSettings_ValidationError tests validation errors
func TestSystemSettingsHandler_UpdateSettings_ValidationError(t *testing.T) {
	mockService := &MockSystemService{}
	handler := NewSystemSettingsHandler(mockService)
	router := setupSettingsTestRouter(handler, user.RoleSystemAdmin, 1, "admin")

	// Missing required fields
	reqBody := `{"address": "789 Updated St"}`
	req, _ := http.NewRequest("PUT", "/api/v1/settings", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusOK, w.Code, "Should not return 200 OK for validation error")
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusUnprocessableEntity}, w.Code)
}

// TestSystemSettingsHandler_UpdateSettings_InvalidEmail tests email validation
func TestSystemSettingsHandler_UpdateSettings_InvalidEmail(t *testing.T) {
	mockService := &MockSystemService{}
	handler := NewSystemSettingsHandler(mockService)
	router := setupSettingsTestRouter(handler, user.RoleSystemAdmin, 1, "admin")

	reqBody := `{
		"businessName": "Test Pharmacy",
		"email": "not-a-valid-email"
	}`
	req, _ := http.NewRequest("PUT", "/api/v1/settings", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusOK, w.Code, "Should not return 200 OK for invalid email")
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusUnprocessableEntity}, w.Code)
}

// TestSystemSettingsHandler_UpdateSettings_RBAC_NotSystemAdmin tests RBAC enforcement
func TestSystemSettingsHandler_UpdateSettings_RBAC_NotSystemAdmin(t *testing.T) {
	mockService := &MockSystemService{}
	handler := NewSystemSettingsHandler(mockService)
	router := setupSettingsTestRouter(handler, user.RoleCashier, 2, "cashier1")

	reqBody := `{
		"businessName": "Test Pharmacy",
		"email": "test@example.com"
	}`
	req, _ := http.NewRequest("PUT", "/api/v1/settings", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusOK, w.Code, "Should not return 200 OK for non-admin")
	assert.Contains(t, []int{http.StatusForbidden, http.StatusUnauthorized}, w.Code)
}

// TestSystemSettingsHandler_UpdateSettings_ServiceError tests service error handling
func TestSystemSettingsHandler_UpdateSettings_ServiceError(t *testing.T) {
	mockService := &MockSystemService{
		updateSettingsFunc: func(ctx context.Context, settings *models.PharmacySettings, adminID uint, adminUsername, ipAddress string) error {
			return errors.New("database error")
		},
	}

	handler := NewSystemSettingsHandler(mockService)
	router := setupSettingsTestRouter(handler, user.RoleSystemAdmin, 1, "admin")

	reqBody := `{
		"businessName": "Test Pharmacy",
		"email": "test@example.com"
	}`
	req, _ := http.NewRequest("PUT", "/api/v1/settings", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusOK, w.Code, "Should not return 200 OK on service error")
	assert.Contains(t, []int{http.StatusInternalServerError}, w.Code)
}

// TestSystemSettingsHandler_GetPublicSettings_Success tests public settings endpoint
func TestSystemSettingsHandler_GetPublicSettings_Success(t *testing.T) {
	mockService := &MockSystemService{
		getPublicSettingsFunc: func(ctx context.Context) (*models.PublicSettings, error) {
			return &models.PublicSettings{
				BusinessName: "Public Pharmacy",
				Address:      "321 Public St",
				Phone:        "+62-21-1111-2222",
				Email:        "public@example.com",
			}, nil
		},
	}

	handler := NewSystemSettingsHandler(mockService)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/settings/public", handler.GetPublicSettings)

	req, _ := http.NewRequest("GET", "/api/v1/settings/public", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "Public Pharmacy", data["businessName"])
	assert.Equal(t, "321 Public St", data["address"])
	assert.Equal(t, "+62-21-1111-2222", data["phone"])
	assert.Equal(t, "public@example.com", data["email"])
}

// TestSystemSettingsHandler_GetPublicSettings_NoAuthentication tests no auth required
func TestSystemSettingsHandler_GetPublicSettings_NoAuthentication(t *testing.T) {
	mockService := &MockSystemService{}
	handler := NewSystemSettingsHandler(mockService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	// No middleware - should work without authentication
	router.GET("/api/v1/settings/public", handler.GetPublicSettings)

	req, _ := http.NewRequest("GET", "/api/v1/settings/public", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Should return 200 OK without authentication")
}

// TestSystemSettingsHandler_UpdateSettings_OptionalFields tests optional fields
func TestSystemSettingsHandler_UpdateSettings_OptionalFields(t *testing.T) {
	mockService := &MockSystemService{
		updateSettingsFunc: func(ctx context.Context, settings *models.PharmacySettings, adminID uint, adminUsername, ipAddress string) error {
			assert.Equal(t, "Test Pharmacy", settings.Name)
			assert.Empty(t, settings.Address, "Address should be empty")
			assert.Empty(t, settings.Phone, "Phone should be empty")
			assert.Equal(t, "test@example.com", settings.Email)
			return nil
		},
	}

	handler := NewSystemSettingsHandler(mockService)
	router := setupSettingsTestRouter(handler, user.RoleSystemAdmin, 1, "admin")

	// Only required fields
	reqBody := `{
		"businessName": "Test Pharmacy",
		"email": "test@example.com"
	}`
	req, _ := http.NewRequest("PUT", "/api/v1/settings", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestSystemSettingsHandler_GetSettings_EmptyFields tests empty optional fields
func TestSystemSettingsHandler_GetSettings_EmptyFields(t *testing.T) {
	mockService := &MockSystemService{
		getPharmacySettingsFunc: func(ctx context.Context) (*models.PharmacySettings, error) {
			return &models.PharmacySettings{
				Name:    "Minimal Pharmacy",
				Address: "",
				Phone:   "",
				Email:   "minimal@example.com",
				LogoURL: "",
			}, nil
		},
	}

	handler := NewSystemSettingsHandler(mockService)
	router := setupSettingsTestRouter(handler, user.RoleSystemAdmin, 1, "admin")

	req, _ := http.NewRequest("GET", "/api/v1/settings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "Minimal Pharmacy", data["businessName"])
	assert.Equal(t, "", data["address"])
	assert.Equal(t, "", data["phone"])
	assert.Equal(t, "minimal@example.com", data["email"])
}
