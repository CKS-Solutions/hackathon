package utils

import (
	"os"
	"testing"

	"github.com/cks-solutions/hackathon/ms-notify/internal/infra/aws"
)

func TestGetRegion(t *testing.T) {
	const key = "AWS_REGION"
	restore := os.Getenv(key)
	defer func() { os.Setenv(key, restore) }()

	t.Run("set", func(t *testing.T) {
		os.Setenv(key, "us-west-2")
		got := GetRegion()
		if got != aws.Region("us-west-2") {
			t.Errorf("GetRegion() = %q, want us-west-2", got)
		}
	})

	t.Run("unset", func(t *testing.T) {
		os.Unsetenv(key)
		got := GetRegion()
		if got != aws.REGION_US_EAST_1 {
			t.Errorf("GetRegion() = %q, want %q", got, aws.REGION_US_EAST_1)
		}
	})

	t.Run("empty", func(t *testing.T) {
		os.Setenv(key, "")
		got := GetRegion()
		if got != aws.REGION_US_EAST_1 {
			t.Errorf("GetRegion() = %q, want %q", got, aws.REGION_US_EAST_1)
		}
	})
}

func TestGetStage(t *testing.T) {
	const key = "AWS_STAGE"
	restore := os.Getenv(key)
	defer func() { os.Setenv(key, restore) }()

	t.Run("set", func(t *testing.T) {
		os.Setenv(key, "api")
		got := GetStage()
		if got != aws.STAGE_PROD {
			t.Errorf("GetStage() = %q, want api", got)
		}
	})

	t.Run("unset", func(t *testing.T) {
		os.Unsetenv(key)
		got := GetStage()
		if got != aws.STAGE_LOCAL {
			t.Errorf("GetStage() = %q, want %q", got, aws.STAGE_LOCAL)
		}
	})

	t.Run("empty", func(t *testing.T) {
		os.Setenv(key, "")
		got := GetStage()
		if got != aws.STAGE_LOCAL {
			t.Errorf("GetStage() = %q, want %q", got, aws.STAGE_LOCAL)
		}
	})
}
