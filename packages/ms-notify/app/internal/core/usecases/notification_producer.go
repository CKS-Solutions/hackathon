package usecases

import (
	"context"

	"github.com/cks-solutions/hackathon/ms-notify/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/entities"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/ports"
)

type NotificationProducerUsecase interface {
	Run(ctx context.Context, input dto.NotificationInput) error
}

type NotificationProducerUsecaseImpl struct {
	notificationQueue ports.NotificationQueue
}

func NewNotificationProducerUsecase(queue ports.NotificationQueue) NotificationProducerUsecase {
	return &NotificationProducerUsecaseImpl{
		notificationQueue: queue,
	}
}

func (n *NotificationProducerUsecaseImpl) Run(ctx context.Context, input dto.NotificationInput) error {
	notification := entities.FromInput(input)

	// TODO: validate fields

	err := n.notificationQueue.Push(ctx, notification)
	if err != nil {
		return err
	}

	return nil
}
