package dynamo

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/entities"
)

type fakeDynamoClient struct {
	putItemErr error
}

func (f *fakeDynamoClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	if f.putItemErr != nil {
		return nil, f.putItemErr
	}
	return &dynamodb.PutItemOutput{}, nil
}

func TestNotificationTableImpl_Put(t *testing.T) {
	ctx := context.Background()
	notification := entities.NotificationDB{
		Id: "1", Subject: "s", From: "f@x.com", To: []string{"t@x.com"}, Html: "h", Status: entities.NotificationSuccess,
	}

	t.Run("success", func(t *testing.T) {
		table := NewNotificationTable(&fakeDynamoClient{}).(*NotificationTableImpl)
		err := table.Put(ctx, notification)
		if err != nil {
			t.Errorf("Put: %v", err)
		}
	})

	t.Run("PutItem error", func(t *testing.T) {
		table := NewNotificationTable(&fakeDynamoClient{putItemErr: errors.New("dynamo error")}).(*NotificationTableImpl)
		err := table.Put(ctx, notification)
		if err == nil {
			t.Error("expected error")
		}
	})
}
