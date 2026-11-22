package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr               string
	MongoURI               string
	MongoDB                string
	RedisAddr              string
	RedisPassword          string
	RedisTTLSeconds        int
	ShutdownTimeoutSeconds int
	LogLevel               string
}

func Load() *Config {
	return &Config{
		HTTPAddr:               getEnv("HTTP_ADDR", ":8080"),
		MongoURI:               getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:                getEnv("MONGO_DB", "hello_world"),
		RedisAddr:              getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:          getEnv("REDIS_PASSWORD", ""),
		RedisTTLSeconds:        getEnvAsInt("REDIS_TTL_SECONDS", 300),
		ShutdownTimeoutSeconds: getEnvAsInt("SHUTDOWN_TIMEOUT_SECONDS", 10),
		LogLevel:               getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}
