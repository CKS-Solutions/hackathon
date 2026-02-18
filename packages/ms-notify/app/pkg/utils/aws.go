package utils

import (
	"os"

	"github.com/cks-solutions/hackathon/ms-notify/internal/infra/aws"
)

func GetRegion() aws.Region {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		return aws.REGION_US_EAST_1
	}

	return aws.Region(region)
}

func GetStage() aws.Stage {
	stage := os.Getenv("AWS_STAGE")
	if stage == "" {
		return aws.STAGE_LOCAL
	}

	return aws.Stage(stage)
}
