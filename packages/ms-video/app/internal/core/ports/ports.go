package ports

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/entities"
)

type VideoRepository interface {
	Save(ctx context.Context, video *entities.Video) error
	FindByID(ctx context.Context, videoID string) (*entities.Video, error)
	FindByUserID(ctx context.Context, userID string) ([]*entities.Video, error)
	Update(ctx context.Context, video *entities.Video) error
}

type VideoQueue interface {
	Send(ctx context.Context, message dto.VideoProcessMessage) error
	Get(ctx context.Context) ([]types.Message, error)
	Delete(ctx context.Context, message types.Message) error
}

type StorageService interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) error
	Download(ctx context.Context, key string) ([]byte, error)
	GetPresignedURL(ctx context.Context, key string, expirationMinutes int) (string, error)
	Delete(ctx context.Context, key string) error
}

type TokenService interface {
	Validate(tokenString string) (*TokenClaims, error)
}

type TokenClaims struct {
	UserID string
	Email  string
}

type NotificationService interface {
	SendVideoProcessedNotification(ctx context.Context, email, videoID, originalName string) error
	SendVideoFailedNotification(ctx context.Context, email, videoID, originalName, errorMessage string) error
}
