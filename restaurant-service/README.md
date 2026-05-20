# AITUdel — Restaurant & Catalog Service

Microservice responsible for restaurant management, menu catalog, item search, stock reservation, ratings and cache optimization.

## Architecture

restaurant-service/
├── cmd/main.go                         # Entry point — wires all layers
├── proto/restaurant.proto              # gRPC contract (6 endpoints)
├── internal/
│   ├── domain/
│   │   ├── restaurant.go
│   │   ├── menu_item.go
│   │   └── review.go
│   │
│   ├── repository/
│   │   ├── restaurant_repository.go   # Repository interfaces
│   │   └── postgres_repository.go     # PostgreSQL implementation
│   │
│   ├── usecase/
│   │   ├── restaurant_usecase.go      # Restaurant business logic
│   │   └── stock_usecase.go           # Stock reservation transaction logic
│   │
│   └── delivery/grpc/
│       ├── server.go                  # gRPC ↔ usecase mapping
│       └── health.go                  # Health check endpoint
│
├── migrations/
│   ├── 001_create_restaurants.up.sql
│   ├── 002_create_menu_items.up.sql
│   ├── 003_create_reviews.up.sql
│   ├── 004_add_stock_column.up.sql
│   └── 005_create_categories.up.sql
│
├── pkg/
│   ├── database/postgres.go           # PostgreSQL connection
│   ├── cache/redis.go                 # Redis cache layer
│   └── messaging/
│       ├── nats.go                    # NATS connection
│       └── subscriber.go              # Event subscribers
│
├── tests/
│   ├── setup_test.go                  # Test DB setup
│   └── stock_test.go                  # Stock reservation integration test
│
├── docker-compose.yml
├── Dockerfile
└── go.mod

## gRPC Endpoints (6)

| Method | Description |
|---|---|
| `ListRestaurants` | Return all restaurants |
| `GetRestaurant` | Return restaurant by ID |
| `SearchItems` | Full-text item search |
| `GetItem` | Return menu item by ID |
| `RateRestaurant` | Add review + update rating |
| `HealthCheck` | Service health status |

## Business Logic Highlights

### Restaurant cache layer

Popular restaurant data is cached in Redis.

Cache key format:

restaurant:{id}

TTL:

5 minutes

Flow:

CACHE HIT  → return Redis data  
CACHE MISS → fetch from PostgreSQL → save to Redis

---

### Stock reservation transaction

Restaurant Service subscribes to:

order.created

and reserves stock atomically.

Transaction flow:

BEGIN  
SELECT stock FROM menu_items  
WHERE id = $1  
FOR UPDATE  

UPDATE menu_items  
SET stock = stock - quantity  

COMMIT  

If stock is insufficient:

ROLLBACK

---

### Pessimistic locking

To avoid race conditions during parallel orders:

SELECT ... FOR UPDATE

is used.

This guarantees that stock cannot become negative even under concurrent requests.

---

### Rating recalculation

After every new review:

SELECT AVG(rating)  
FROM reviews  
WHERE restaurant_id = $1

Restaurant rating is updated automatically.

---

### Search system

Menu item search uses:

ILIKE '%query%'

for case-insensitive matching.

## NATS events consumed

| Subject | Action |
|---|---|
| `order.created` | Reserve stock transaction |

## NATS events published

| Subject | When |
|---|---|
| `stock.reserved` | Stock reserved successfully |
| `stock.failed` | Not enough stock / transaction failed |

## Running locally

```bash
# 1. Start PostgreSQL
# 2. Start Redis
# 3. Start NATS

# 4. Run service
go run ./cmd

# Run all tests
go test ./...

# Verbose mode
go test ./... -v

# Run publisher
go run test_publish.go