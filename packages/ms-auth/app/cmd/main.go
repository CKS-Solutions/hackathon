package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	http_internal "github.com/cks-solutions/hackathon/ms-auth/cmd/http"
	"github.com/cks-solutions/hackathon/ms-auth/internal/adapters/driven/sm"
	awsinfra "github.com/cks-solutions/hackathon/ms-auth/internal/infra/aws"
	"github.com/cks-solutions/hackathon/ms-auth/internal/infra/database"
	"github.com/cks-solutions/hackathon/ms-auth/pkg/utils"
)

func main() {
	ctx := context.Background()
	
	stage := awsinfra.Stage(utils.GetEnv("STAGE", "local"))
	region := awsinfra.Region(utils.GetEnv("AWS_REGION", "us-east-1"))
	
	var dbConfig database.Config
	var jwtSecret string
	
	if stage == awsinfra.STAGE_PROD {
		log.Println("🔐 Loading credentials from AWS Secrets Manager...")
		
		awsConfig := awsinfra.NewConfig(region, stage)
		secretsService := sm.NewSecretsManagerService(awsConfig)
		
		// Load database credentials
		dbSecretName := utils.GetEnv("DB_SECRET_NAME", "hackathon-prod-ms-auth-db-password")
		credentials, err := secretsService.GetDBCredentials(ctx, dbSecretName)
		if err != nil {
			log.Fatal("Failed to get DB credentials from Secrets Manager:", err)
		}
		
		dbConfig = database.Config{
			Host:     credentials.Host,
			Port:     fmt.Sprintf("%d", credentials.Port),
			User:     credentials.Username,
			Password: credentials.Password,
			DBName:   credentials.DBName,
			SSLMode:  utils.GetEnv("DB_SSLMODE", "require"),
		}
		
		log.Println("✅ Database credentials loaded from Secrets Manager")
		
		// Load JWT secret
		jwtSecretName := utils.GetEnv("JWT_SECRET_NAME", "hackathon-prod-jwt-secret")
		jwtSecret, err = secretsService.GetJWTSecret(ctx, jwtSecretName)
		if err != nil {
			log.Fatal("Failed to get JWT secret from Secrets Manager:", err)
		}
		
		log.Println("✅ JWT secret loaded from Secrets Manager")
	} else {
		log.Println("🔧 Loading credentials from environment variables...")
		
		dbConfig = database.Config{
			Host:     utils.GetEnv("DB_HOST", "localhost"),
			Port:     utils.GetEnv("DB_PORT", "5432"),
			User:     utils.GetEnv("DB_USER", "postgres"),
			Password: utils.GetEnv("DB_PASSWORD", "postgres"),
			DBName:   utils.GetEnv("DB_NAME", "auth_db"),
			SSLMode:  utils.GetEnv("DB_SSLMODE", "disable"),
		}
		
		jwtSecret = utils.GetEnv("JWT_SECRET", "your-secret-key-change-in-production")
	}

	jwtExpirationStr := utils.GetEnv("JWT_EXPIRATION_HOURS", "24")
	jwtExpiration, err := strconv.Atoi(jwtExpirationStr)
	if err != nil {
		jwtExpiration = 24
	}

	port := utils.GetEnv("PORT", "8080")

	db, err := database.NewPostgresConnection(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := database.InitSchema(db); err != nil {
		log.Fatal("Failed to initialize schema:", err)
	}

	router := http_internal.NewRouter(ctx, db, jwtSecret, jwtExpiration)

	log.Printf("🚀 Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("Server error:", err)
	}
}
