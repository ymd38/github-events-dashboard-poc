package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	defaultBackendPort = 8080
	defaultMySQLPort   = 3306
	defaultMySQLHost   = "db"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	GitHubClientID     string
	GitHubClientSecret string
	GitHubWebhookSecret string
	MySQLHost          string
	MySQLPort          int
	MySQLUser          string
	MySQLPassword      string
	MySQLDatabase      string
	BackendPort        int
	FrontendURL        string
	SessionSecret      string
}

// DSN returns the MySQL Data Source Name for database/sql connection.
func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4",
		c.MySQLUser, c.MySQLPassword, c.MySQLHost, c.MySQLPort, c.MySQLDatabase)
}

// Load reads configuration from environment variables and returns a Config.
func Load() (*Config, error) {
	cfg := &Config{
		GitHubClientID:      getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret:  getEnv("GITHUB_CLIENT_SECRET", ""),
		GitHubWebhookSecret: getEnv("GITHUB_WEBHOOK_SECRET", ""),
		MySQLHost:           getEnv("MYSQL_HOST", defaultMySQLHost),
		MySQLUser:           getEnv("MYSQL_USER", ""),
		MySQLPassword:       getEnv("MYSQL_PASSWORD", ""),
		MySQLDatabase:       getEnv("MYSQL_DATABASE", ""),
		FrontendURL:         getEnv("FRONTEND_URL", "http://localhost:3000"),
		SessionSecret:       getEnv("SESSION_SECRET", ""),
	}
	port, err := strconv.Atoi(getEnv("BACKEND_PORT", strconv.Itoa(defaultBackendPort)))
	if err != nil {
		return nil, fmt.Errorf("invalid BACKEND_PORT: %w", err)
	}
	cfg.BackendPort = port
	mysqlPort, err := strconv.Atoi(getEnv("MYSQL_PORT", strconv.Itoa(defaultMySQLPort)))
	if err != nil {
		return nil, fmt.Errorf("invalid MYSQL_PORT: %w", err)
	}
	cfg.MySQLPort = mysqlPort
	if cfg.MySQLUser == "" || cfg.MySQLPassword == "" || cfg.MySQLDatabase == "" {
		return nil, fmt.Errorf("MYSQL_USER, MYSQL_PASSWORD, and MYSQL_DATABASE are required")
	}
	return cfg, nil
}

func getEnv(key string, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
