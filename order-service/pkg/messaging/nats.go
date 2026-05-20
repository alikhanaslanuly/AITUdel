package messaging

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

type NatsClient struct {
	conn *nats.Conn
}

func NewNats(url string) (*NatsClient, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	return &NatsClient{conn: conn}, nil
}

func (n *NatsClient) Publish(subject string, data any) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return n.conn.Publish(subject, bytes)
}

func (n *NatsClient) Subscribe(subject string, handler func(data []byte)) error {
	_, err := n.conn.Subscribe(subject, func(msg *nats.Msg) {
		handler(msg.Data)
	})
	return err
}

func (n *NatsClient) Close() {
	n.conn.Close()
}
