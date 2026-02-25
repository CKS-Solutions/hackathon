package utils

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	const key = "TEST_GETENV_KEY_UNIQUE"
	defer os.Unsetenv(key)

	t.Run("key set", func(t *testing.T) {
		os.Setenv(key, "set-value")
		got := GetEnv(key, "default")
		if got != "set-value" {
			t.Errorf("GetEnv(%q, \"default\") = %q, want \"set-value\"", key, got)
		}
	})

	t.Run("key unset", func(t *testing.T) {
		os.Unsetenv(key)
		got := GetEnv(key, "default")
		if got != "default" {
			t.Errorf("GetEnv(%q, \"default\") = %q, want \"default\"", key, got)
		}
	})

	t.Run("key empty", func(t *testing.T) {
		os.Setenv(key, "")
		got := GetEnv(key, "default")
		if got != "default" {
			t.Errorf("GetEnv(%q, \"default\") with empty value = %q, want \"default\"", key, got)
		}
	})
}
