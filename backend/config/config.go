package config

import (
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Load reads and parses configuration from environment
func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	viper.AutomaticEnv()

	cfg := &Config{
		App: AppConfig{
			Name:  getEnvOrDefault("APP_NAME", "pm-tool"),
			Env:   getEnvOrDefault("APP_ENV", "development"),
			Port:  getEnvAsIntOrDefault("APP_PORT", 8080),
			Debug: getEnvAsBoolOrDefault("APP_DEBUG", true),
		},
		Database: DatabaseConfig{
			Type:            getEnvOrDefault("DB_TYPE", "postgres"),
			Host:            getEnvOrDefault("DB_HOST", "localhost"),
			Port:            getEnvAsIntOrDefault("DB_PORT", 8432),
			Name:            getEnvOrDefault("DB_NAME", "pm_tool"),
			User:            getEnvOrDefault("DB_USER", "admin"),
			Password:        getEnvOrDefault("DB_PASSWORD", "example"),
			SSLMode:         getEnvOrDefault("DB_SSL_MODE", "disable"),
			MaxIdleConns:    getEnvAsIntOrDefault("DB_MAX_IDLE_CONNS", 10),
			MaxOpenConns:    getEnvAsIntOrDefault("DB_MAX_OPEN_CONNS", 100),
			ConnMaxLifetime: time.Duration(getEnvAsIntOrDefault("DB_CONN_MAX_LIFETIME", 3600)) * time.Second,
		},
		Redis: RedisConfig{
			Host:     getEnvOrDefault("REDIS_HOST", "localhost"),
			Port:     getEnvAsIntOrDefault("REDIS_PORT", 6379),
			Password: getEnvOrDefault("REDIS_PASSWORD", ""),
			DB:       getEnvAsIntOrDefault("REDIS_DB", 0),
		},
		Telemetry: TelemetryConfig{
			Enabled:      getEnvAsBoolOrDefault("TELEMETRY_ENABLED", true),
			Provider:     getEnvOrDefault("TELEMETRY_PROVIDER", "opentelemetry"),
			OTLPEndpoint: getEnvOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
			ServiceName:  getEnvOrDefault("OTEL_SERVICE_NAME", "pm-tool"),
		},
		JWT: JWTConfig{
			Secret:     getEnvOrDefault("JWT_SECRET", "your-secret-key-change-this"),
			Expiration: time.Duration(getEnvAsIntOrDefault("JWT_EXPIRATION", 24)) * time.Hour,
		},
		Log: LogConfig{
			Level:  getEnvOrDefault("LOG_LEVEL", "debug"),
			Format: getEnvOrDefault("LOG_FORMAT", "json"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvOrDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
			AllowedMethods: getEnvOrDefault("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"),
			AllowedHeaders: getEnvOrDefault("CORS_ALLOWED_HEADERS", "Content-Type,Authorization"),
		},
		RateLimit: RateLimitConfig{
			Enabled:           getEnvAsBoolOrDefault("RATE_LIMIT_ENABLED", true),
			RequestsPerSecond: getEnvAsIntOrDefault("RATE_LIMIT_REQUESTS_PER_SECOND", 100),
		},
		Oauth: OAuthConfig{
			ClientID:     getEnvOrDefault("OAUTH_CLIENT_ID", ""),
			ClientSecret: getEnvOrDefault("OAUTH_CLIENT_SECRET", ""),
			RedirectURL:  getEnvOrDefault("OAUTH_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
		},
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	viper.SetDefault(key, defaultValue)
	return viper.GetString(key)
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	viper.SetDefault(key, defaultValue)
	return viper.GetInt(key)
}

func getEnvAsBoolOrDefault(key string, defaultValue bool) bool {
	viper.SetDefault(key, defaultValue)
	return viper.GetBool(key)
}
