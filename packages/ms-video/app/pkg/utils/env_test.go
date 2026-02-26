package utils

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		shouldSetEnv bool
		expected     string
	}{
		{
			name:         "should return environment variable when set",
			key:          "TEST_VAR_1",
			defaultValue: "default",
			envValue:     "custom_value",
			shouldSetEnv: true,
			expected:     "custom_value",
		},
		{
			name:         "should return default when environment variable not set",
			key:          "TEST_VAR_2",
			defaultValue: "default_value",
			envValue:     "",
			shouldSetEnv: false,
			expected:     "default_value",
		},
		{
			name:         "should return empty string when env is empty and default is empty",
			key:          "TEST_VAR_3",
			defaultValue: "",
			envValue:     "",
			shouldSetEnv: false,
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variable before test
			os.Unsetenv(tt.key)

			if tt.shouldSetEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := GetEnv(tt.key, tt.defaultValue)

			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetRegion(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "should return custom region when AWS_REGION is set",
			envValue: "eu-west-1",
			expected: "eu-west-1",
		},
		{
			name:     "should return default us-east-1 when AWS_REGION is not set",
			envValue: "",
			expected: "us-east-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variable before test
			os.Unsetenv("AWS_REGION")

			if tt.envValue != "" {
				os.Setenv("AWS_REGION", tt.envValue)
				defer os.Unsetenv("AWS_REGION")
			}

			result := GetRegion()

			if result != tt.expected {
				t.Errorf("expected region '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetStage(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "should return custom stage when STAGE is set",
			envValue: "production",
			expected: "production",
		},
		{
			name:     "should return staging stage when STAGE is set to staging",
			envValue: "staging",
			expected: "staging",
		},
		{
			name:     "should return default dev when STAGE is not set",
			envValue: "",
			expected: "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variable before test
			os.Unsetenv("STAGE")

			if tt.envValue != "" {
				os.Setenv("STAGE", tt.envValue)
				defer os.Unsetenv("STAGE")
			}

			result := GetStage()

			if result != tt.expected {
				t.Errorf("expected stage '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetEnv_MultipleCallsSameKey(t *testing.T) {
	key := "TEST_MULTI_VAR"
	defaultValue := "default"

	os.Unsetenv(key)

	// First call without env var
	result1 := GetEnv(key, defaultValue)
	if result1 != defaultValue {
		t.Errorf("first call: expected '%s', got '%s'", defaultValue, result1)
	}

	// Set env var
	os.Setenv(key, "new_value")
	defer os.Unsetenv(key)

	// Second call with env var
	result2 := GetEnv(key, defaultValue)
	if result2 != "new_value" {
		t.Errorf("second call: expected 'new_value', got '%s'", result2)
	}
}
