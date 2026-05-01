package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type HTTPConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type Config struct {
	ServiceName string
	AppEnv      string
	Port        string
	LogLevel    string
	HTTP        HTTPConfig

	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	GoogleAuthURL      string
	GoogleTokenURL     string
	GoogleUserInfoURL  string
	OAuthStateSecret   string
	OAuthStateTTL      time.Duration
	JWTSecret          string
	JWTIssuer          string
	JWTExpiresIn       time.Duration
	SessionTTL         time.Duration
	RedisHost          string
	RedisPort          string
	RedisPassword      string
	RedisDB            int
	UpstreamTimeout    time.Duration
}

func Load() Config {
	cfg := Config{
		ServiceName: getEnv("SERVICE_NAME", "service"),
		AppEnv:      getEnv("APP_ENV", "development"),
		Port:        getEnv("PORT", "8080"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		HTTP: HTTPConfig{
			ReadTimeout:  getDurationSeconds("HTTP_READ_TIMEOUT", 10),
			WriteTimeout: getDurationSeconds("HTTP_WRITE_TIMEOUT", 10),
			IdleTimeout:  getDurationSeconds("HTTP_IDLE_TIMEOUT", 60),
		},
	}

	cfg.GoogleClientID = getEnv("GOOGLE_CLIENT_ID", "")
	cfg.GoogleClientSecret = getEnv("GOOGLE_CLIENT_SECRET", "")
	cfg.GoogleRedirectURL = getEnv("GOOGLE_REDIRECT_URL", "")
	cfg.GoogleAuthURL = getEnv("GOOGLE_AUTH_URL", "https://accounts.google.com/o/oauth2/v2/auth")
	cfg.GoogleTokenURL = getEnv("GOOGLE_TOKEN_URL", "https://oauth2.googleapis.com/token")
	cfg.GoogleUserInfoURL = getEnv("GOOGLE_USERINFO_URL", "https://openidconnect.googleapis.com/v1/userinfo")
	cfg.JWTSecret = getEnv("JWT_SECRET", "change-me")
	cfg.JWTIssuer = getEnv("JWT_ISSUER", cfg.ServiceName)
	cfg.OAuthStateSecret = getEnv("OAUTH_STATE_SECRET", cfg.JWTSecret)
	cfg.OAuthStateTTL = getDuration("OAUTH_STATE_TTL", 10*time.Minute)
	cfg.JWTExpiresIn = getDuration("JWT_EXPIRES_IN", 24*time.Hour)
	cfg.SessionTTL = getDuration("SESSION_TTL", 24*time.Hour)
	cfg.RedisHost = getEnv("REDIS_HOST", "localhost")
	cfg.RedisPort = getEnv("REDIS_PORT", "6379")
	cfg.RedisPassword = getEnv("REDIS_PASSWORD", "")
	cfg.RedisDB = getEnvInt("REDIS_DB", 0)
	cfg.UpstreamTimeout = getDuration("UPSTREAM_TIMEOUT", 10*time.Second)

	validate(cfg)
	return cfg
}

func validate(cfg Config) {
	if strings.TrimSpace(cfg.Port) == "" {
		log.Fatal("PORT is required")
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func getDurationSeconds(key string, fallback int) time.Duration {
	raw := getEnv(key, strconv.Itoa(fallback))
	v, err := strconv.Atoi(raw)
	if err != nil {
		return time.Duration(fallback) * time.Second
	}
	return time.Duration(v) * time.Second
}

func getDuration(key string, fallback time.Duration) time.Duration {
	raw := getEnv(key, fallback.String())
	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	raw := getEnv(key, strconv.Itoa(fallback))
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
