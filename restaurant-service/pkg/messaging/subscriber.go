package messaging

import (
	"context"
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"

	"restaurant-service/internal/usecase"
)

type OrderCreatedEvent struct {
	ItemID   int64 `json:"item_id"`
	Quantity int   `json:"quantity"`
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
				event.ItemID,
				event.Quantity,
			)

			if err != nil {

				log.Println("Stock failed")

				nc.Publish(
					"stock.failed",
					[]byte(err.Error()),
				)

				return
			}

			log.Println("Stock reserved")

			nc.Publish(
				"stock.reserved",
				[]byte("success"),
			)
		},
	)
}
