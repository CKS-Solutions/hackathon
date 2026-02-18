package aws

import "github.com/aws/aws-sdk-go-v2/service/dynamodb"

type DynamoClient struct {
	*dynamodb.Client
}

func NewDynamoClient(region Region, stage Stage) *DynamoClient {
	return &DynamoClient{
		Client: dynamodb.NewFromConfig(NewConfig(region, stage)),
	}
}
