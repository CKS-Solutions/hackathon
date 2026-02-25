package main

import (
	"context"
	"log"
	"net/http"

	http_internal "github.com/cks-solutions/hackathon/ms-video/cmd/http"
	sqs_internal "github.com/cks-solutions/hackathon/ms-video/cmd/sqs"
	awsinfra "github.com/cks-solutions/hackathon/ms-video/internal/infra/aws"
	"github.com/cks-solutions/hackathon/ms-video/pkg/utils"
)

func main() {
	region := awsinfra.Region(utils.GetRegion())
	stage := awsinfra.Stage(utils.GetStage())
	jwtSecret := utils.GetEnv("JWT_SECRET", "your-secret-key-change-in-production")
	port := utils.GetEnv("PORT", "8080")

	ctx := context.TODO()

	router := http_internal.NewRouter(ctx, region, stage, jwtSecret)

	consumer := sqs_internal.NewSQSConsumer(ctx, region, stage)
	go consumer.Start()

	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("📦 Stage: %s, Region: %s", stage, region)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("Server error:", err)
	}
}
