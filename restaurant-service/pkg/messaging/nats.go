package messaging

import (
	"log"
	"os"

	"github.com/nats-io/nats.go"
)

func NewNATSConnection() *nats.Conn {

	url := os.Getenv("NATS_URL")

	if url == "" {
		url = nats.DefaultURL
	}

	nc, err := nats.Connect(url)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to NATS")

	return nc
}
