package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuthService(t *testing.T) {
	svc := NewAuthService("test-password")
	assert.NotNil(t, svc)
}

func TestService_AdminLogin_Success(t *testing.T) {
	svc := NewAuthService("test-password")
	token, err := svc.AdminLogin(context.Background(), "test-password")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestService_AdminLogin_WrongPassword(t *testing.T) {
	svc := NewAuthService("test-password")
	_, err := svc.AdminLogin(context.Background(), "wrong-password")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPasswordNotMatch)
}

func TestService_AdminLogin_EmptyPassword(t *testing.T) {
	svc := NewAuthService("test-password")
	_, err := svc.AdminLogin(context.Background(), "")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPasswordNotMatch)
}

func TestService_AdminLogin_EmptyStoredPassword(t *testing.T) {
	svc := NewAuthService("")
	// When both stored and provided passwords are empty,
	// bcrypt will compare hash of "" with "", which matches
	token, err := svc.AdminLogin(context.Background(), "")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestService_ParseToken_ValidToken(t *testing.T) {
	svc := NewAuthService("test-password")
	token, err := svc.AdminLogin(context.Background(), "test-password")
	require.NoError(t, err)

	claims, err := svc.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, "admin", claims.Role)
}

func TestService_ParseToken_InvalidToken(t *testing.T) {
	svc := NewAuthService("test-password")
	_, err := svc.ParseToken("invalid-token")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCheckToken)
}

func TestService_ParseToken_EmptyToken(t *testing.T) {
	svc := NewAuthService("test-password")
	_, err := svc.ParseToken("")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCheckToken)
}

func TestService_ParseToken_TokenFromDifferentService(t *testing.T) {
	svc1 := NewAuthService("password1")
	svc2 := NewAuthService("password2")

	token, err := svc1.AdminLogin(context.Background(), "password1")
	require.NoError(t, err)

	// Try to parse with different service (different jwt secret)
	_, err = svc2.ParseToken(token)
	assert.Error(t, err)
}

func TestService_AdminLogin_ReturnsValidToken(t *testing.T) {
	svc := NewAuthService("test-password")
	token, err := svc.AdminLogin(context.Background(), "test-password")
	require.NoError(t, err)

	// Parse the returned token and verify it's valid
	claims, err := svc.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, "admin", claims.Role)
	assert.NotEmpty(t, claims.ID)
	assert.NotEmpty(t, claims.Issuer)
}

func TestService_AuthService_Interface(t *testing.T) {
	// Ensure authService implements AuthService interface
	var _ AuthService = (*authService)(nil)

	svc := NewAuthService("test")
	assert.NotNil(t, svc)
}

func TestService_MultipleAdminLogins_UniqueTokens(t *testing.T) {
	svc := NewAuthService("test-password")

	token1, err := svc.AdminLogin(context.Background(), "test-password")
	require.NoError(t, err)

	token2, err := svc.AdminLogin(context.Background(), "test-password")
	require.NoError(t, err)

	assert.NotEqual(t, token1, token2)
}

func TestService_AdminLogin_ContextCancelled(t *testing.T) {
	svc := NewAuthService("test-password")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// The auth service doesn't actually use the context for DB operations,
	// so this should still succeed
	token, err := svc.AdminLogin(ctx, "test-password")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}
