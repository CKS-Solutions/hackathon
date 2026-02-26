package http

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driven/dynamodb"
	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driven/jwt"
	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driven/s3"
	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driven/sqs"
	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driver/controller"
	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driver/middleware"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/usecases"
	awsinfra "github.com/cks-solutions/hackathon/ms-video/internal/infra/aws"
	"github.com/cks-solutions/hackathon/ms-video/pkg/utils"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type Router struct {
	*http.ServeMux
	Ctx context.Context
}

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
			Message: "internal server error",
		})
		log.Println("Internal error:", err)
	}
}

func NewRouter(ctx context.Context, region awsinfra.Region, stage awsinfra.Stage, jwtSecret string) *Router {
	mux := &Router{ServeMux: http.NewServeMux(), Ctx: ctx}

	s3Client := awsinfra.NewS3Client(region, stage)
	dynamoClient := awsinfra.NewDynamoClient(region, stage)
	sqsClient := awsinfra.NewSQSClient(region, stage)

	storageService := s3.NewS3StorageService(s3Client)
	videoRepository := dynamodb.NewDynamoVideoRepository(dynamoClient)
	videoQueue := sqs.NewSQSVideoQueue(sqsClient)
	tokenService := jwt.NewTokenService(jwtSecret)

	uploadUsecase := usecases.NewUploadVideoUsecase(videoRepository, storageService, videoQueue)
	listUsecase := usecases.NewListVideosUsecase(videoRepository)
	downloadUsecase := usecases.NewDownloadVideoUsecase(videoRepository, storageService)

	videoController := controller.NewVideoController(uploadUsecase, listUsecase, downloadUsecase)

	healthResp := []byte(`{"status":"healthy","service":"ms-video"}`)
	mux.HandleFunc("/video/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(healthResp)
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(healthResp)
	})

	mux.HandleFunc("/video/upload", middleware.AuthMiddleware(tokenService, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := videoController.Upload(r.Context(), w, r); err != nil {
			handleError(w, err)
		}
	}))

	mux.HandleFunc("/video/list", middleware.AuthMiddleware(tokenService, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := videoController.List(r.Context(), w, r); err != nil {
			handleError(w, err)
		}
	}))

	mux.HandleFunc("/video/download", middleware.AuthMiddleware(tokenService, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := videoController.Download(r.Context(), w, r); err != nil {
			handleError(w, err)
		}
	}))

	return mux
}
