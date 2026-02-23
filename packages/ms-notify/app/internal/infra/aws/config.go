package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
)

const (
	LOCALSTACK_ENDPOINT = "http://localstack:4566"
)

type Stage string

const (
	STAGE_LOCAL Stage = "local"
	STAGE_PROD  Stage = "api"
)

type Region string

const (
	REGION_US_EAST_1 Region = "us-east-1"
)

// loadLocal returns a config for LocalStack (sem credential chain).
func loadLocal(region Region) aws.Config {
	return aws.Config{
		BaseEndpoint: aws.String(LOCALSTACK_ENDPOINT),
		Region:       string(region),
	}
}

// loadProd loads the default credential chain (IRSA, env, instance metadata).
func loadProd(ctx context.Context, region Region) (aws.Config, error) {
	return awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(string(region)))
}

// NewConfig returns AWS config. Em prod usa LoadDefaultConfig (IRSA/env/IMDS). Em local usa endpoint LocalStack.
func NewConfig(region Region, stage Stage) aws.Config {
	if stage == STAGE_LOCAL {
		return loadLocal(region)
	}
	cfg, err := loadProd(context.Background(), region)
	if err != nil {
		// Fallback mínimo para não quebrar; em EKS o loadProd não deve falhar
		return aws.Config{Region: string(region)}
	}
	return cfg
}
