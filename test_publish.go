package main

import (
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
)

type OrderCreatedEvent struct {
	ItemID   int64 `json:"item_id"`
	Quantity int   `json:"quantity"`
}

func main() {

	nc, err := nats.Connect(
		nats.DefaultURL,
	)

	if err != nil {
		log.Fatal(err)
	}

	event := OrderCreatedEvent{
		ItemID:   1,
		Quantity: 2,
	}

	data, _ := json.Marshal(event)

	err = nc.Publish(
		"order.created",
		data,
	)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("order.created published")
}
