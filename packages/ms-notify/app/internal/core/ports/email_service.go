package ports

import (
	"context"

	"github.com/cks-solutions/hackathon/ms-notify/internal/core/entities"
)

type EmailService interface {
	Send(ctx context.Context, notification entities.Notification) error
}
