package grpc

import (
	"context"
	"encoding/json"
	"log"
	"order-service/internal/usecase"
	"order-service/pkg/messaging"
)

type SagaReply struct {
	OrderID string `json:"order_id"`
}

func SubscribeNATS(nc *messaging.NatsClient, orderUC *usecase.OrderUsecase) {
	nc.Subscribe("stock.reserved", func(data []byte) {
		var reply SagaReply
		if err := json.Unmarshal(data, &reply); err != nil {
			log.Println("stock.reserved unmarshal error:", err)
			return
		}
		if reply.OrderID != "" {
			err := orderUC.ConfirmOrder(context.Background(), reply.OrderID)
			if err != nil {
				log.Println("confirm order error:", err)
			} else {
				log.Println("Order confirmed:", reply.OrderID)
			}
		}
	})

	nc.Subscribe("stock.failed", func(data []byte) {
		var reply SagaReply
		if err := json.Unmarshal(data, &reply); err != nil {
			log.Println("stock.failed unmarshal error:", err)
			return
		}
		if reply.OrderID != "" {
			err := orderUC.FailOrder(context.Background(), reply.OrderID)
			if err != nil {
				log.Println("fail order error:", err)
			} else {
				log.Println("Order failed (stock issues):", reply.OrderID)
			}
		}
	})
}
