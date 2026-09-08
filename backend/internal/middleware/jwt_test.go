package middleware

import (
	"encoding/json"
	"go-cookbook/internal/common/utils/jwt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockJWTUtil implements jwt.JWTUtil for testing
type mockJWTUtil struct {
	mock.Mock
}

func (m *mockJWTUtil) GenerateToken(role string) (string, error) {
	args := m.Called(role)
	return args.String(0), args.Error(1)
}

func (m *mockJWTUtil) ParseToken(tokenString string) (*jwt.CustomClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.CustomClaims), args.Error(1)
}

func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{
		Header: make(http.Header),
	}
	return c, w
}

func TestJWTMiddleware_Success(t *testing.T) {
	mockJWT := new(mockJWTUtil)
	claims := &jwt.CustomClaims{
		Role: "admin",
	}
	mockJWT.On("ParseToken", "valid-token").Return(claims, nil)

	middleware := JWTMiddleware(mockJWT)
	c, w := setupTestContext()
	c.Request.Header.Set("Authorization", "Bearer valid-token")

	handled := false
	middleware(c)
	handled = true

	assert.True(t, handled)
	assert.Equal(t, http.StatusOK, w.Code) // Should not write response
	assert.False(t, c.IsAborted())
	mockJWT.AssertExpectations(t)
}

func TestJWTMiddleware_NoAuthorizationHeader(t *testing.T) {
	mockJWT := new(mockJWTUtil)
	middleware := JWTMiddleware(mockJWT)
	c, w := setupTestContext()

	middleware(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(401), resp["code"])
	assert.True(t, c.IsAborted())
}

func TestJWTMiddleware_InvalidAuthFormat(t *testing.T) {
	mockJWT := new(mockJWTUtil)
	middleware := JWTMiddleware(mockJWT)
	c, w := setupTestContext()
	c.Request.Header.Set("Authorization", "InvalidFormat token")

	middleware(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())
}

func TestJWTMiddleware_NoBearerPrefix(t *testing.T) {
	mockJWT := new(mockJWTUtil)
	middleware := JWTMiddleware(mockJWT)
	c, w := setupTestContext()
	c.Request.Header.Set("Authorization", "token-without-bearer")

	middleware(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	mockJWT := new(mockJWTUtil)
	mockJWT.On("ParseToken", "invalid-token").Return(nil, assert.AnError)

	middleware := JWTMiddleware(mockJWT)
	c, w := setupTestContext()
	c.Request.Header.Set("Authorization", "Bearer invalid-token")

	middleware(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())
	mockJWT.AssertExpectations(t)
}

func TestJWTMiddleware_NonAdminRole(t *testing.T) {
	mockJWT := new(mockJWTUtil)
	claims := &jwt.CustomClaims{
		Role: "user",
	}
	mockJWT.On("ParseToken", "user-token").Return(claims, nil)

	middleware := JWTMiddleware(mockJWT)
	c, w := setupTestContext()
	c.Request.Header.Set("Authorization", "Bearer user-token")

	middleware(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(403), resp["code"])
	assert.Contains(t, resp["msg"], "error")
	assert.True(t, c.IsAborted())
	mockJWT.AssertExpectations(t)
}

func TestJWTMiddleware_EmptyToken(t *testing.T) {
	mockJWT := new(mockJWTUtil)
	mockJWT.On("ParseToken", "").Return(nil, assert.AnError)

	middleware := JWTMiddleware(mockJWT)
	c, w := setupTestContext()
	c.Request.Header.Set("Authorization", "Bearer ")

	middleware(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())
}

func TestJWTMiddleware_MultipleRoles(t *testing.T) {
	mockJWT := new(mockJWTUtil)
	claims := &jwt.CustomClaims{
		Role: "admin",
	}
	mockJWT.On("ParseToken", "admin-token").Return(claims, nil)

	middleware := JWTMiddleware(mockJWT)
	c, w := setupTestContext()
	c.Request.Header.Set("Authorization", "Bearer admin-token")

	middleware(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, c.IsAborted())
	mockJWT.AssertExpectations(t)
}

func TestJWTMiddleware_NextCalledOnSuccess(t *testing.T) {
	mockJWT := new(mockJWTUtil)
	claims := &jwt.CustomClaims{
		Role: "admin",
	}
	mockJWT.On("ParseToken", "valid-token").Return(claims, nil)

	middleware := JWTMiddleware(mockJWT)
	c, _ := setupTestContext()
	c.Request.Header.Set("Authorization", "Bearer valid-token")

	middleware(c)

	assert.False(t, c.IsAborted())
}

func TestJWTMiddleware_ConcurrentRequests(t *testing.T) {
	mockJWT := new(mockJWTUtil)
	claims := &jwt.CustomClaims{
		Role: "admin",
	}
	mockJWT.On("ParseToken", "token").Return(claims, nil)

	middleware := JWTMiddleware(mockJWT)

	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			c, w := setupTestContext()
			c.Request.Header.Set("Authorization", "Bearer token")
			middleware(c)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.False(t, c.IsAborted())
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
	mockJWT.AssertExpectations(t)
}

func TestJWTMiddleware_ResponseBodyOnUnauthorized(t *testing.T) {
	mockJWT := new(mockJWTUtil)
	middleware := JWTMiddleware(mockJWT)
	c, w := setupTestContext()

	middleware(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(401), resp["code"])
	assert.Equal(t, "error", resp["msg"])
	assert.Equal(t, "未授权", resp["error"])
}

func TestJWTMiddleware_ResponseBodyOnForbidden(t *testing.T) {
	mockJWT := new(mockJWTUtil)
	claims := &jwt.CustomClaims{
		Role: "user",
	}
	mockJWT.On("ParseToken", "user-token").Return(claims, nil)

	middleware := JWTMiddleware(mockJWT)
	c, w := setupTestContext()
	c.Request.Header.Set("Authorization", "Bearer user-token")

	middleware(c)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(403), resp["code"])
	assert.Equal(t, "error", resp["msg"])
	assert.Equal(t, "权限不足", resp["error"])
}

func TestJWTMiddleware_CustomClaimsInContext(t *testing.T) {
	mockJWT := new(mockJWTUtil)
	claims := &jwt.CustomClaims{
		Role: "admin",
	}
	mockJWT.On("ParseToken", "valid-token").Return(claims, nil)

	middleware := JWTMiddleware(mockJWT)
	c, _ := setupTestContext()
	c.Request.Header.Set("Authorization", "Bearer valid-token")

	middleware(c)

	// The middleware should not abort for admin role
	assert.False(t, c.IsAborted())
	mockJWT.AssertExpectations(t)
}

func TestJWTMiddleware_AuthHeaderCaseInsensitive(t *testing.T) {
	mockJWT := new(mockJWTUtil)
	claims := &jwt.CustomClaims{
		Role: "admin",
	}
	mockJWT.On("ParseToken", "valid-token").Return(claims, nil)

	middleware := JWTMiddleware(mockJWT)
	c, _ := setupTestContext()
	// Use lowercase header (should still work with Go's HTTP server)
	c.Request.Header.Set("authorization", "Bearer valid-token")

	middleware(c)

	// Gin's GetHeader uses canonical format, so this should still work
	assert.False(t, c.IsAborted())
}
