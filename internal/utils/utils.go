package utils

import (
	"os"
	"strconv"
)

func EnvOrString(envKey string, constant string) string {
	if val, ok := os.LookupEnv(envKey); ok {
		return val
	}

	return constant
}

func EnvOrInt(envKey string, constant int) int {
	if val, ok := os.LookupEnv(envKey); ok {
		v, err := strconv.Atoi(val)

		if err != nil {
			// Fallback to default if unparsable env
			return constant
		}

		return v
	}

	return constant
}
