package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/cks-solutions/hackathon/ms-notify/internal/adapters/driven/sqs"
	"github.com/cks-solutions/hackathon/ms-notify/internal/adapters/driver/controller"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/usecases"
	awsinfra "github.com/cks-solutions/hackathon/ms-notify/internal/infra/aws"
	"github.com/cks-solutions/hackathon/ms-notify/pkg/utils"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type Router struct {
	*http.ServeMux
	Ctx context.Context
}

type HTTPHandlerWithErr func(context.Context, http.ResponseWriter, *http.Request) error

func handleError(w http.ResponseWriter, err error) {
	var httpErr *utils.HttpError
	if errors.As(err, &httpErr) {
		w.WriteHeader(httpErr.StatusCode)
		json.NewEncoder(w).Encode(ErrorResponse{
			Message: err.Error(),
		})
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Message: err.Error(),
		})
	}
}

func NewRouter(ctx context.Context, region awsinfra.Region, stage awsinfra.Stage) *Router {
	mux := &Router{ServeMux: http.NewServeMux(), Ctx: ctx}

	sqsClient := awsinfra.NewSQSClient(region, stage)
	notificationQueue := sqs.NewNotificationQueue(*sqsClient)

	notificationProducerUsecase := usecases.NewNotificationProducerUsecase(notificationQueue)

	notificationController := controller.NewNotificationController(notificationProducerUsecase)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"ms-notify"}`))
	})

	mux.Handle("/notification", notificationController.Create)

	return mux
}

func (r *Router) Handle(path string, handler HTTPHandlerWithErr) {
	r.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
		if err := handler(r.Ctx, w, req); err != nil {
			handleError(w, err)
		}
	})
}
