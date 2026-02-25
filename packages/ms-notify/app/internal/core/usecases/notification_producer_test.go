package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/cks-solutions/hackathon/ms-notify/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/entities"
)

type fakeNotificationQueue struct {
	pushErr error
}

func (f *fakeNotificationQueue) Push(ctx context.Context, message entities.Notification) error {
	return f.pushErr
}

func (f *fakeNotificationQueue) Get(ctx context.Context) ([]types.Message, error) {
	return nil, nil
}

func (f *fakeNotificationQueue) Delete(ctx context.Context, _ types.Message) error {
	return nil
}

func TestNotificationProducerUsecase_Run(t *testing.T) {
	ctx := context.Background()
	input := dto.NotificationInput{
		Subject: "Sub",
		To:      []string{"a@b.com"},
		Html:    "body",
	}

	t.Run("push error", func(t *testing.T) {
		q := &fakeNotificationQueue{pushErr: errors.New("queue error")}
		uc := NewNotificationProducerUsecase(q)
		err := uc.Run(ctx, input)
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "queue error" {
			t.Errorf("err = %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		q := &fakeNotificationQueue{}
		uc := NewNotificationProducerUsecase(q)
		err := uc.Run(ctx, input)
		if err != nil {
			t.Errorf("Run() err = %v", err)
		}
	})
}
