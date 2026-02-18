package aws

import "github.com/aws/aws-sdk-go-v2/service/sqs"

type SQSClient struct {
	*sqs.Client
}

func NewSQSClient(region Region, stage Stage) *SQSClient {
	return &SQSClient{
		Client: sqs.NewFromConfig(NewConfig(region, stage)),
	}
}
