package authController

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	authService "go-cookbook/internal/service/auth"
	"go-cookbook/internal/service/jwt"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) AdminLogin(ctx context.Context, password string) (string, error) {
	args := m.Called(ctx, password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ParseToken(tokenString string) (*jwt.CustomClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.CustomClaims), args.Error(1)
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	return r
}

func TestAuthController_AdminLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	mockService := new(MockAuthService)
	controller := NewAuthController(mockService)

	router.POST("/api/auth/admin/login", controller.AdminLogin)

	mockService.On("AdminLogin", mock.Anything, "correct_password").Return("test_token", nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/auth/admin/login", strings.NewReader(`{"password":"correct_password"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["msg"])
	assert.Equal(t, "test_token", response["data"])
}

func TestAuthController_AdminLogin_WrongPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	mockService := new(MockAuthService)
	controller := NewAuthController(mockService)

	router.POST("/api/auth/admin/login", controller.AdminLogin)

	mockService.On("AdminLogin", mock.Anything, "wrong_password").Return("", authService.ErrPasswordNotMatch)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/auth/admin/login", strings.NewReader(`{"password":"wrong_password"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response["msg"])
	assert.Equal(t, "密码错误", response["error"])
}

func TestAuthController_AdminLogin_GenerateTokenError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	mockService := new(MockAuthService)
	controller := NewAuthController(mockService)

	router.POST("/api/auth/admin/login", controller.AdminLogin)

	mockService.On("AdminLogin", mock.Anything, "password").Return("", authService.ErrGenerateToken)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/auth/admin/login", strings.NewReader(`{"password":"password"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response["msg"])
	assert.Equal(t, "生成token失败", response["error"])
}

func TestAuthController_AdminLogin_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	mockService := new(MockAuthService)
	controller := NewAuthController(mockService)

	router.POST("/api/auth/admin/login", controller.AdminLogin)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/auth/admin/login", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
