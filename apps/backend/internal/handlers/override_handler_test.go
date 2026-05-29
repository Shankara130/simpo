package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
)

// MockProductRepoForOverride is a mock for product repository
type MockProductRepoForOverride struct {
	mock.Mock
}

func (m *MockProductRepoForOverride) GetStockForProduct(ctx context.Context, productID uint) (int64, error) {
	args := m.Called(ctx, productID)
	return args.Get(0).(int64), args.Error(1)
}

// MockConflictAuditForOverride is a mock for conflict audit service
type MockConflictAuditForOverride struct {
	mock.Mock
}

func (m *MockConflictAuditForOverride) LogConflictResolution(ctx context.Context, log services.ConflictResolutionLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

// TestOverrideHandler_SuccessfulOverride tests successful manual override
func TestOverrideHandler_SuccessfulOverride(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockProductRepoForOverride)
	mockAudit := new(MockConflictAuditForOverride)
	mockConflictSvc := services.NewConflictResolutionService(mockRepo, mockAudit)

	handler := NewOverrideHandler(mockConflictSvc)
	router := gin.New()
	router.POST("/api/v1/override/transaction", handler.OverrideTransaction)

	// Prepare override request
	overrideReq := map[string]interface{}{
		"transaction_id":   "TRX-001",
		"admin_user_id":    1,
		"admin_username":  "admin",
		"reason":          "Override for urgent customer need",
		"force_processing": true,
	}

	body, _ := json.Marshal(overrideReq)
	req, _ := http.NewRequest("POST", "/api/v1/override/transaction", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should succeed
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
}

// TestOverrideHandler_MissingAdminAuth tests missing admin authorization
func TestOverrideHandler_MissingAdminAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockProductRepoForOverride)
	mockAudit := new(MockConflictAuditForOverride)
	mockConflictSvc := services.NewConflictResolutionService(mockRepo, mockAudit)

	handler := NewOverrideHandler(mockConflictSvc)
	router := gin.New()
	router.POST("/api/v1/override/transaction", handler.OverrideTransaction)

	// Prepare override request WITH admin_user_id = 0 (no authorization)
	overrideReq := map[string]interface{}{
		"transaction_id":   "TRX-001",
		"admin_user_id":    0, // No admin authorization
		"admin_username":  "user",
		"reason":          "Test override",
		"force_processing": true,
	}

	body, _ := json.Marshal(overrideReq)
	req, _ := http.NewRequest("POST", "/api/v1/override/transaction", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Debug: print response
	t.Logf("Response status: %d", w.Code)
	t.Logf("Response body: %s", w.Body.String())

	// Should fail with 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["message"].(string), "admin authorization")
}
