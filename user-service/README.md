# AITUdel — User & Auth Service

Microservice responsible for user lifecycle, JWT authentication, courier management and email notifications.

## Architecture

```
user-service/
├── cmd/main.go                         # Entry point — wires all layers
├── proto/user/user.proto               # gRPC contract (9 endpoints)
├── internal/
│   ├── domain/user.go                  # Pure domain models (no framework deps)
│   ├── config/config.go                # Env-based config
│   ├── repository/
│   │   ├── user_repo.go
│   │   ├── token_repo.go
│   │   ├── courier_repo.go
│   │   └── notif_repo.go
│   ├── usecase/
│   │   ├── auth_usecase.go             # Register / Login / Refresh / Logout / Profile
│   │   ├── notif_usecase.go            # Email dispatch + logging
│   │   └── courier_usecase.go          # Courier status management
│   └── delivery/grpc/
│       ├── user_handler.go             # gRPC ↔ usecase mapping
│       └── nats_subscriber.go          # NATS event listeners
├── migrations/                         # golang-migrate SQL files (5 up + 5 down)
├── pkg/
│   ├── database/postgres.go
│   ├── cache/redis.go                  # Refresh-token store + blacklist
│   ├── messaging/nats.go               # Publish / Subscribe helpers
│   ├── jwt/jwt.go                      # Token generation & validation
│   └── email/smtp.go                   # SMTP sender + HTML templates
├── tests/
│   ├── jwt_test.go                     # Unit: 8 JWT tests
│   └── auth_integration_test.go        # Integration: 6 tests via testcontainers
├── docker-compose.yml
└── Dockerfile
```

## gRPC Endpoints (9)

| Method | Description |
|--------|-------------|
| `Register` | Create user/courier; auto-detects AITU students |
| `Login` | Validate credentials, issue JWT pair |
| `RefreshToken` | Rotate refresh token (old token revoked) |
| `Logout` | Blacklist refresh token in Redis |
| `GetProfile` | Return user by ID |
| `UpdateProfile` | Update name / phone |
| `SendNotification` | Internal endpoint — triggers email via SMTP |
| `GetCourier` | Fetch courier info by ID |
| `UpdateCourierStatus` | Set active/offline + GPS coordinates |

## Business Logic Highlights

### Student auto-detection
If the registered email ends with `@aitu.edu.kz`:
- `is_student = true` is set automatically
- `promo_assigned = "AITU2025"` is written in the same DB transaction
- A `user.registered` NATS event carries `promo_code: "AITU2025"` → Order Service activates the 15% discount

### Atomic registration transaction
```
BEGIN
  INSERT users
  INSERT couriers (if role = courier)
  UPDATE users SET promo_assigned (if student)
  INSERT notifications_log
COMMIT
```
Any failure rolls back everything.

### JWT token rotation
- Access token: 15 minutes, HS256
- Refresh token: 7 days, stored in Redis (key `refresh:{userID}:{token}`)
- On `RefreshToken`: old token moved to Redis blacklist (`blacklist:{token}`), new pair issued
- On `Logout`: same blacklist mechanism

### NATS events consumed
| Subject | Action |
|---------|--------|
| `order.delivered` | Send delivery receipt email to user |

### NATS events published
| Subject | When |
|---------|------|
| `user.registered` | After successful registration |

## Running locally

```bash
# 1. Copy environment template
cp .env.example .env
# Fill in SMTP_USER, SMTP_PASSWORD, JWT_SECRET

# 2. Start dependencies
docker-compose up -d postgres redis nats

# 3. Run service
go run ./cmd/main.go
```

## Running tests

```bash
# Unit tests only (no Docker required)
go test ./tests/... -run TestJWT -run TestIsAITU -v

# All tests including integration (requires Docker)
go test ./tests/... -v -timeout 120s

# Skip integration tests
go test ./tests/... -short -v
```

## Migrations

Managed by `golang-migrate`. Applied automatically on startup.

| File | Description |
|------|-------------|
| `001_create_users` | Core users table |
| `002_create_refresh_tokens` | Persistent refresh token store |
| `003_create_couriers` | Courier location + status |
| `004_add_student_flag` | `promo_assigned` column + index |
| `005_create_notifications_log` | Audit log for sent emails |
