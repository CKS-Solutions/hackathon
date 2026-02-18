package aws

import "github.com/aws/aws-sdk-go-v2/aws"

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

func load(region Region) aws.Config {
	return aws.Config{
		Region: string(region),
	}
}

func loadLocal(region Region) aws.Config {
	return aws.Config{
		BaseEndpoint: aws.String(LOCALSTACK_ENDPOINT),
		Region:       string(region),
	}
}

func NewConfig(region Region, stage Stage) aws.Config {
	if stage == STAGE_LOCAL {
		return loadLocal(region)
	}

	return load(region)
}
