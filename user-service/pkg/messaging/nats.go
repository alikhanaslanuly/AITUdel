package messaging

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

type NatsClient struct {
	conn *nats.Conn
}

func NewNats(url string) (*NatsClient, error) {
	conn, err := nats.Connect(url,
		nats.Name("user-service"),
		nats.MaxReconnects(5),
		nats.ReconnectWait(nats.DefaultReconnectWait),
	)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	log.Println("NATS connected")
	return &NatsClient{conn: conn}, nil
}

func (n *NatsClient) Publish(subject string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("nats marshal: %w", err)
	}
	return n.conn.Publish(subject, data)
}

func (n *NatsClient) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return n.conn.Subscribe(subject, handler)
}

func (n *NatsClient) Close() {
	err := n.conn.Drain()
	if err != nil {
		return
	}
}

type UserRegisteredEvent struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	IsStudent bool   `json:"is_student"`
	PromoCode string `json:"promo_code,omitempty"`
}

type OrderDeliveredEvent struct {
	UserID  string  `json:"user_id"`
	OrderID string  `json:"order_id"`
	Total   float64 `json:"total_price"`
}
