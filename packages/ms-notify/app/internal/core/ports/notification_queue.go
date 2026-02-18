package ports

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/entities"
)

type NotificationQueue interface {
	Push(ctx context.Context, message entities.Notification) error
	Get(ctx context.Context) ([]types.Message, error)
	Delete(ctx context.Context, message types.Message) error
}
