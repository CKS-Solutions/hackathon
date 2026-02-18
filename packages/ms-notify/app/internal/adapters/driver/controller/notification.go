package controller

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/cks-solutions/hackathon/ms-notify/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/usecases"
	"github.com/cks-solutions/hackathon/ms-notify/pkg/utils"
)

type NotificationController interface {
	Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error
}

type NotificationControllerImpl struct {
	Usecase usecases.NotificationProducerUsecase
}

func NewNotificationController(usecase usecases.NotificationProducerUsecase) NotificationController {
	return &NotificationControllerImpl{
		Usecase: usecase,
	}
}

func (c *NotificationControllerImpl) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	input := dto.NotificationInput{}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return utils.HTTPBadRequest("invalid input")
	}

	err = c.Usecase.Run(ctx, input)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Notification request accepted",
	})
	return nil
}
