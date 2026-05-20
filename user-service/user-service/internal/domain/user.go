package domain

import (
	"strings"
	"time"
)

// Roles
const (
	RoleUser    = "user"
	RoleCourier = "courier"
	RoleAdmin   = "admin"
)

// Courier statuses
const (
	CourierActive  = "active"
	CourierOffline = "offline"
)

// Notification types
const (
	NotifWelcome        = "welcome"
	NotifOrderDelivered = "order_delivered"
	NotifPromo          = "promo"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Name         string
	Phone        string
	Role         string
	IsStudent    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) IsAITUStudent() bool {
	return strings.HasSuffix(strings.ToLower(u.Email), "@aitu.edu.kz")
}

type RefreshToken struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type Courier struct {
	ID        string
	UserID    string
	Status    string
	Latitude  float64
	Longitude float64
	UpdatedAt time.Time
}

type NotificationLog struct {
	ID      string
	UserID  string
	Type    string
	Subject string
	SentAt  time.Time
	Success bool
}
