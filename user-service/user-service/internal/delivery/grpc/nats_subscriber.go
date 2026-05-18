package grpc

import (
	"context"
	"encoding/json"
	"log"
	"user-service/internal/domain"
	"user-service/internal/usecase"
	"user-service/pkg/email"
	"user-service/pkg/messaging"

	"github.com/nats-io/nats.go"
)

func SubscribeNATS(nc *messaging.NatsClient, notifUC *usecase.NotifUsecase) {
	_, err := nc.Subscribe("order.delivered", func(msg *nats.Msg) {
		var event messaging.OrderDeliveredEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("NATS order.delivered: unmarshal error: %v", err)
			return
		}
		log.Printf("NATS order.delivered: userID=%s orderID=%s", event.UserID, event.OrderID)

		emailSubject := "Ваш заказ доставлен!"
		emailBody := email.OrderDeliveredEmail("", event.OrderID, event.Total)

		if err := notifUC.Send(
			context.Background(),
			event.UserID,
			domain.NotifOrderDelivered,
			emailSubject,
			emailBody,
		); err != nil {
			log.Printf("NATS order.delivered: send email error: %v", err)
		}
	})
	if err != nil {
		log.Printf("NATS subscribe order.delivered: %v", err)
	}
}
