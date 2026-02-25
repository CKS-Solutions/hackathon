package usecases

import (
	"context"
	"fmt"

	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/entities"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/ports"
	"github.com/cks-solutions/hackathon/ms-video/pkg/utils"
)

type DownloadVideoUsecase struct {
	videoRepository ports.VideoRepository
	storageService  ports.StorageService
}

func NewDownloadVideoUsecase(
	videoRepository ports.VideoRepository,
	storageService ports.StorageService,
) *DownloadVideoUsecase {
	return &DownloadVideoUsecase{
		videoRepository: videoRepository,
		storageService:  storageService,
	}
}

func (u *DownloadVideoUsecase) Execute(ctx context.Context, videoID, userID string) (*dto.DownloadVideoOutput, error) {
	video, err := u.videoRepository.FindByID(ctx, videoID)
	if err != nil {
		return nil, utils.NewNotFoundError("video not found")
	}

	if video.UserID != userID {
		return nil, utils.NewUnauthorizedError("you don't have permission to download this video")
	}

	if video.Status != entities.VideoStatusCompleted {
		return nil, utils.NewBadRequestError(fmt.Sprintf("video is not ready for download. Current status: %s", video.Status))
	}

	expirationMinutes := 15
	presignedURL, err := u.storageService.GetPresignedURL(ctx, video.ProcessedS3Key, expirationMinutes)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to generate download URL")
	}

	return &dto.DownloadVideoOutput{
		PresignedURL: presignedURL,
		VideoID:      video.ID,
		FileName:     video.OriginalName + ".zip",
		ExpiresIn:    expirationMinutes * 60,
	}, nil
}
