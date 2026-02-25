package utils

import "os"

func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetRegion() string {
	return GetEnv("AWS_REGION", "us-east-1")
}

func GetStage() string {
	return GetEnv("STAGE", "dev")
}
