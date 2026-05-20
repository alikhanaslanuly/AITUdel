package usecase

import (
	"context"
	"user-service/internal/domain"
	"user-service/internal/repository"
)

type CourierUsecase struct {
	courierRepo *repository.CourierRepository
}

func NewCourierUsecase(courierRepo *repository.CourierRepository) *CourierUsecase {
	return &CourierUsecase{courierRepo: courierRepo}
}

func (uc *CourierUsecase) GetCourier(ctx context.Context, courierID string) (*domain.Courier, error) {
	return uc.courierRepo.FindByID(ctx, courierID)
}

func (uc *CourierUsecase) UpdateStatus(ctx context.Context, courierID, status string, lat, lng float64) error {
	if status != domain.CourierActive && status != domain.CourierOffline {
		status = domain.CourierOffline
	}
	return uc.courierRepo.UpdateStatus(ctx, courierID, status, lat, lng)
}
