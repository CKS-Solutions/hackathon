package ports

import (
	"context"

	"github.com/cks-solutions/hackathon/ms-notify/internal/core/entities"
)

type NotificationTable interface {
	Put(ctx context.Context, notification entities.NotificationDB) error
}
