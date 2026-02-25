package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Region string
type Stage string

func NewAWSConfig(region Region, stage Stage) aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(string(region)),
	)
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	if stage == "dev" {
		endpoint := os.Getenv("AWS_ENDPOINT_URL")
		if endpoint != "" {
			cfg.BaseEndpoint = &endpoint
		}
	}

	return cfg
}

func NewS3Client(region Region, stage Stage) *s3.Client {
	cfg := NewAWSConfig(region, stage)

	options := []func(*s3.Options){}
	if stage == "dev" {
		endpoint := os.Getenv("AWS_ENDPOINT_URL")
		if endpoint != "" {
			options = append(options, func(o *s3.Options) {
				o.UsePathStyle = true
			})
		}
	}

	return s3.NewFromConfig(cfg, options...)
}

func NewDynamoClient(region Region, stage Stage) *dynamodb.Client {
	cfg := NewAWSConfig(region, stage)
	return dynamodb.NewFromConfig(cfg)
}

func NewSQSClient(region Region, stage Stage) *sqs.Client {
	cfg := NewAWSConfig(region, stage)
	return sqs.NewFromConfig(cfg)
}

func GetTableName(baseName string, stage Stage) string {
	return fmt.Sprintf("%s-%s", baseName, stage)
}

func GetQueueName(baseName string, stage Stage) string {
	return fmt.Sprintf("%s-%s", baseName, stage)
}

func GetBucketName(baseName string, stage Stage) string {
	return fmt.Sprintf("%s-%s", baseName, stage)
}
