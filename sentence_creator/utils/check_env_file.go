package utils

import (
	"errors"
	"fmt"
	"os"
)

func GetEnvFileKey(envKey string) string {
	envValue := os.Getenv(envKey)
	if envValue == "" {
		panic(fmt.Sprintf("no env value for %s", envKey))
	}
	if _, err := os.Stat(envValue); os.IsNotExist(err) {
		panic(errors.New(fmt.Sprintf("env value for %s is provided (%s) but no file found", envKey, envValue)))
	}
	return envValue
}
