package usecase

import (
	"context"
	"fmt"
	"time"
	"user-service/internal/domain"
	"user-service/internal/repository"
	"user-service/pkg/cache"
	jwtpkg "user-service/pkg/jwt"
	"user-service/pkg/messaging"

	"golang.org/x/crypto/bcrypt"
)

const studentPromoCode = "AITU2025"

type AuthUsecase struct {
	userRepo    *repository.UserRepository
	tokenRepo   *repository.TokenRepository
	courierRepo *repository.CourierRepository
	notifRepo   *repository.NotifRepository
	redis       *cache.RedisClient
	jwt         *jwtpkg.Manager
	nats        *messaging.NatsClient
}

func NewAuthUsecase(
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	courierRepo *repository.CourierRepository,
	notifRepo *repository.NotifRepository,
	redis *cache.RedisClient,
	jwt *jwtpkg.Manager,
	nats *messaging.NatsClient,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		courierRepo: courierRepo,
		notifRepo:   notifRepo,
		redis:       redis,
		jwt:         jwt,
		nats:        nats,
	}
}

func (uc *AuthUsecase) Register(ctx context.Context, email, password, name, phone, role string) (*domain.User, string, string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", fmt.Errorf("hash password: %w", err)
	}

	u := &domain.User{
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Phone:        phone,
		Role:         role,
	}

	if u.IsAITUStudent() {
		u.IsStudent = true
		u.Role = domain.RoleUser
	}
	if role == domain.RoleCourier {
		u.Role = domain.RoleCourier
	}

	tx, err := uc.userRepo.BeginTx(ctx)
	if err != nil {
		return nil, "", "", fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return
			}
		}
	}()

	if err = uc.userRepo.Create(ctx, tx, u); err != nil {
		return nil, "", "", fmt.Errorf("create user: %w", err)
	}

	if u.Role == domain.RoleCourier {
		c := &domain.Courier{
			UserID: u.ID,
			Status: domain.CourierOffline,
		}
		if err = uc.courierRepo.Create(ctx, tx, c); err != nil {
			return nil, "", "", fmt.Errorf("create courier: %w", err)
		}
	}

	promoCode := ""
	if u.IsStudent {
		promoCode = studentPromoCode
		if err = uc.userRepo.SetPromoAssigned(ctx, tx, u.ID, promoCode); err != nil {
			return nil, "", "", fmt.Errorf("set promo: %w", err)
		}
	}

	subject := "Добро пожаловать в AITUdel!"
	if u.IsStudent {
		subject = "Добро пожаловать, студент AITU!"
	}
	notif := &domain.NotificationLog{
		UserID:  u.ID,
		Type:    domain.NotifWelcome,
		Subject: subject,
		Success: false,
	}
	if err = uc.notifRepo.Log(ctx, tx, notif); err != nil {
		return nil, "", "", fmt.Errorf("log notif: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, "", "", fmt.Errorf("commit tx: %w", err)
	}

	if uc.nats != nil {
		event := messaging.UserRegisteredEvent{
			UserID:    u.ID,
			Email:     u.Email,
			Name:      u.Name,
			IsStudent: u.IsStudent,
			PromoCode: promoCode,
		}
		if pubErr := uc.nats.Publish("user.registered", event); pubErr != nil {
			fmt.Printf("failed to publish user registered event: %v\n", pubErr)
		}
	}

	accessToken, err := uc.jwt.GenerateAccessToken(u.ID, u.Role, u.IsStudent)
	if err != nil {
		return nil, "", "", fmt.Errorf("access token: %w", err)
	}
	refreshToken, exp, err := uc.jwt.GenerateRefreshToken(u.ID, u.Role, u.IsStudent)
	if err != nil {
		return nil, "", "", fmt.Errorf("refresh token: %w", err)
	}

	err = uc.redis.StoreRefreshToken(ctx, u.ID, refreshToken, uc.jwt.RefreshExpiry())
	if err != nil {
		return nil, "", "", err
	}
	rt := &domain.RefreshToken{UserID: u.ID, Token: refreshToken, ExpiresAt: exp}
	if err := uc.tokenRepo.Save(ctx, nil, rt); err != nil {
		fmt.Printf("failed to save refresh token: %v\n", err)
	}

	return u, accessToken, refreshToken, nil
}

func (uc *AuthUsecase) Login(ctx context.Context, email, password string) (*domain.User, string, string, error) {
	u, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, "", "", fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", fmt.Errorf("invalid credentials")
	}

	accessToken, err := uc.jwt.GenerateAccessToken(u.ID, u.Role, u.IsStudent)
	if err != nil {
		return nil, "", "", fmt.Errorf("access token: %w", err)
	}
	refreshToken, exp, err := uc.jwt.GenerateRefreshToken(u.ID, u.Role, u.IsStudent)
	if err != nil {
		return nil, "", "", fmt.Errorf("refresh token: %w", err)
	}

	err = uc.redis.StoreRefreshToken(ctx, u.ID, refreshToken, uc.jwt.RefreshExpiry())
	if err != nil {
		return nil, "", "", err
	}
	rt := &domain.RefreshToken{UserID: u.ID, Token: refreshToken, ExpiresAt: exp}
	if err := uc.tokenRepo.Save(ctx, nil, rt); err != nil {
		fmt.Printf("failed to save refresh token: %v\n", err)
	}

	return u, accessToken, refreshToken, nil
}

func (uc *AuthUsecase) RefreshToken(ctx context.Context, oldRefreshToken string) (string, string, error) {
	claims, err := uc.jwt.Parse(oldRefreshToken)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token")
	}
	if claims.TokenType != "refresh" {
		return "", "", fmt.Errorf("invalid token type")
	}

	valid, err := uc.redis.ValidateRefreshToken(ctx, claims.UserID, oldRefreshToken)
	if err != nil || !valid {
		return "", "", fmt.Errorf("refresh token revoked or expired")
	}

	err = uc.redis.RevokeRefreshToken(ctx, claims.UserID, oldRefreshToken, uc.jwt.RefreshExpiry())
	if err != nil {
		return "", "", err
	}
	err = uc.tokenRepo.Delete(ctx, oldRefreshToken)
	if err != nil {
		return "", "", err
	}

	accessToken, err := uc.jwt.GenerateAccessToken(claims.UserID, claims.Role, claims.IsStudent)
	if err != nil {
		return "", "", err
	}
	newRefresh, exp, err := uc.jwt.GenerateRefreshToken(claims.UserID, claims.Role, claims.IsStudent)
	if err != nil {
		return "", "", err
	}

	err = uc.redis.StoreRefreshToken(ctx, claims.UserID, newRefresh, uc.jwt.RefreshExpiry())
	if err != nil {
		return "", "", err
	}
	rt := &domain.RefreshToken{UserID: claims.UserID, Token: newRefresh, ExpiresAt: exp}
	_ = uc.tokenRepo.Save(ctx, nil, rt)

	return accessToken, newRefresh, nil
}

func (uc *AuthUsecase) Logout(ctx context.Context, refreshToken string) error {
	claims, err := uc.jwt.Parse(refreshToken)
	if err != nil {
		return nil
	}
	err = uc.redis.RevokeRefreshToken(ctx, claims.UserID, refreshToken, uc.jwt.RefreshExpiry())
	if err != nil {
		return err
	}
	return uc.tokenRepo.Delete(ctx, refreshToken)
}

func (uc *AuthUsecase) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	return uc.userRepo.FindByID(ctx, userID)
}

func (uc *AuthUsecase) UpdateProfile(ctx context.Context, userID, name, phone string) (*domain.User, error) {
	u, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if name != "" {
		u.Name = name
	}
	if phone != "" {
		u.Phone = phone
	}
	u.UpdatedAt = time.Now()
	if err := uc.userRepo.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (uc *AuthUsecase) HandleOrderDelivered(userID, orderID string, total float64) {
	ctx := context.Background()
	u, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return
	}
	_ = u
}
