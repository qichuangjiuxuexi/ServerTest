package config

import (
	"os"
	"strconv"
)

// Config 存储服务器配置
type Config struct {
	Port   int
	JWTKey string
}

var cfg *Config

// GetConfig 返回配置，如果未初始化则加载配置
func GetConfig() *Config {
	if cfg == nil {
		cfg = loadConfig()
	}
	return cfg
}

// loadConfig 从环境变量加载配置
func loadConfig() *Config {
	port := 12138
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	return &Config{
		Port:   port,
		JWTKey: getEnvOrDefault("JWT_SECRET", "your-jwt-secret-key"),
	}
}

// getEnvOrDefault 获取环境变量，如不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
