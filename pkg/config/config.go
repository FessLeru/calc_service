// pkg/config/config.go
package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	ServerAddress  string
	ComputingPower int
}

func LoadConfig() *Config {
	return &Config{
		ServerAddress:  getEnv("SERVER_ADDRESS", ":8080"),
		ComputingPower: mustAtoi(getEnv("COMPUTING_POWER", "4")),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func mustAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("invalid integer value: %v", err)
	}
	return i
}
