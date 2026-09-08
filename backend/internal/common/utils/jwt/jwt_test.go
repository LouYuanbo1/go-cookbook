package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestJWTUtil() JWTUtil {
	return NewJWTUtil("test-secret-key", 24, "test-issuer", []string{"test-audience"})
}

func TestNewJWTUtil(t *testing.T) {
	ju := NewJWTUtil("secret", 24, "issuer", []string{"aud"})
	assert.NotNil(t, ju)
}

func TestGenerateToken_Success(t *testing.T) {
	ju := newTestJWTUtil()
	token, err := ju.GenerateToken("admin")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateToken_DifferentRoles(t *testing.T) {
	ju := newTestJWTUtil()

	roles := []string{"admin", "user", "guest", "moderator"}
	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			token, err := ju.GenerateToken(role)
			require.NoError(t, err)
			assert.NotEmpty(t, token)

			// Parse back and verify role
			claims, err := ju.ParseToken(token)
			require.NoError(t, err)
			assert.Equal(t, role, claims.Role)
		})
	}
}

func TestParseToken_ValidToken(t *testing.T) {
	ju := newTestJWTUtil()
	token, err := ju.GenerateToken("admin")
	require.NoError(t, err)

	claims, err := ju.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, "admin", claims.Role)
	assert.Equal(t, "test-issuer", claims.Issuer)
	assert.NotEmpty(t, claims.ID)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.NotBefore)
}

func TestParseToken_InvalidToken(t *testing.T) {
	ju := newTestJWTUtil()
	_, err := ju.ParseToken("invalid-token")
	assert.Error(t, err)
}

func TestParseToken_EmptyToken(t *testing.T) {
	ju := newTestJWTUtil()
	_, err := ju.ParseToken("")
	assert.Error(t, err)
}

func TestParseToken_WrongSigningMethod(t *testing.T) {
	ju := newTestJWTUtil()
	// Create a token signed with a different method (none)
	claims := CustomClaims{
		Role: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    "test-issuer",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	_, err = ju.ParseToken(tokenString)
	assert.Error(t, err)
}

func TestParseToken_DifferentSecretKey(t *testing.T) {
	ju1 := NewJWTUtil("secret1", 24, "issuer", []string{"aud"})
	ju2 := NewJWTUtil("secret2", 24, "issuer", []string{"aud"})

	token, err := ju1.GenerateToken("admin")
	require.NoError(t, err)

	// Try to parse with a different secret key
	_, err = ju2.ParseToken(token)
	assert.Error(t, err)
}

func TestParseToken_TokenExpired(t *testing.T) {
	// Create a JWT util with 0 hour expiration (already expired)
	ju := NewJWTUtil("test-secret", 0, "test-issuer", []string{"aud"})

	claims := CustomClaims{
		Role: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "test-issuer",
			ID:        "test-id",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	require.NoError(t, err)

	_, err = ju.ParseToken(tokenString)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "已过期")
}

func TestGenerateToken_UniqueTokens(t *testing.T) {
	ju := newTestJWTUtil()
	token1, err := ju.GenerateToken("admin")
	require.NoError(t, err)

	token2, err := ju.GenerateToken("admin")
	require.NoError(t, err)

	// Each token should have a unique JTI
	assert.NotEqual(t, token1, token2)

	claims1, err := ju.ParseToken(token1)
	require.NoError(t, err)
	claims2, err := ju.ParseToken(token2)
	require.NoError(t, err)
	assert.NotEqual(t, claims1.ID, claims2.ID)
}

func TestCustomClaims_ValidTokenClaims(t *testing.T) {
	ju := newTestJWTUtil()
	token, err := ju.GenerateToken("admin")
	require.NoError(t, err)

	claims, err := ju.ParseToken(token)
	require.NoError(t, err)

	// Verify all registered claims
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.NotBefore)
	assert.Equal(t, "test-issuer", claims.Issuer)
	assert.NotEmpty(t, claims.ID)

	// Verify expiry is in the future
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))

	// Verify issued at is in the past
	assert.True(t, claims.IssuedAt.Time.Before(time.Now()))
}

func TestJWTUtil_Interface(t *testing.T) {
	// Ensure jwtUtil implements JWTUtil interface
	var _ JWTUtil = (*jwtUtil)(nil)

	ju := newTestJWTUtil()
	assert.NotNil(t, ju)
}
