package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go-cookbook/internal/common/utils/jwt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockAuthHandlerService implements AuthService for handler testing
type mockAuthHandlerService struct {
	mock.Mock
}

func (m *mockAuthHandlerService) AdminLogin(ctx context.Context, password string) (string, error) {
	args := m.Called(ctx, password)
	return args.String(0), args.Error(1)
}

func (m *mockAuthHandlerService) ParseToken(tokenString string) (*jwt.CustomClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.CustomClaims), args.Error(1)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func performRequest(router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	var req *http.Request
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req, _ = http.NewRequest(method, path, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest(method, path, nil)
	}
	router.ServeHTTP(w, req)
	return w
}

func TestNewAuthHandler(t *testing.T) {
	mockSvc := new(mockAuthHandlerService)
	handler := NewAuthHandler(mockSvc)
	assert.NotNil(t, handler)
}

func TestHandler_AdminLogin_Success(t *testing.T) {
	mockSvc := new(mockAuthHandlerService)
	mockSvc.On("AdminLogin", mock.Anything, "correct-password").Return("test-token-123", nil)

	handler := NewAuthHandler(mockSvc)
	router := setupTestRouter()
	handler.RegisterRoutes(router)

	w := performRequest(router, "POST", "/api/auth/admin/login", AdminLoginReq{Password: "correct-password"})
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])
	assert.Equal(t, "success", resp["msg"])
	assert.Equal(t, "test-token-123", resp["data"])

	mockSvc.AssertExpectations(t)
}

func TestHandler_AdminLogin_WrongPassword(t *testing.T) {
	mockSvc := new(mockAuthHandlerService)
	mockSvc.On("AdminLogin", mock.Anything, "wrong-password").Return("", ErrPasswordNotMatch)

	handler := NewAuthHandler(mockSvc)
	router := setupTestRouter()
	handler.RegisterRoutes(router)

	w := performRequest(router, "POST", "/api/auth/admin/login", AdminLoginReq{Password: "wrong-password"})
	require.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(401), resp["code"])
	assert.Contains(t, resp["msg"], "密码错误")

	mockSvc.AssertExpectations(t)
}

func TestHandler_AdminLogin_TokenGenerateFailed(t *testing.T) {
	mockSvc := new(mockAuthHandlerService)
	mockSvc.On("AdminLogin", mock.Anything, "correct-password").Return("", ErrGenerateToken)

	handler := NewAuthHandler(mockSvc)
	router := setupTestRouter()
	handler.RegisterRoutes(router)

	w := performRequest(router, "POST", "/api/auth/admin/login", AdminLoginReq{Password: "correct-password"})
	require.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(500), resp["code"])
	assert.Contains(t, resp["msg"], "生成token失败")

	mockSvc.AssertExpectations(t)
}

func TestHandler_AdminLogin_UnknownError(t *testing.T) {
	mockSvc := new(mockAuthHandlerService)
	mockSvc.On("AdminLogin", mock.Anything, "correct-password").Return("", errors.New("some unknown error"))

	handler := NewAuthHandler(mockSvc)
	router := setupTestRouter()
	handler.RegisterRoutes(router)

	w := performRequest(router, "POST", "/api/auth/admin/login", AdminLoginReq{Password: "correct-password"})
	require.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(500), resp["code"])
	assert.Contains(t, resp["msg"], "未知错误")

	mockSvc.AssertExpectations(t)
}

func TestHandler_AdminLogin_InvalidJSON(t *testing.T) {
	mockSvc := new(mockAuthHandlerService)
	handler := NewAuthHandler(mockSvc)
	router := setupTestRouter()
	handler.RegisterRoutes(router)

	// Send invalid JSON
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/admin/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(400), resp["code"])
}

func TestHandler_AdminLogin_MissingPassword(t *testing.T) {
	mockSvc := new(mockAuthHandlerService)
	handler := NewAuthHandler(mockSvc)
	router := setupTestRouter()
	handler.RegisterRoutes(router)

	w := performRequest(router, "POST", "/api/auth/admin/login", AdminLoginReq{Password: ""})
	require.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(400), resp["code"])
}

func TestHandler_AdminLogin_EmptyRequestBody(t *testing.T) {
	mockSvc := new(mockAuthHandlerService)
	handler := NewAuthHandler(mockSvc)
	router := setupTestRouter()
	handler.RegisterRoutes(router)

	w := performRequest(router, "POST", "/api/auth/admin/login", nil)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_RegisterRoutes(t *testing.T) {
	mockSvc := new(mockAuthHandlerService)
	mockSvc.On("AdminLogin", mock.Anything, "test").Return("token", nil)
	handler := NewAuthHandler(mockSvc)
	router := setupTestRouter()
	handler.RegisterRoutes(router)

	// Verify the route exists by making a request
	w := performRequest(router, "POST", "/api/auth/admin/login", AdminLoginReq{Password: "test"})
	// Should get a response (not 404)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_AdminLogin_ConcurrentRequests(t *testing.T) {
	mockSvc := new(mockAuthHandlerService)
	mockSvc.On("AdminLogin", mock.Anything, "password").Return("token", nil)

	handler := NewAuthHandler(mockSvc)
	router := setupTestRouter()
	handler.RegisterRoutes(router)

	// Make concurrent requests
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			w := performRequest(router, "POST", "/api/auth/admin/login", AdminLoginReq{Password: "password"})
			assert.Equal(t, http.StatusOK, w.Code)
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
	mockSvc.AssertExpectations(t)
}
