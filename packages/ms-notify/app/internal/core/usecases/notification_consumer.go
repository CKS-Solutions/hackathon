package usecases

import (
	"context"
	"fmt"

	"github.com/cks-solutions/hackathon/ms-notify/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/entities"
	ports "github.com/cks-solutions/hackathon/ms-notify/internal/core/ports"
)

type NotificationConsumerUsecase interface {
	Run(ctx context.Context, input dto.NotificationInput) error
}

type NotificationConsumerUsecaseImpl struct {
	notificationTable ports.NotificationTable
	emailService      ports.EmailService
}

func NewNotificationConsumerUsecase(
	notificationTable ports.NotificationTable,
	emailService ports.EmailService,
) NotificationConsumerUsecase {
	return &NotificationConsumerUsecaseImpl{
		notificationTable: notificationTable,
		emailService:      emailService,
	}
}

func (n *NotificationConsumerUsecaseImpl) Run(ctx context.Context, input dto.NotificationInput) error {
	notification := entities.FromInput(input)
	notificationDB := notification.ToDB(entities.NotificationSuccess)

	err := n.emailService.Send(ctx, notification)
	if err != nil {
		fmt.Println("Error sending notification: ", err)
		notificationDB.Status = entities.NotificationFailure
	}

	err = n.notificationTable.Put(ctx, notificationDB)
	if err != nil {
		return err
	}

	return nil
}
