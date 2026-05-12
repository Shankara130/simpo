package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	apiErrors "github.com/vahiiiid/go-rest-api-boilerplate/internal/errors"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"
)

// TestHandler_DeactivateUser tests the user deactivation endpoint (Story 1.10)
func TestHandler_DeactivateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Now()
	adminID := uint(1)
	targetUserID := uint(10)

	tests := []struct {
		name           string
		targetUserID   string
		requestBody    interface{}
		setupContext   func(*gin.Context)
		setupMocks     func(*MockService, *MockAuthService, *MockAuditLogger)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:         "successful deactivation - Story 1.10 AC1, AC6",
			targetUserID: "10",
			requestBody: DeactivateUserRequest{
				Reason: "Staff resignation",
			},
			setupContext: func(c *gin.Context) {
				// Set auth.Claims for contextutil.GetUserID()
				c.Set("user", &auth.Claims{
					UserID: adminID,
					Email:  "admin@simpo.pharmacy",
					Name:   "System Admin",
					Roles:  []string{"SYSTEM_ADMIN"},
				})
				// Set UserContext for middleware.GetUserRole()
				c.Set(middleware.UserContextKey, &middleware.UserContext{
					UserID:   adminID,
					Username: "admin",
					Email:    "admin@simpo.pharmacy",
					Role:     middleware.RoleSystemAdmin,
				})
			},
			setupMocks: func(ms *MockService, mas *MockAuthService, mal *MockAuditLogger) {
				deactivatedUser := &User{
					ID:                 targetUserID,
					Username:           "formerstaff",
					Email:              "formerstaff@simpo.pharmacy",
					Status:             UserStatusInactive,
					DeactivatedAt:      &now,
					DeactivatedBy:      &adminID,
					DeactivationReason: "Staff resignation",
				}
				ms.On("DeactivateUser", mock.Anything, uint(10), adminID, "Staff resignation").Return(deactivatedUser, nil)
				mal.On("LogUserDeactivation", mock.Anything, adminID, targetUserID, "admin", "formerstaff", "Staff resignation", "127.0.0.1").Return(nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.True(t, response["success"].(bool))
				data := response["data"].(map[string]interface{})
				assert.Equal(t, float64(10), data["id"])
				assert.Equal(t, "formerstaff", data["username"])
				assert.Equal(t, "formerstaff@simpo.pharmacy", data["email"])
				assert.Equal(t, "INACTIVE", data["status"])
				assert.Contains(t, data, "deactivated_at")
				assert.Equal(t, float64(1), data["deactivated_by"])
				assert.Equal(t, "Staff resignation", data["deactivation_reason"])
			},
		},
		{
			name:         "self-deactivation prevented - Story 1.10 AC4",
			targetUserID: "1",
			requestBody: DeactivateUserRequest{
				Reason: "Testing",
			},
			setupContext: func(c *gin.Context) {
				// Set auth.Claims for contextutil.GetUserID()
				c.Set("user", &auth.Claims{
					UserID: adminID,
					Email:  "admin@simpo.pharmacy",
					Name:   "System Admin",
					Roles:  []string{"SYSTEM_ADMIN"},
				})
				// Set UserContext for middleware.GetUserRole()
				c.Set(middleware.UserContextKey, &middleware.UserContext{
					UserID:   adminID,
					Username: "admin",
					Email:    "admin@simpo.pharmacy",
					Role:     middleware.RoleSystemAdmin,
				})
			},
			setupMocks: func(ms *MockService, mas *MockAuthService, mal *MockAuditLogger) {
				ms.On("DeactivateUser", mock.Anything, uint(1), adminID, "Testing").Return(nil, ErrCannotDeactivateSelf)
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.False(t, response["success"].(bool))
				errorInfo := response["error"].(map[string]interface{})
				assert.Contains(t, errorInfo["message"], "cannot deactivate your own account")
			},
		},
		{
			name:         "user not found - Story 1.10 AC7",
			targetUserID: "999",
			requestBody: DeactivateUserRequest{
				Reason: "Testing",
			},
			setupContext: func(c *gin.Context) {
				// Set auth.Claims for contextutil.GetUserID()
				c.Set("user", &auth.Claims{
					UserID: adminID,
					Email:  "admin@simpo.pharmacy",
					Name:   "System Admin",
					Roles:  []string{"SYSTEM_ADMIN"},
				})
				// Set UserContext for middleware.GetUserRole()
				c.Set(middleware.UserContextKey, &middleware.UserContext{
					UserID:   adminID,
					Username: "admin",
					Email:    "admin@simpo.pharmacy",
					Role:     middleware.RoleSystemAdmin,
				})
			},
			setupMocks: func(ms *MockService, mas *MockAuthService, mal *MockAuditLogger) {
				ms.On("DeactivateUser", mock.Anything, uint(999), adminID, "Testing").Return(nil, ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.False(t, response["success"].(bool))
				errorInfo := response["error"].(map[string]interface{})
				assert.Contains(t, errorInfo["message"], "not found")
			},
		},
		{
			name:         "missing reason field - Story 1.10 AC2, AC7",
			targetUserID: "10",
			requestBody: map[string]string{},
			setupContext: func(c *gin.Context) {
				// Set auth.Claims for contextutil.GetUserID()
				c.Set("user", &auth.Claims{
					UserID: adminID,
					Email:  "admin@simpo.pharmacy",
					Name:   "System Admin",
					Roles:  []string{"SYSTEM_ADMIN"},
				})
				// Set UserContext for middleware.GetUserRole()
				c.Set(middleware.UserContextKey, &middleware.UserContext{
					UserID:   adminID,
					Username: "admin",
					Email:    "admin@simpo.pharmacy",
					Role:     middleware.RoleSystemAdmin,
				})
			},
			setupMocks: func(ms *MockService, mas *MockAuthService, mal *MockAuditLogger) {
				// No mocks called due to validation failure
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.False(t, response["success"].(bool))
				errorInfo := response["error"].(map[string]interface{})
				assert.Contains(t, errorInfo["message"], "Reason")
			},
		},
		{
			name:         "non-SYSTEM_ADMIN role denied - Story 1.10 AC1, AC7",
			targetUserID: "10",
			requestBody: DeactivateUserRequest{
				Reason: "Testing",
			},
			setupContext: func(c *gin.Context) {
				// Set auth.Claims for contextutil.GetUserID()
				c.Set("user", &auth.Claims{
					UserID: uint(5),
					Email:  "owner@simpo.pharmacy",
					Name:   "Owner",
					Roles:  []string{"OWNER"},
				})
				// Set UserContext for middleware.GetUserRole()
				c.Set(middleware.UserContextKey, &middleware.UserContext{
					UserID:   uint(5),
					Username: "owner",
					Email:    "owner@simpo.pharmacy",
					Role:     middleware.RoleOwner,
				})
			},
			setupMocks: func(ms *MockService, mas *MockAuthService, mal *MockAuditLogger) {
				// No service call - blocked by RBAC
			},
			expectedStatus: http.StatusForbidden,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.False(t, response["success"].(bool))
				errorInfo := response["error"].(map[string]interface{})
				assert.Contains(t, errorInfo["message"], "SYSTEM_ADMIN")
			},
		},
		{
			name:         "invalid user ID",
			targetUserID: "invalid",
			requestBody: DeactivateUserRequest{
				Reason: "Testing",
			},
			setupContext: func(c *gin.Context) {
				// Set auth.Claims for contextutil.GetUserID()
				c.Set("user", &auth.Claims{
					UserID: adminID,
					Email:  "admin@simpo.pharmacy",
					Name:   "System Admin",
					Roles:  []string{"SYSTEM_ADMIN"},
				})
				// Set UserContext for middleware.GetUserRole()
				c.Set(middleware.UserContextKey, &middleware.UserContext{
					UserID:   adminID,
					Username: "admin",
					Email:    "admin@simpo.pharmacy",
					Role:     middleware.RoleSystemAdmin,
				})
			},
			setupMocks: func(ms *MockService, mas *MockAuthService, mal *MockAuditLogger) {
				// No service call - invalid ID
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.False(t, response["success"].(bool))
				errorInfo := response["error"].(map[string]interface{})
				assert.Contains(t, errorInfo["message"], "Invalid user ID")
			},
		},
		{
			name:         "unauthenticated request",
			targetUserID: "10",
			requestBody: DeactivateUserRequest{
				Reason: "Testing",
			},
			setupContext: func(c *gin.Context) {
				// No user context set
			},
			setupMocks: func(ms *MockService, mas *MockAuthService, mal *MockAuditLogger) {
				// No service call - no auth
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.False(t, response["success"].(bool))
			},
		},
		{
			name:         "idempotent deactivation - already inactive user",
			targetUserID: "10",
			requestBody: DeactivateUserRequest{
				Reason: "Testing",
			},
			setupContext: func(c *gin.Context) {
				// Set auth.Claims for contextutil.GetUserID()
				c.Set("user", &auth.Claims{
					UserID: adminID,
					Email:  "admin@simpo.pharmacy",
					Name:   "System Admin",
					Roles:  []string{"SYSTEM_ADMIN"},
				})
				// Set UserContext for middleware.GetUserRole()
				c.Set(middleware.UserContextKey, &middleware.UserContext{
					UserID:   adminID,
					Username: "admin",
					Email:    "admin@simpo.pharmacy",
					Role:     middleware.RoleSystemAdmin,
				})
			},
			setupMocks: func(ms *MockService, mas *MockAuthService, mal *MockAuditLogger) {
				alreadyInactiveUser := &User{
					ID:                 targetUserID,
					Username:           "formerstaff",
					Email:              "formerstaff@simpo.pharmacy",
					Status:             UserStatusInactive,
					DeactivatedAt:      &now,
					DeactivatedBy:      &adminID,
					DeactivationReason: "Previous deactivation",
				}
				ms.On("DeactivateUser", mock.Anything, uint(10), adminID, "Test").Return(alreadyInactiveUser, nil)
				mal.On("LogUserDeactivation", mock.Anything, adminID, targetUserID, "admin", "formerstaff", "Test", "127.0.0.1").Return(nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Idempotent - returns success even if already inactive
				assert.True(t, response["success"].(bool))
				data := response["data"].(map[string]interface{})
				assert.Equal(t, "INACTIVE", data["status"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockService)
			mockAuthService := new(MockAuthService)
			mockAuditLogger := new(MockAuditLogger)

			tt.setupMocks(mockService, mockAuthService, mockAuditLogger)

			handler := &Handler{
				userService: mockService,
				authService: mockAuthService,
				auditLogger: mockAuditLogger,
			}

			// Create request body
			bodyBytes, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPut, "/api/v1/users/"+tt.targetUserID+"/deactivate", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.RemoteAddr = "127.0.0.1:12345"

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			tt.setupContext(c)

			// Execute
			handler.DeactivateUser(c)

			// Handle errors manually for testing (simulating error handling middleware)
			if len(c.Errors) > 0 {
				err := c.Errors.Last()
				if apiErr, ok := err.Err.(*apiErrors.APIError); ok {
					c.JSON(apiErr.Status, apiErrors.Response{
						Success: false,
						Error: &apiErrors.ErrorInfo{
							Title:  apiErr.Message,
							Status: apiErr.Status,
							Code:   apiErr.Code,
						},
					})
				} else {
					c.JSON(http.StatusInternalServerError, apiErrors.Response{
						Success: false,
						Error: &apiErrors.ErrorInfo{
							Title:  err.Error(),
							Status: http.StatusInternalServerError,
						},
					})
				}
			}

			// Verify
			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)

			// Verify mock expectations
			mockService.AssertExpectations(t)
			mockAuditLogger.AssertExpectations(t)
		})
	}
}

// TestService_DeactivateUser tests the deactivation service logic (Story 1.10)
func TestService_DeactivateUser(t *testing.T) {
	tests := []struct {
		name          string
		targetUserID  uint
		adminID       uint
		reason        string
		setupMocks    func(uint, *MockRepository, *mockSessionManager)
		expectedError error
		verifyUser    func(*testing.T, *User)
	}{
		{
			name:         "successful deactivation",
			targetUserID: 10,
			adminID:      1,
			reason:       "Staff resignation",
			setupMocks: func(adminID uint, mr *MockRepository, msm *mockSessionManager) {
				user := &User{
					ID:       10,
					Username: "staff",
					Email:    "staff@simpo.pharmacy",
					Status:   UserStatusActive,
				}
				mr.On("FindByID", mock.Anything, uint(10)).Return(user, nil).Once()
				mr.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil).Once()
				msm.On("RevokeAllUserTokens", mock.Anything, uint(10)).Return(nil).Once()
				mr.On("Transaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil).Once()
			},
			expectedError: nil,
			verifyUser: func(t *testing.T, u *User) {
				assert.Equal(t, UserStatusInactive, u.Status)
				assert.NotNil(t, u.DeactivatedAt)
				assert.Equal(t, uint(1), *u.DeactivatedBy)
				assert.Equal(t, "Staff resignation", u.DeactivationReason)
			},
		},
		{
			name:         "self-deactivation prevented",
			targetUserID: 1,
			adminID:      1,
			reason:       "Test",
			setupMocks: func(adminID uint, mr *MockRepository, msm *mockSessionManager) {
				// No transaction called for self-deactivation
			},
			expectedError: ErrCannotDeactivateSelf,
			verifyUser:    nil,
		},
		{
			name:         "user not found",
			targetUserID: 999,
			adminID:      1,
			reason:       "Test",
			setupMocks: func(adminID uint, mr *MockRepository, msm *mockSessionManager) {
				mr.On("FindByID", mock.Anything, uint(999)).Return(nil, nil).Once()
				mr.On("Transaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil).Once()
			},
			expectedError: ErrUserNotFound,
			verifyUser:    nil,
		},
		{
			name:         "already inactive - idempotent",
			targetUserID: 10,
			adminID:      1,
			reason:       "Test",
			setupMocks: func(adminID uint, mr *MockRepository, msm *mockSessionManager) {
				user := &User{
					ID:                 10,
					Username:           "staff",
					Email:              "staff@simpo.pharmacy",
					Status:             UserStatusInactive,
					DeactivatedAt:      &time.Time{},
					DeactivatedBy:      &adminID,
					DeactivationReason: "Previous",
				}
				mr.On("FindByID", mock.Anything, uint(10)).Return(user, nil).Once()
				mr.On("Transaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil).Once()
			},
			expectedError: nil,
			verifyUser: func(t *testing.T, u *User) {
				assert.Equal(t, UserStatusInactive, u.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockSessionMgr := new(mockSessionManager)

			tt.setupMocks(tt.adminID, mockRepo, mockSessionMgr)

			svc := &service{
				repo:           mockRepo,
				sessionManager: mockSessionMgr,
			}

			user, err := svc.DeactivateUser(context.Background(), tt.targetUserID, tt.adminID, tt.reason)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError), "expected error %v, got %v", tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				if tt.verifyUser != nil {
					tt.verifyUser(t, user)
				}
			}

			mockRepo.AssertExpectations(t)
			mockSessionMgr.AssertExpectations(t)
		})
	}
}

// mockSessionManager is a mock for testing
type mockSessionManager struct {
	mock.Mock
}

func (m *mockSessionManager) SaveSession(ctx context.Context, tokenID string, session middleware.SessionInfo) error {
	args := m.Called(ctx, tokenID, session)
	return args.Error(0)
}

func (m *mockSessionManager) GetSession(ctx context.Context, userID uint, tokenID string) (*middleware.SessionInfo, error) {
	args := m.Called(ctx, userID, tokenID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*middleware.SessionInfo), args.Error(1)
}

func (m *mockSessionManager) UpdateLastActivity(ctx context.Context, userID uint, tokenID string) error {
	args := m.Called(ctx, userID, tokenID)
	return args.Error(0)
}

func (m *mockSessionManager) DeleteSession(ctx context.Context, userID uint, tokenID string) error {
	args := m.Called(ctx, userID, tokenID)
	return args.Error(0)
}

func (m *mockSessionManager) RevokeToken(ctx context.Context, tokenID string, ttl time.Duration) error {
	args := m.Called(ctx, tokenID, ttl)
	return args.Error(0)
}

func (m *mockSessionManager) IsTokenRevoked(ctx context.Context, tokenID string) (bool, error) {
	args := m.Called(ctx, tokenID)
	return args.Bool(0), args.Error(1)
}

func (m *mockSessionManager) GetAllUserSessions(ctx context.Context, userID uint) ([]middleware.SessionInfo, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]middleware.SessionInfo), args.Error(1)
}

func (m *mockSessionManager) RevokeAllUserSessions(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *mockSessionManager) RevokeAllUserTokens(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
