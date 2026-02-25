package controller

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driver/middleware"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/usecases"
	"github.com/cks-solutions/hackathon/ms-video/pkg/utils"
)

type VideoController struct {
	uploadUsecase   *usecases.UploadVideoUsecase
	listUsecase     *usecases.ListVideosUsecase
	downloadUsecase *usecases.DownloadVideoUsecase
}

func NewVideoController(
	uploadUsecase *usecases.UploadVideoUsecase,
	listUsecase *usecases.ListVideosUsecase,
	downloadUsecase *usecases.DownloadVideoUsecase,
) *VideoController {
	return &VideoController{
		uploadUsecase:   uploadUsecase,
		listUsecase:     listUsecase,
		downloadUsecase: downloadUsecase,
	}
}

func (c *VideoController) Upload(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return utils.NewHttpError(http.StatusMethodNotAllowed, "method not allowed")
	}

	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	userEmail, err := middleware.GetEmailFromContext(ctx)
	if err != nil {
		return err
	}

	if err := r.ParseMultipartForm(500 << 20); err != nil {
		return utils.NewBadRequestError("failed to parse multipart form")
	}

	file, fileHeader, err := r.FormFile("video")
	if err != nil {
		return utils.NewBadRequestError("missing video file")
	}
	defer file.Close()

	input := dto.UploadVideoInput{
		File:      fileHeader,
		UserID:    userID,
		UserEmail: userEmail,
	}

	result, err := c.uploadUsecase.Execute(ctx, input)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(result)
}

func (c *VideoController) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return utils.NewHttpError(http.StatusMethodNotAllowed, "method not allowed")
	}

	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	result, err := c.listUsecase.Execute(ctx, userID)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(result)
}

func (c *VideoController) Download(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return utils.NewHttpError(http.StatusMethodNotAllowed, "method not allowed")
	}

	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	videoID := r.URL.Query().Get("id")
	if videoID == "" {
		return utils.NewBadRequestError("missing video id parameter")
	}

	result, err := c.downloadUsecase.Execute(ctx, videoID, userID)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(result)
}
