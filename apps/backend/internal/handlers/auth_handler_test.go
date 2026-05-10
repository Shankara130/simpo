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
	"github.com/stretchr/testify/require"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// TestAuthHandler_Login_Success tests successful login via HTTP endpoint (Story 1.5, AC1, AC4)
func TestAuthHandler_Login_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockUserService := &MockAuthService{
		loginFunc: func(ctx context.Context, username, password, ipAddress string) (*dto.LoginResponse, error) {
			return &dto.LoginResponse{
				AccessToken: "test.jwt.token",
				TokenType:   "Bearer",
				ExpiresIn:   28800,
				User: dto.UserInfo{
					ID:       1,
					Username: "admin",
					Email:    "admin@simpo.pharmacy",
					Role:     user.RoleSystemAdmin,
					BranchID: nil,
				},
			}, nil
		},
	}

	handler := NewAuthHandler(mockUserService)
	router.POST("/api/v1/auth/login", handler.Login)

	reqBody := `{"username":"admin","password":"SecurePassword123!"}`
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Should return 200 OK")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "test.jwt.token", data["access_token"])
	assert.Equal(t, "Bearer", data["token_type"])
	assert.Equal(t, float64(28800), data["expires_in"])

	userData := data["user"].(map[string]interface{})
	assert.Equal(t, float64(1), userData["id"])
	assert.Equal(t, "admin", userData["username"])
	assert.Equal(t, "SYSTEM_ADMIN", userData["role"])
}

// TestAuthHandler_Login_InvalidCredentials tests handler returns error for invalid credentials
func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockUserService := &MockAuthService{
		loginFunc: func(ctx context.Context, username, password, ipAddress string) (*dto.LoginResponse, error) {
			return nil, services.ErrInvalidPassword
		},
	}

	handler := NewAuthHandler(mockUserService)
	router.POST("/api/v1/auth/login", handler.Login)

	reqBody := `{"username":"admin","password":"WrongPassword"}`
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Without error middleware, check that no success response is written
	assert.NotEqual(t, http.StatusOK, w.Code, "Should not return 200 OK for invalid credentials")
	assert.True(t, len(w.Body.Bytes()) == 0 || w.Code != http.StatusOK, "Should not write success body")
}

// TestAuthHandler_Login_InactiveUser tests handler returns error for inactive user
func TestAuthHandler_Login_InactiveUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockUserService := &MockAuthService{
		loginFunc: func(ctx context.Context, username, password, ipAddress string) (*dto.LoginResponse, error) {
			return nil, services.ErrUserInactive
		},
	}

	handler := NewAuthHandler(mockUserService)
	router.POST("/api/v1/auth/login", handler.Login)

	reqBody := `{"username":"admin","password":"Password123!"}`
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusOK, w.Code, "Should not return 200 OK for inactive user")
	assert.True(t, len(w.Body.Bytes()) == 0 || w.Code != http.StatusOK, "Should not write success body")
}

// MockAuthService is a mock for testing
type MockAuthService struct {
	loginFunc func(ctx context.Context, username, password, ipAddress string) (*dto.LoginResponse, error)
}

func (m *MockAuthService) Login(ctx context.Context, username, password, ipAddress string) (*dto.LoginResponse, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, username, password, ipAddress)
	}
	return nil, nil
}
