package jwt

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	IsStudent bool   `json:"is_student"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret          []byte
	accessExpMin    int
	refreshExpHours int
}

func NewManager(secret, accessExpMin, refreshExpHours string) (*Manager, error) {
	if secret == "" {
		return nil, errors.New("jwt: secret must not be empty")
	}
	aExp, err := strconv.Atoi(accessExpMin)
	if err != nil {
		aExp = 15
	}
	rExp, err := strconv.Atoi(refreshExpHours)
	if err != nil {
		rExp = 168
	}
	return &Manager{
		secret:          []byte(secret),
		accessExpMin:    aExp,
		refreshExpHours: rExp,
	}, nil
}

func (m *Manager) GenerateAccessToken(userID, role string, isStudent bool) (string, error) {
	claims := &Claims{
		UserID:    userID,
		Role:      role,
		IsStudent: isStudent,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(m.accessExpMin) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) GenerateRefreshToken(userID, role string, isStudent bool) (string, time.Time, error) {
	exp := time.Now().Add(time.Duration(m.refreshExpHours) * time.Hour)
	claims := &Claims{
		UserID:    userID,
		Role:      role,
		IsStudent: isStudent,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	return signed, exp, err
}

func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt parse: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("jwt: invalid token")
	}
	return claims, nil
}

func (m *Manager) RefreshExpiry() time.Duration {
	return time.Duration(m.refreshExpHours) * time.Hour
}
