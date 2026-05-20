package tests

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"user-service/internal/domain"
	"user-service/internal/repository"
	"user-service/internal/usecase"
	"user-service/pkg/cache"
	jwtpkg "user-service/pkg/jwt"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startPostgres(t *testing.T) (string, func()) {
	t.Helper()
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:17-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf("host=%s port=%s user=test password=test dbname=testdb sslmode=disable", host, port.Port())

	return dsn, func() {
		err := c.Terminate(ctx)
		if err != nil {
			return
		}
	}
}

func startRedis(t *testing.T) (string, func()) {
	t.Helper()
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "redis:8-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(30 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "6379")
	addr := fmt.Sprintf("%s:%s", host, port.Port())

	return addr, func() {
		err := c.Terminate(ctx)
		if err != nil {
			return
		}
	}
}

func applySchema(t *testing.T, db *sql.DB) {
	t.Helper()
	schema := `
	CREATE EXTENSION IF NOT EXISTS pgcrypto;

	CREATE TABLE IF NOT EXISTS users (
		id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email         VARCHAR(255) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		name          VARCHAR(255) NOT NULL,
		phone         VARCHAR(50),
		role          VARCHAR(50)  NOT NULL DEFAULT 'user',
		is_student    BOOLEAN      NOT NULL DEFAULT FALSE,
		promo_assigned VARCHAR(50),
		created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
		updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS refresh_tokens (
		id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token      TEXT        NOT NULL UNIQUE,
		expires_at TIMESTAMPTZ NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS couriers (
		id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id    UUID        NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
		status     VARCHAR(20) NOT NULL DEFAULT 'offline',
		latitude   DOUBLE PRECISION NOT NULL DEFAULT 0,
		longitude  DOUBLE PRECISION NOT NULL DEFAULT 0,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS notifications_log (
		id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		type       VARCHAR(50) NOT NULL,
		subject    TEXT        NOT NULL,
		success    BOOLEAN     NOT NULL DEFAULT FALSE,
		sent_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`

	_, err := db.ExecContext(context.Background(), schema)
	require.NoError(t, err)
}

func buildUsecase(t *testing.T, db *sql.DB, redisAddr string) *usecase.AuthUsecase {
	t.Helper()
	redisClient, err := cache.NewRedis(redisAddr)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := redisClient.Close()
		if err != nil {
			return
		}
	})

	jwtMgr, err := jwtpkg.NewManager("integration-test-secret!", "15", "168")
	require.NoError(t, err)

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	courierRepo := repository.NewCourierRepository(db)
	notifRepo := repository.NewNotifRepository(db)

	return usecase.NewAuthUsecase(userRepo, tokenRepo, courierRepo, notifRepo, redisClient, jwtMgr, nil)
}

func TestIntegration_Register_NormalUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dsn, stopPG := startPostgres(t)
	defer stopPG()
	redisAddr, stopRedis := startRedis(t)
	defer stopRedis()

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)
	applySchema(t, db)

	uc := buildUsecase(t, db, redisAddr)
	ctx := context.Background()

	user, accessToken, refreshToken, err := uc.Register(ctx, "test@example.com", "password123", "Test User", "+7700000000", domain.RoleUser)
	require.NoError(t, err)

	assert.NotEmpty(t, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, domain.RoleUser, user.Role)
	assert.False(t, user.IsStudent)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
}

func TestIntegration_Register_AITUStudent_GetsPromo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dsn, stopPG := startPostgres(t)
	defer stopPG()
	redisAddr, stopRedis := startRedis(t)
	defer stopRedis()

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)
	applySchema(t, db)

	uc := buildUsecase(t, db, redisAddr)
	ctx := context.Background()

	user, _, _, err := uc.Register(ctx, "nursultan@aitu.edu.kz", "pass123", "Nursultan", "", domain.RoleUser)
	require.NoError(t, err)

	assert.True(t, user.IsStudent)
	assert.Equal(t, domain.RoleUser, user.Role)

	var promoAssigned sql.NullString
	err = db.QueryRowContext(ctx, `SELECT promo_assigned FROM users WHERE id = $1`, user.ID).Scan(&promoAssigned)
	require.NoError(t, err)
	assert.Equal(t, "AITU2025", promoAssigned.String)
}

func TestIntegration_Login_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dsn, stopPG := startPostgres(t)
	defer stopPG()
	redisAddr, stopRedis := startRedis(t)
	defer stopRedis()

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)
	applySchema(t, db)

	uc := buildUsecase(t, db, redisAddr)
	ctx := context.Background()

	_, _, _, err = uc.Register(ctx, "login@example.com", "mypassword", "Login Test", "", domain.RoleUser)
	require.NoError(t, err)

	user, access, refresh, err := uc.Login(ctx, "login@example.com", "mypassword")
	require.NoError(t, err)

	assert.Equal(t, "login@example.com", user.Email)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)
}

func TestIntegration_Login_WrongPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dsn, stopPG := startPostgres(t)
	defer stopPG()
	redisAddr, stopRedis := startRedis(t)
	defer stopRedis()

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)
	applySchema(t, db)

	uc := buildUsecase(t, db, redisAddr)
	ctx := context.Background()

	_, _, _, err = uc.Register(ctx, "wrong@example.com", "correct", "Wrong Test", "", domain.RoleUser)
	require.NoError(t, err)

	_, _, _, err = uc.Login(ctx, "wrong@example.com", "wrongpassword")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid credentials")
}

func TestIntegration_RefreshToken_Rotation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dsn, stopPG := startPostgres(t)
	defer stopPG()
	redisAddr, stopRedis := startRedis(t)
	defer stopRedis()

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)
	applySchema(t, db)

	uc := buildUsecase(t, db, redisAddr)
	ctx := context.Background()

	_, _, refreshToken, err := uc.Register(ctx, "refresh@example.com", "pass123", "Refresh Test", "", domain.RoleUser)
	require.NoError(t, err)

	newAccess, newRefresh, err := uc.RefreshToken(ctx, refreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, newAccess)
	assert.NotEmpty(t, newRefresh)
	assert.NotEqual(t, refreshToken, newRefresh, "refresh token should be rotated")

	_, _, err = uc.RefreshToken(ctx, refreshToken)
	assert.Error(t, err)
}

func TestIntegration_DuplicateEmail_ReturnsError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dsn, stopPG := startPostgres(t)
	defer stopPG()
	redisAddr, stopRedis := startRedis(t)
	defer stopRedis()

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)
	applySchema(t, db)

	uc := buildUsecase(t, db, redisAddr)
	ctx := context.Background()

	_, _, _, err = uc.Register(ctx, "dup@example.com", "pass", "User One", "", domain.RoleUser)
	require.NoError(t, err)

	_, _, _, err = uc.Register(ctx, "dup@example.com", "pass", "User Two", "", domain.RoleUser)
	assert.Error(t, err, "duplicate email should return an error")
}
