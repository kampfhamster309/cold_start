// Package config loads app-backend configuration from the environment,
// matching the self-hosted docker-compose deployment target (tech-stack §8).
package config

import "os"

type Config struct {
	// Port is the TCP port the HTTP server listens on.
	Port string
}

func Load() Config {
	return Config{
		Port: getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
