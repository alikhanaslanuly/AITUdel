package messaging

import (
	"context"
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"

	"restaurant-service/internal/usecase"
)

type OrderCreatedEvent struct {
	OrderID string              `json:"order_id"`
	Items   []usecase.OrderItem `json:"items"`
}

type SagaReply struct {
	OrderID string `json:"order_id"`
}

func SubscribeOrderCreated(
	nc *nats.Conn,
	stockUsecase *usecase.StockUsecase,
) {

	nc.Subscribe(
		"order.created",
		func(msg *nats.Msg) {

			var event OrderCreatedEvent

			err := json.Unmarshal(
				msg.Data,
				&event,
			)

			if err != nil {
				log.Println(err)
				return
			}

			err = stockUsecase.ReserveStock(
				context.Background(),
				event.Items,
			)

			if err != nil {

				log.Println("Stock failed", err)

				reply, _ := json.Marshal(SagaReply{OrderID: event.OrderID})
				nc.Publish(
					"stock.failed",
					reply,
				)

				return
			}

			log.Println("Stock reserved")

			reply, _ := json.Marshal(SagaReply{OrderID: event.OrderID})
			nc.Publish(
				"stock.reserved",
				reply,
			)
		},
	)
}
