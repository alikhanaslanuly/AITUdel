package tests

import "user-service/internal/domain"

func mockUser(email string) *domain.User {
	return &domain.User{Email: email}
}
