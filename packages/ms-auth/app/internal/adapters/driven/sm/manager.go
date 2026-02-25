package sm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type SecretsManagerService struct {
	client *secretsmanager.Client
}

type DBCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DBName   string `json:"dbname"`
}

type JWTSecret struct {
	JWTSecret string `json:"jwt_secret"`
}

func NewSecretsManagerService(cfg aws.Config) *SecretsManagerService {
	return &SecretsManagerService{
		client: secretsmanager.NewFromConfig(cfg),
	}
}

func (s *SecretsManagerService) GetDBCredentials(ctx context.Context, secretName string) (*DBCredentials, error) {
	result, err := s.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret value: %w", err)
	}

	if result.SecretString == nil {
		return nil, fmt.Errorf("secret string is nil")
	}

	var credentials DBCredentials
	if err := json.Unmarshal([]byte(*result.SecretString), &credentials); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret: %w", err)
	}

	return &credentials, nil
}

func (s *SecretsManagerService) GetJWTSecret(ctx context.Context, secretName string) (string, error) {
	result, err := s.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get secret value: %w", err)
	}

	if result.SecretString == nil {
		return "", fmt.Errorf("secret string is nil")
	}

	var jwtSecret JWTSecret
	if err := json.Unmarshal([]byte(*result.SecretString), &jwtSecret); err != nil {
		return "", fmt.Errorf("failed to unmarshal secret: %w", err)
	}

	return jwtSecret.JWTSecret, nil
}
