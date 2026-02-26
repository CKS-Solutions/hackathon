package mocks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/entities"
)

// MockVideoRepository is a mock implementation of VideoRepository interface
type MockVideoRepository struct {
	SaveFunc       func(ctx context.Context, video *entities.Video) error
	FindByIDFunc   func(ctx context.Context, videoID string) (*entities.Video, error)
	FindByUserIDFunc func(ctx context.Context, userID string) ([]*entities.Video, error)
	UpdateFunc     func(ctx context.Context, video *entities.Video) error
}

func (m *MockVideoRepository) Save(ctx context.Context, video *entities.Video) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, video)
	}
	return nil
}

func (m *MockVideoRepository) FindByID(ctx context.Context, videoID string) (*entities.Video, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, videoID)
	}
	return nil, nil
}

func (m *MockVideoRepository) FindByUserID(ctx context.Context, userID string) ([]*entities.Video, error) {
	if m.FindByUserIDFunc != nil {
		return m.FindByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockVideoRepository) Update(ctx context.Context, video *entities.Video) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, video)
	}
	return nil
}

// MockStorageService is a mock implementation of StorageService interface
type MockStorageService struct {
	UploadFunc          func(ctx context.Context, key string, data []byte, contentType string) error
	DownloadFunc        func(ctx context.Context, key string) ([]byte, error)
	GetPresignedURLFunc func(ctx context.Context, key string, expirationMinutes int) (string, error)
	DeleteFunc          func(ctx context.Context, key string) error
}

func (m *MockStorageService) Upload(ctx context.Context, key string, data []byte, contentType string) error {
	if m.UploadFunc != nil {
		return m.UploadFunc(ctx, key, data, contentType)
	}
	return nil
}

func (m *MockStorageService) Download(ctx context.Context, key string) ([]byte, error) {
	if m.DownloadFunc != nil {
		return m.DownloadFunc(ctx, key)
	}
	return nil, nil
}

func (m *MockStorageService) GetPresignedURL(ctx context.Context, key string, expirationMinutes int) (string, error) {
	if m.GetPresignedURLFunc != nil {
		return m.GetPresignedURLFunc(ctx, key, expirationMinutes)
	}
	return "", nil
}

func (m *MockStorageService) Delete(ctx context.Context, key string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, key)
	}
	return nil
}

// MockVideoQueue is a mock implementation of VideoQueue interface
type MockVideoQueue struct {
	SendFunc   func(ctx context.Context, message dto.VideoProcessMessage) error
	GetFunc    func(ctx context.Context) ([]types.Message, error)
	DeleteFunc func(ctx context.Context, message types.Message) error
}

func (m *MockVideoQueue) Send(ctx context.Context, message dto.VideoProcessMessage) error {
	if m.SendFunc != nil {
		return m.SendFunc(ctx, message)
	}
	return nil
}

func (m *MockVideoQueue) Get(ctx context.Context) ([]types.Message, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx)
	}
	return nil, nil
}

func (m *MockVideoQueue) Delete(ctx context.Context, message types.Message) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, message)
	}
	return nil
}

// MockNotificationService is a mock implementation of NotificationService interface
type MockNotificationService struct {
	SendVideoProcessedNotificationFunc func(ctx context.Context, email, videoID, originalName string) error
	SendVideoFailedNotificationFunc    func(ctx context.Context, email, videoID, originalName, errorMessage string) error
}

func (m *MockNotificationService) SendVideoProcessedNotification(ctx context.Context, email, videoID, originalName string) error {
	if m.SendVideoProcessedNotificationFunc != nil {
		return m.SendVideoProcessedNotificationFunc(ctx, email, videoID, originalName)
	}
	return nil
}

func (m *MockNotificationService) SendVideoFailedNotification(ctx context.Context, email, videoID, originalName, errorMessage string) error {
	if m.SendVideoFailedNotificationFunc != nil {
		return m.SendVideoFailedNotificationFunc(ctx, email, videoID, originalName, errorMessage)
	}
	return nil
}
