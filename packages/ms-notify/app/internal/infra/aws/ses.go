package aws

import (
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

type SESClient struct {
	*ses.Client
}

func NewSESClient(region Region, stage Stage) *SESClient {
	return &SESClient{
		Client: ses.NewFromConfig(NewConfig(region, stage)),
	}
}
