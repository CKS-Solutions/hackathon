package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/cks-solutions/hackathon/ms-notify/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/entities"
)

type fakeEmailService struct {
	sendErr error
}

func (f *fakeEmailService) Send(ctx context.Context, notification entities.Notification) error {
	return f.sendErr
}

type fakeNotificationTable struct {
	putErr   error
	lastPut  *entities.NotificationDB
}

func (f *fakeNotificationTable) Put(ctx context.Context, notification entities.NotificationDB) error {
	f.lastPut = &notification
	return f.putErr
}

func TestNotificationConsumerUsecase_Run(t *testing.T) {
	ctx := context.Background()
	input := dto.NotificationInput{
		Subject: "Sub",
		To:      []string{"a@b.com"},
		Html:    "body",
	}

	t.Run("send ok put ok", func(t *testing.T) {
		table := &fakeNotificationTable{}
		uc := NewNotificationConsumerUsecase(table, &fakeEmailService{})
		err := uc.Run(ctx, input)
		if err != nil {
			t.Errorf("Run() err = %v", err)
		}
		if table.lastPut == nil || table.lastPut.Status != entities.NotificationSuccess {
			t.Errorf("expected Put with Status SUCCESS, got %v", table.lastPut)
		}
	})

	t.Run("send error then put with failure", func(t *testing.T) {
		table := &fakeNotificationTable{}
		uc := NewNotificationConsumerUsecase(table, &fakeEmailService{sendErr: errors.New("send failed")})
		err := uc.Run(ctx, input)
		if err != nil {
			t.Errorf("Run() err = %v", err)
		}
		if table.lastPut == nil || table.lastPut.Status != entities.NotificationFailure {
			t.Errorf("expected Put with Status FAILURE, got %v", table.lastPut)
		}
	})

	t.Run("put error", func(t *testing.T) {
		table := &fakeNotificationTable{putErr: errors.New("put failed")}
		uc := NewNotificationConsumerUsecase(table, &fakeEmailService{})
		err := uc.Run(ctx, input)
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "put failed" {
			t.Errorf("err = %v", err)
		}
	})
}
