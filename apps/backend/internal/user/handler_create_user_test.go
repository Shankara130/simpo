package user

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	apiErrors "github.com/vahiiiid/go-rest-api-boilerplate/internal/errors"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
)

// Helper function to set up authenticated context
func setupAuthContext(c *gin.Context, userID uint, email, name, role string) {
	claims := &auth.Claims{
		UserID: userID,
		Email:  email,
		Name:   name,
	}
	c.Set(auth.KeyUser, claims)
	if role != "" {
		c.Set("role", role)
	}
}

// TestHandler_CreateUser tests the CreateUser handler (Story 1.7)
func TestHandler_CreateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		adminID        uint
		adminRole      string
		setupMocks     func(*MockService, *MockAuthService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful creation - SYSTEM_ADMIN user",
			requestBody: CreateUserRequest{
				Username: "newadmin",
				Password: "SecurePass123!",
				Email:    "newadmin@example.com",
				Role:     RoleSystemAdmin,
			},
			adminID:   100,
			adminRole: RoleSystemAdmin,
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				ms.On("RegisterUserForAdmin", mock.Anything, mock.Anything, mock.Anything).Return(&User{
					ID:        5,
					Username:  "newadmin",
					Email:     "newadmin@example.com",
					Role:      RoleSystemAdmin,
					Status:    UserStatusActive,
					CreatedAt: time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response apiErrors.Response
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)

				dataBytes, err := json.Marshal(response.Data)
				require.NoError(t, err)
				var userResp CreateUserResponse
				err = json.Unmarshal(dataBytes, &userResp)
				require.NoError(t, err)

				assert.Equal(t, uint(5), userResp.ID)
				assert.Equal(t, "newadmin", userResp.Username)
				assert.Equal(t, "newadmin@example.com", userResp.Email)
				assert.Equal(t, RoleSystemAdmin, userResp.Role)
			},
		},
		{
			name: "successful creation - OWNER user",
			requestBody: CreateUserRequest{
				Username: "newowner",
				Password: "SecurePass123!",
				Email:    "newowner@example.com",
				Role:     RoleOwner,
			},
			adminID:   100,
			adminRole: RoleSystemAdmin,
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				ms.On("RegisterUserForAdmin", mock.Anything, mock.AnythingOfType("user.CreateUserRequest"), uint(100)).Return(&User{
					ID:        6,
					Username:  "newowner",
					Email:     "newowner@example.com",
					Role:      RoleOwner,
					Status:    UserStatusActive,
					CreatedAt: time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response apiErrors.Response
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response.Success)

				dataBytes, err := json.Marshal(response.Data)
				require.NoError(t, err)
				var userResp CreateUserResponse
				err = json.Unmarshal(dataBytes, &userResp)
				require.NoError(t, err)

				assert.Equal(t, RoleOwner, userResp.Role)
			},
		},
		{
			name: "successful creation - CASHIER user with branch_id",
			requestBody: CreateUserRequest{
				Username: "newcashier",
				Password: "SecurePass123!",
				Email:    "newcashier@example.com",
				Role:     RoleCashier,
				BranchID: uintPtr(10),
			},
			adminID:   100,
			adminRole: RoleSystemAdmin,
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				branchID := uint(10)
				ms.On("RegisterUserForAdmin", mock.Anything, mock.MatchedBy(func(req CreateUserRequest) bool {
					return req.Username == "newcashier" && req.Role == RoleCashier && req.BranchID != nil && *req.BranchID == 10
				}), mock.MatchedBy(func(adminID uint) bool {
					return adminID == 100
				})).Return(&User{
					ID:        7,
					Username:  "newcashier",
					Email:     "newcashier@example.com",
					Role:      RoleCashier,
					BranchID:  &branchID,
					Status:    UserStatusActive,
					CreatedAt: time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response apiErrors.Response
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response.Success)

				dataBytes, err := json.Marshal(response.Data)
				require.NoError(t, err)
				var userResp CreateUserResponse
				err = json.Unmarshal(dataBytes, &userResp)
				require.NoError(t, err)

				assert.Equal(t, RoleCashier, userResp.Role)
				assert.NotNil(t, userResp.BranchID)
				assert.Equal(t, uint(10), *userResp.BranchID)
			},
		},
		{
			name: "duplicate username - returns 409 Conflict",
			requestBody: CreateUserRequest{
				Username: "existinguser",
				Password: "SecurePass123!",
				Email:    "newemail@example.com",
				Role:     RoleCashier,
			},
			adminID:   100,
			adminRole: RoleSystemAdmin,
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				ms.On("RegisterUserForAdmin", mock.Anything, mock.Anything, mock.MatchedBy(func(adminID uint) bool {
					return adminID == 100
				})).Return(nil, ErrUsernameExists)
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response apiErrors.Response
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.Contains(t, response.Error.Detail, "Username already exists")
			},
		},
		{
			name: "duplicate email - returns 409 Conflict",
			requestBody: CreateUserRequest{
				Username: "newuser",
				Password: "SecurePass123!",
				Email:    "existing@example.com",
				Role:     RoleCashier,
			},
			adminID:   100,
			adminRole: RoleSystemAdmin,
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				ms.On("RegisterUserForAdmin", mock.Anything, mock.Anything, mock.MatchedBy(func(adminID uint) bool {
					return adminID == 100
				})).Return(nil, ErrEmailExists)
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response apiErrors.Response
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.Contains(t, response.Error.Detail, "Email already exists")
			},
		},
		{
			name: "invalid role - returns 400 Bad Request",
			requestBody: CreateUserRequest{
				Username: "newuser",
				Password: "SecurePass123!",
				Email:    "new@example.com",
				Role:     "INVALID_ROLE",
			},
			adminID:   100,
			adminRole: RoleSystemAdmin,
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				ms.On("RegisterUserForAdmin", mock.Anything, mock.Anything, mock.MatchedBy(func(adminID uint) bool {
					return adminID == 100
				})).Return(nil, ErrInvalidRoleForCreate)
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response apiErrors.Response
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.Contains(t, response.Error.Detail, "Invalid role")
			},
		},
		{
			name: "missing branch_id for CASHIER - returns 400 Bad Request",
			requestBody: CreateUserRequest{
				Username: "newcashier",
				Password: "SecurePass123!",
				Email:    "cashier@example.com",
				Role:     RoleCashier,
				BranchID: nil,
			},
			adminID:   100,
			adminRole: RoleSystemAdmin,
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				ms.On("RegisterUserForAdmin", mock.Anything, mock.Anything, mock.MatchedBy(func(adminID uint) bool {
					return adminID == 100
				})).Return(nil, ErrBranchIDRequired)
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response apiErrors.Response
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.Contains(t, response.Error.Detail, "branch_id is required")
			},
		},
		{
			name: "unauthenticated request - returns 401 Unauthorized",
			requestBody: CreateUserRequest{
				Username: "newuser",
				Password: "SecurePass123!",
				Email:    "new@example.com",
				Role:     RoleOwner,
			},
			adminID:        0, // No user ID set
			adminRole:      "",
			setupMocks:     func(ms *MockService, mas *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response apiErrors.Response
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.Contains(t, response.Error.Detail, "not authenticated")
			},
		},
		}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			mockAuthService := new(MockAuthService)
			mockAuditLogger := new(MockAuditLogger)
			handler := NewHandler(mockService, mockAuthService, mockAuditLogger)

			tt.setupMocks(mockService, mockAuthService)

			// Set up audit logging expectations for successful tests
			if tt.expectedStatus == http.StatusCreated {
				mockAuditLogger.On("LogUserCreation", mock.Anything, tt.adminID, mock.AnythingOfType("uint"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set up request body
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			c.Request, _ = http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			// Set up authenticated context
			if tt.adminID > 0 {
				setupAuthContext(c, tt.adminID, "admin@example.com", "Admin User", tt.adminRole)
			}

			handler.CreateUser(c)
			apiErrors.ErrorHandler()(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}

			mockService.AssertExpectations(t)
			mockAuditLogger.AssertExpectations(t)
		})
	}
}

// TestHandler_CreateUser_ValidationErrors tests request validation (Story 1.7, AC3)
func TestHandler_CreateUser_ValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockService)
	mockAuthService := new(MockAuthService)
	handler := NewHandler(mockService, mockAuthService, nil)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		shouldContain  string
	}{
		{
			name: "missing username",
			requestBody: map[string]interface{}{
				"password": "SecurePass123!",
				"email":    "test@example.com",
				"role":     RoleOwner,
			},
			expectedStatus: http.StatusBadRequest,
			shouldContain:  "Username",
		},
		{
			name: "username too short",
			requestBody: map[string]interface{}{
				"username": "ab",
				"password": "SecurePass123!",
				"email":    "test@example.com",
				"role":     RoleOwner,
			},
			expectedStatus: http.StatusBadRequest,
			shouldContain:  "Username",
		},
		{
			name: "missing password",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"email":    "test@example.com",
				"role":     RoleOwner,
			},
			expectedStatus: http.StatusBadRequest,
			shouldContain:  "Password",
		},
		{
			name: "password too short",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"password": "short",
				"email":    "test@example.com",
				"role":     RoleOwner,
			},
			expectedStatus: http.StatusBadRequest,
			shouldContain:  "Password",
		},
		{
			name: "missing email",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"password": "SecurePass123!",
				"role":     RoleOwner,
			},
			expectedStatus: http.StatusBadRequest,
			shouldContain:  "Email",
		},
		{
			name: "invalid email format",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"password": "SecurePass123!",
				"email":    "notanemail",
				"role":     RoleOwner,
			},
			expectedStatus: http.StatusBadRequest,
			shouldContain:  "Email",
		},
		{
			name: "missing role",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"password": "SecurePass123!",
				"email":    "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			shouldContain:  "Role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set up authenticated context
			setupAuthContext(c, uint(100), "admin@example.com", "Admin User", RoleSystemAdmin)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			c.Request, _ = http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.CreateUser(c)
			apiErrors.ErrorHandler()(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response apiErrors.Response
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.False(t, response.Success)
			assert.NotEmpty(t, response.Error.Detail)
		})
	}
}

// TestHandler_CreateUser_ServiceErrors tests service error handling (Story 1.7, AC8)
func TestHandler_CreateUser_ServiceErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockService)
	mockAuthService := new(MockAuthService)
	handler := NewHandler(mockService, mockAuthService, nil)

	tests := []struct {
		name           string
		requestBody    CreateUserRequest
		serviceError   error
		expectedStatus int
		errorMessage   string
	}{
		{
			name: "database connection error",
			requestBody: CreateUserRequest{
				Username: "newuser",
				Password: "SecurePass123!",
				Email:    "new@example.com",
				Role:     RoleOwner,
			},
			serviceError:   errors.New("database connection failed"),
			expectedStatus: http.StatusInternalServerError,
			errorMessage:   "database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.On("RegisterUserForAdmin", mock.Anything, mock.Anything, mock.MatchedBy(func(adminID uint) bool {
				return adminID == 100
			})).Return(nil, tt.serviceError)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set up authenticated context
			setupAuthContext(c, uint(100), "admin@example.com", "Admin User", RoleSystemAdmin)

			body, _ := json.Marshal(tt.requestBody)
			c.Request, _ = http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.CreateUser(c)
			apiErrors.ErrorHandler()(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response apiErrors.Response
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.False(t, response.Success)
			assert.NotEmpty(t, response.Error.Detail)

			mockService.AssertExpectations(t)
			mockService.ExpectedCalls = nil // Reset for next test
		})
	}
}
