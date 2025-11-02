package config

import "time"

// Config holds all application configuration
type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	Telemetry TelemetryConfig
	JWT       JWTConfig
	Log       LogConfig
	CORS      CORSConfig
	RateLimit RateLimitConfig
	Oauth     OAuthConfig
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Name  string
	Env   string
	Port  int
	Debug bool
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type            string
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// TelemetryConfig holds telemetry configuration
type TelemetryConfig struct {
	Enabled      bool
	Provider     string
	OTLPEndpoint string
	ServiceName  string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string
	Format string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins string
	AllowedMethods string
	AllowedHeaders string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool
	RequestsPerSecond int
}
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}
