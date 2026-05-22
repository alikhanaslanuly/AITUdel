package tests

import (
	"testing"
	"time"
	jwtpkg "user-service/pkg/jwt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newManager(t *testing.T) *jwtpkg.Manager {
	m, err := jwtpkg.NewManager("test-secret-key-32bytes!!!", "15", "168")
	require.NoError(t, err)
	return m
}

func TestGenerateAccessToken_ContainsCorrectClaims(t *testing.T) {
	m := newManager(t)
	token, err := m.GenerateAccessToken("user-123", "user", true)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := m.Parse(token)
	require.NoError(t, err)

	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "user", claims.Role)
	assert.True(t, claims.IsStudent)
	assert.Equal(t, "access", claims.TokenType)
}

func TestGenerateAccessToken_ExpiryShouldBe15Minutes(t *testing.T) {
	m := newManager(t)
	token, err := m.GenerateAccessToken("user-456", "courier", false)
	require.NoError(t, err)

	claims, err := m.Parse(token)
	require.NoError(t, err)

	remaining := time.Until(claims.ExpiresAt.Time)
	assert.True(t, remaining > 14*time.Minute, "expected ~15m expiry, got %v", remaining)
	assert.True(t, remaining <= 15*time.Minute, "expected ~15m expiry, got %v", remaining)
}

func TestGenerateRefreshToken_ContainsCorrectClaims(t *testing.T) {
	m := newManager(t)
	token, exp, err := m.GenerateRefreshToken("user-789", "admin", false)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, exp.After(time.Now()))

	claims, err := m.Parse(token)
	require.NoError(t, err)

	assert.Equal(t, "user-789", claims.UserID)
	assert.Equal(t, "refresh", claims.TokenType)
}

func TestGenerateRefreshToken_ExpiryIs168Hours(t *testing.T) {
	m := newManager(t)
	_, exp, err := m.GenerateRefreshToken("user-999", "user", false)
	require.NoError(t, err)

	expected := time.Now().Add(168 * time.Hour)
	diff := exp.Sub(expected)
	if diff < 0 {
		diff = -diff
	}
	assert.Less(t, diff, 5*time.Second, "refresh expiry should be ~168h")
}

func TestParse_InvalidTokenReturnsError(t *testing.T) {
	m := newManager(t)
	_, err := m.Parse("this.is.not.a.valid.jwt")
	assert.Error(t, err)
}

func TestParse_WrongSecretReturnsError(t *testing.T) {
	m1, _ := jwtpkg.NewManager("secret-one", "15", "168")
	m2, _ := jwtpkg.NewManager("secret-two", "15", "168")

	token, err := m1.GenerateAccessToken("u1", "user", false)
	require.NoError(t, err)

	_, err = m2.Parse(token)
	assert.Error(t, err)
}

func TestParse_EmptyTokenReturnsError(t *testing.T) {
	m := newManager(t)
	_, err := m.Parse("")
	assert.Error(t, err)
}

func TestIsAITUStudent_ValidDomain(t *testing.T) {
	// We test domain logic directly
	emails := []struct {
		email     string
		isStudent bool
	}{
		{"nursultan@aitu.edu.kz", true},
		{"ali@gmail.com", false},
		{"student@edu.kz", false},
		{"x@aitu.edu.kz", true},
	}

	for _, tc := range emails {
		u := mockUser(tc.email)
		assert.Equal(t, tc.isStudent, u.IsAITUStudent(), "email: %s", tc.email)
	}
}

func TestNewManager_EmptySecretReturnsError(t *testing.T) {
	_, err := jwtpkg.NewManager("", "15", "168")
	assert.Error(t, err)
}

func TestNewManager_InvalidExpiryFallsBackToDefault(t *testing.T) {
	m, err := jwtpkg.NewManager("valid-secret-key!", "bad", "also-bad")
	require.NoError(t, err)
	assert.NotNil(t, m)
}
