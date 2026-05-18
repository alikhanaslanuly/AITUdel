package tests

import "user-service/internal/domain"

// mockUser is a thin helper that creates a domain.User with only the email set,
// so we can call IsAITUStudent() in unit tests without a real DB.
func mockUser(email string) *domain.User {
	return &domain.User{Email: email}
}
