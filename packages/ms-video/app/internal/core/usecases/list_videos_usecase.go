package usecases

import (
	"context"

	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/ports"
)

type ListVideosUsecase struct {
	videoRepository ports.VideoRepository
}

func NewListVideosUsecase(videoRepository ports.VideoRepository) *ListVideosUsecase {
	return &ListVideosUsecase{
		videoRepository: videoRepository,
	}
}

func (u *ListVideosUsecase) Execute(ctx context.Context, userID string) (*dto.ListVideosOutput, error) {
	videos, err := u.videoRepository.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	videoOutputs := make([]dto.VideoOutput, len(videos))
	for i, video := range videos {
		videoOutputs[i] = dto.VideoOutput{
			ID:              video.ID,
			OriginalName:    video.OriginalName,
			Status:          string(video.Status),
			ProgressPercent: video.ProgressPercent,
			FileSize:        video.FileSize,
			ErrorMessage:    video.ErrorMessage,
			CreatedAt:       video.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:       video.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return &dto.ListVideosOutput{
		Videos: videoOutputs,
	}, nil
}
