package whitelist

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	apiErrors "github.com/vahiiiid/go-rest-api-boilerplate/internal/errors"
)

// setupHandlerTest creates a test environment with in-memory database
func setupHandlerTest(t *testing.T) (*gorm.DB, *Handler, *gin.Engine) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Migrate tables
	err = db.AutoMigrate(&WhitelistEntry{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Create service and handler
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	// Create Gin router
	router := gin.New()
	router.Use(apiErrors.ErrorHandler())

	// Register routes (without RBAC middleware for testing)
	router.POST("/whitelist", handler.AddDomain)
	router.GET("/whitelist", handler.ListDomains)
	router.GET("/whitelist/:id", handler.GetDomain)
	router.PUT("/whitelist/:id", handler.UpdateDomain)
	router.DELETE("/whitelist/:id", handler.DeleteDomain)

	return db, handler, router
}

// TestHandlerAddDomain verifies POST /whitelist endpoint
func TestHandlerAddDomain(t *testing.T) {
	_, _, router := setupHandlerTest(t)

	reqBody := AddWhitelistEntryRequest{
		Domain:      "test.pharmacy",
		DefaultRole: "CASHIER",
		Description: "Test pharmacy domain",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/whitelist", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.True(t, response["success"].(bool))
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "test.pharmacy", data["domain"])
	assert.Equal(t, "CASHIER", data["default_role"])
}

// TestHandlerAddDomainInvalidRequest verifies validation errors
func TestHandlerAddDomainInvalidRequest(t *testing.T) {
	_, _, router := setupHandlerTest(t)

	testCases := []struct {
		name       string
		reqBody    AddWhitelistEntryRequest
		expectCode int
	}{
		{
			name: "Empty domain",
			reqBody: AddWhitelistEntryRequest{
				Domain:      "",
				DefaultRole: "CASHIER",
			},
			expectCode: http.StatusBadRequest,
		},
		{
			name: "Invalid role",
			reqBody: AddWhitelistEntryRequest{
				Domain:      "test.pharmacy",
				DefaultRole: "INVALID_ROLE",
			},
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.reqBody)
			req, _ := http.NewRequest("POST", "/whitelist", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectCode, w.Code)
		})
	}
}

// TestHandlerAddDomainDuplicate verifies duplicate domain rejection
func TestHandlerAddDomainDuplicate(t *testing.T) {
	_, _, router := setupHandlerTest(t)

	reqBody := AddWhitelistEntryRequest{
		Domain:      "test.pharmacy",
		DefaultRole: "CASHIER",
	}

	// First request
	body, _ := json.Marshal(reqBody)
	req1, _ := http.NewRequest("POST", "/whitelist", bytes.NewBuffer(body))
	req1.Header.Set("Content-Type", "application/json")

	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusCreated, w1.Code)

	// Second request with same domain
	req2, _ := http.NewRequest("POST", "/whitelist", bytes.NewBuffer(body))
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusConflict, w2.Code)
}

// TestHandlerListDomains verifies GET /whitelist endpoint
func TestHandlerListDomains(t *testing.T) {
	db, _, router := setupHandlerTest(t)
	db.Create(&WhitelistEntry{
		Domain:      "test1.pharmacy",
		DefaultRole: "CASHIER",
	})

	req, _ := http.NewRequest("GET", "/whitelist", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.True(t, response["success"].(bool))
	data := response["data"].([]interface{})
	assert.GreaterOrEqual(t, len(data), 1)
}

// TestHandlerGetDomain verifies GET /whitelist/:id endpoint
func TestHandlerGetDomain(t *testing.T) {
	db, _, router := setupHandlerTest(t)

	created := &WhitelistEntry{
		Domain:      "test.pharmacy",
		DefaultRole: "CASHIER",
	}
	db.Create(created)

	req, _ := http.NewRequest("GET", "/whitelist/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.True(t, response["success"].(bool))
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "test.pharmacy", data["domain"])
}

// TestHandlerGetDomainNotFound verifies 404 response
func TestHandlerGetDomainNotFound(t *testing.T) {
	_, _, router := setupHandlerTest(t)

	req, _ := http.NewRequest("GET", "/whitelist/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestHandlerGetDomainInvalidID verifies 400 response for invalid ID
func TestHandlerGetDomainInvalidID(t *testing.T) {
	_, _, router := setupHandlerTest(t)

	req, _ := http.NewRequest("GET", "/whitelist/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestHandlerUpdateDomain verifies PUT /whitelist/:id endpoint
func TestHandlerUpdateDomain(t *testing.T) {
	db, _, router := setupHandlerTest(t)

	created := &WhitelistEntry{
		Domain:      "test.pharmacy",
		DefaultRole: "CASHIER",
		Description: "Original",
	}
	db.Create(created)

	updateReq := UpdateWhitelistEntryRequest{
		DefaultRole: "OWNER",
		Description: "Updated",
	}

	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/whitelist/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.True(t, response["success"].(bool))
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "OWNER", data["default_role"])
	assert.Equal(t, "Updated", data["description"])
}

// TestHandlerUpdateDomainNotFound verifies 404 response
func TestHandlerUpdateDomainNotFound(t *testing.T) {
	_, _, router := setupHandlerTest(t)

	updateReq := UpdateWhitelistEntryRequest{
		DefaultRole: "OWNER",
	}

	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/whitelist/999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestHandlerDeleteDomain verifies DELETE /whitelist/:id endpoint
func TestHandlerDeleteDomain(t *testing.T) {
	db, _, router := setupHandlerTest(t)

	created := &WhitelistEntry{
		Domain:      "test.pharmacy",
		DefaultRole: "CASHIER",
	}
	db.Create(created)

	req, _ := http.NewRequest("DELETE", "/whitelist/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

// TestHandlerDeleteDomainNotFound verifies 404 response
func TestHandlerDeleteDomainNotFound(t *testing.T) {
	_, _, router := setupHandlerTest(t)

	req, _ := http.NewRequest("DELETE", "/whitelist/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
