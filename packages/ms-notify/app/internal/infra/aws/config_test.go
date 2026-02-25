package aws

import (
	"testing"
)

func TestNewConfig_STAGE_LOCAL(t *testing.T) {
	cfg := NewConfig(REGION_US_EAST_1, STAGE_LOCAL)
	if cfg.Region != string(REGION_US_EAST_1) {
		t.Errorf("Region = %q, want %q", cfg.Region, REGION_US_EAST_1)
	}
	if cfg.BaseEndpoint == nil || *cfg.BaseEndpoint != LOCALSTACK_ENDPOINT {
		t.Errorf("BaseEndpoint = %v, want %q", cfg.BaseEndpoint, LOCALSTACK_ENDPOINT)
	}
}

func TestNewConfig_STAGE_PROD(t *testing.T) {
	// May call loadProd (can fail without creds); then fallback returns config with region only.
	cfg := NewConfig(REGION_US_EAST_1, STAGE_PROD)
	if cfg.Region != string(REGION_US_EAST_1) {
		t.Errorf("Region = %q, want %q", cfg.Region, REGION_US_EAST_1)
	}
}

func TestLoadLocal(t *testing.T) {
	cfg := loadLocal(REGION_US_EAST_1)
	if cfg.Region != string(REGION_US_EAST_1) {
		t.Errorf("Region = %q", cfg.Region)
	}
	if cfg.BaseEndpoint == nil || *cfg.BaseEndpoint != LOCALSTACK_ENDPOINT {
		t.Errorf("BaseEndpoint = %v", cfg.BaseEndpoint)
	}
}
