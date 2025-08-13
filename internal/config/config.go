/*
 * Copyright (C) 2025 Miguel Mamani <miguel.coder.per@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>.
 */

package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"highperf-api/internal/logger"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
	Auth     AuthConfig     `json:"auth"`
	Logger   logger.Config  `json:"logger"`
	Metrics  MetricsConfig  `json:"metrics"`
}

type ServerConfig struct {
	Host              string        `json:"host" envconfig:"SERVER_HOST" default:"0.0.0.0"`
	Port              int           `json:"port" envconfig:"PORT" default:"8080"`
	ReadTimeout       time.Duration `json:"read_timeout" envconfig:"READ_TIMEOUT" default:"5s"`
	WriteTimeout      time.Duration `json:"write_timeout" envconfig:"WRITE_TIMEOUT" default:"10s"`
	IdleTimeout       time.Duration `json:"idle_timeout" envconfig:"IDLE_TIMEOUT" default:"60s"`
	ReadHeaderTimeout time.Duration `json:"read_header_timeout" envconfig:"READ_HEADER_TIMEOUT" default:"2s"`
	MaxHeaderBytes    int           `json:"max_header_bytes" envconfig:"MAX_HEADER_BYTES" default:"8192"`
	GracefulTimeout   time.Duration `json:"graceful_timeout" envconfig:"GRACEFUL_TIMEOUT" default:"15s"`
}

type DatabaseConfig struct {
	Driver          string        `json:"driver" envconfig:"DB_DRIVER" default:"postgres"`
	Host            string        `json:"host" envconfig:"DB_HOST" default:"localhost"`
	Port            int           `json:"port" envconfig:"DB_PORT" default:"5432"`
	Name            string        `json:"name" envconfig:"DB_NAME" default:"api_db"`
	User            string        `json:"user" envconfig:"DB_USER" default:"postgres"`
	Password        string        `json:"password" envconfig:"DB_PASSWORD" default:""`
	SSLMode         string        `json:"ssl_mode" envconfig:"DB_SSL_MODE" default:"disable"`
	MaxOpenConns    int           `json:"max_open_conns" envconfig:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns    int           `json:"max_idle_conns" envconfig:"DB_MAX_IDLE_CONNS" default:"25"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" envconfig:"DB_CONN_MAX_LIFETIME" default:"5m"`
}

type RedisConfig struct {
	Host     string `json:"host" envconfig:"REDIS_HOST" default:"localhost"`
	Port     int    `json:"port" envconfig:"REDIS_PORT" default:"6379"`
	Password string `json:"password" envconfig:"REDIS_PASSWORD" default:""`
	DB       int    `json:"db" envconfig:"REDIS_DB" default:"0"`
}

type AuthConfig struct {
	JWTSecret     string        `json:"jwt_secret" envconfig:"JWT_SECRET" default:"your-secret-key"`
	TokenExpiry   time.Duration `json:"token_expiry" envconfig:"TOKEN_EXPIRY" default:"24h"`
	RefreshExpiry time.Duration `json:"refresh_expiry" envconfig:"REFRESH_EXPIRY" default:"168h"` // 7 days
}

type MetricsConfig struct {
	Enabled bool   `json:"enabled" envconfig:"METRICS_ENABLED" default:"true"`
	Port    int    `json:"port" envconfig:"METRICS_PORT" default:"9090"`
	Path    string `json:"path" envconfig:"METRICS_PATH" default:"/metrics"`
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:              getEnvString("SERVER_HOST", "0.0.0.0"),
			Port:              getEnvInt("PORT", 8080),
			ReadTimeout:       getEnvDuration("READ_TIMEOUT", 5*time.Second),
			WriteTimeout:      getEnvDuration("WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:       getEnvDuration("IDLE_TIMEOUT", 60*time.Second),
			ReadHeaderTimeout: getEnvDuration("READ_HEADER_TIMEOUT", 2*time.Second),
			MaxHeaderBytes:    getEnvInt("MAX_HEADER_BYTES", 8192),
			GracefulTimeout:   getEnvDuration("GRACEFUL_TIMEOUT", 15*time.Second),
		},
		Database: DatabaseConfig{
			Driver:          getEnvString("DB_DRIVER", "postgres"),
			Host:            getEnvString("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			Name:            getEnvString("DB_NAME", "api_db"),
			User:            getEnvString("DB_USER", "postgres"),
			Password:        getEnvString("DB_PASSWORD", ""),
			SSLMode:         getEnvString("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnvString("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnvString("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		Auth: AuthConfig{
			JWTSecret:     getEnvString("JWT_SECRET", "your-secret-key-change-in-production"),
			TokenExpiry:   getEnvDuration("TOKEN_EXPIRY", 24*time.Hour),
			RefreshExpiry: getEnvDuration("REFRESH_EXPIRY", 168*time.Hour),
		},
		Logger: logger.Config{
			Level:     getEnvString("LOG_LEVEL", "info"),
			Format:    getEnvString("LOG_FORMAT", "json"),
			AddSource: getEnvBool("LOG_ADD_SOURCE", true),
		},
		Metrics: MetricsConfig{
			Enabled: getEnvBool("METRICS_ENABLED", true),
			Port:    getEnvInt("METRICS_PORT", 9090),
			Path:    getEnvString("METRICS_PATH", "/metrics"),
		},
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Database.Driver == "" {
		return fmt.Errorf("database driver is required")
	}

	if c.Auth.JWTSecret == "" || c.Auth.JWTSecret == "your-secret-key-change-in-production" {
		return fmt.Errorf("JWT secret must be set and not use default value")
	}

	return nil
}

// DatabaseURL returns the database connection URL
func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.Driver,
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// ServerAddr returns the server address
func (c *Config) ServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// RedisAddr returns the Redis address
func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// Helper functions for environment variable parsing
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}