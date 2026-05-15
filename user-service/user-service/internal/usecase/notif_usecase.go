package usecase

import (
	"context"
	"fmt"
	"user-service/internal/domain"
	"user-service/internal/repository"
	"user-service/pkg/email"
)

type NotifUsecase struct {
	userRepo  *repository.UserRepository
	notifRepo *repository.NotifRepository
	sender    *email.SMTPSender
}

func NewNotifUsecase(
	userRepo *repository.UserRepository,
	notifRepo *repository.NotifRepository,
	sender *email.SMTPSender,
) *NotifUsecase {
	return &NotifUsecase{
		userRepo:  userRepo,
		notifRepo: notifRepo,
		sender:    sender,
	}
}

func (uc *NotifUsecase) Send(ctx context.Context, userID, notifType, subject, body string) error {
	u, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if body == "" {
		switch notifType {
		case domain.NotifWelcome:
			if u.IsStudent {
				subject = "Добро пожаловать, студент AITU!"
				body = email.StudentWelcomeEmail(u.Name, "AITU2025")
			} else {
				subject = "Добро пожаловать в AITUdel! "
				body = email.WelcomeEmail(u.Name)
			}
		case domain.NotifOrderDelivered:
			if subject == "" {
				subject = "Ваш заказ доставлен! "
			}
			body = email.OrderDeliveredEmail(u.Name, "", 0)
		}
	} else if notifType == domain.NotifOrderDelivered {
		body = email.InjectNameIntoOrderEmail(u.Name, body)
	}

	sendErr := uc.sender.Send(u.Email, subject, body)

	log := &domain.NotificationLog{
		UserID:  userID,
		Type:    notifType,
		Subject: subject,
		Success: sendErr == nil,
	}
	_ = uc.notifRepo.Log(ctx, nil, log)

	return sendErr
}
