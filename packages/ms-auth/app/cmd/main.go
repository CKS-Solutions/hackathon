package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	http_internal "github.com/cks-solutions/hackathon/ms-auth/cmd/http"
	"github.com/cks-solutions/hackathon/ms-auth/internal/infra/database"
	"github.com/cks-solutions/hackathon/ms-auth/pkg/utils"
)

func main() {
	dbHost := utils.GetEnv("DB_HOST", "localhost")
	dbPort := utils.GetEnv("DB_PORT", "5432")
	dbUser := utils.GetEnv("DB_USER", "postgres")
	dbPassword := utils.GetEnv("DB_PASSWORD", "postgres")
	dbName := utils.GetEnv("DB_NAME", "auth_db")
	dbSSLMode := utils.GetEnv("DB_SSLMODE", "disable")

	jwtSecret := utils.GetEnv("JWT_SECRET", "your-secret-key-change-in-production")
	jwtExpirationStr := utils.GetEnv("JWT_EXPIRATION_HOURS", "24")
	jwtExpiration, err := strconv.Atoi(jwtExpirationStr)
	if err != nil {
		jwtExpiration = 24
	}

	port := utils.GetEnv("PORT", "8080")

	dbConfig := database.Config{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		DBName:   dbName,
		SSLMode:  dbSSLMode,
	}

	db, err := database.NewPostgresConnection(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := database.InitSchema(db); err != nil {
		log.Fatal("Failed to initialize schema:", err)
	}

	ctx := context.Background()
	router := http_internal.NewRouter(ctx, db, jwtSecret, jwtExpiration)

	log.Printf("🚀 Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("Server error:", err)
	}
}
