module github.com/cks-solutions/hackathon/ms-auth

go 1.23

require (
	github.com/aws/aws-sdk-go-v2 v1.30.3
	github.com/aws/aws-sdk-go-v2/config v1.27.27
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.32.4
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	golang.org/x/crypto v0.31.0
)

require github.com/DATA-DOG/go-sqlmock v1.5.0 // indirect
