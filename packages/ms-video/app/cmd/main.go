package main

import (
	"context"
	"log"
	"net/http"

	http_internal "github.com/cks-solutions/hackathon/ms-video/cmd/http"
	sqs_internal "github.com/cks-solutions/hackathon/ms-video/cmd/sqs"
	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driven/sm"
	awsinfra "github.com/cks-solutions/hackathon/ms-video/internal/infra/aws"
	"github.com/cks-solutions/hackathon/ms-video/pkg/utils"
)

func main() {
	ctx := context.TODO()
	
	region := awsinfra.Region(utils.GetRegion())
	stage := awsinfra.Stage(utils.GetStage())
	port := utils.GetEnv("PORT", "8080")
	
	var jwtSecret string
	
	if stage == awsinfra.STAGE_PROD {
		log.Println("🔐 Loading JWT secret from AWS Secrets Manager...")
		
		awsConfig := awsinfra.NewConfig(region, stage)
		secretsService := sm.NewSecretsManagerService(awsConfig)
		
		jwtSecretName := utils.GetEnv("JWT_SECRET_NAME", "hackathon-prod-jwt-secret")
		var err error
		jwtSecret, err = secretsService.GetJWTSecret(ctx, jwtSecretName)
		if err != nil {
			log.Fatal("Failed to get JWT secret from Secrets Manager:", err)
		}
		
		log.Println("✅ JWT secret loaded from Secrets Manager")
	} else {
		log.Println("🔧 Loading JWT secret from environment variables...")
		jwtSecret = utils.GetEnv("JWT_SECRET", "your-secret-key-change-in-production")
	}

	router := http_internal.NewRouter(ctx, region, stage, jwtSecret)

	consumer := sqs_internal.NewSQSConsumer(ctx, region, stage)
	go consumer.Start()

	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("📦 Stage: %s, Region: %s", stage, region)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("Server error:", err)
	}
}
