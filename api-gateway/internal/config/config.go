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

	AuthServiceURL         string
	UsersServiceURL        string
	PostsServiceURL        string
	FeedServiceURL         string
	NotificationServiceURL string

	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	JWTSecret string
	JWTIssuer string

	UpstreamTimeout        time.Duration
	UpstreamRetryAttempts  int
	UpstreamRetryBackoff   time.Duration
	CircuitBreakerFailures int
	CircuitBreakerOpenFor  time.Duration

	RateLimitPerMinute int
	RateLimitWindow    time.Duration

	TLSEnabled  bool
	TLSCertFile string
	TLSKeyFile  string

	TrustProxyHeaders bool
	RequireHTTPS      bool

	DemoSimulationEnabled bool
	DemoSimulationPath    string
	DemoLatency           time.Duration
	DemoFailureRate       float64
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

	cfg.AuthServiceURL = getEnv("AUTH_SERVICE_URL", "http://localhost:8081")
	cfg.UsersServiceURL = getEnv("USERS_SERVICE_URL", "http://localhost:8082")
	cfg.PostsServiceURL = getEnv("POSTS_SERVICE_URL", "http://localhost:8083")
	cfg.FeedServiceURL = getEnv("FEED_SERVICE_URL", "http://localhost:8084")
	cfg.NotificationServiceURL = getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8085")

	cfg.RedisHost = getEnv("REDIS_HOST", "localhost")
	cfg.RedisPort = getEnv("REDIS_PORT", "6379")
	cfg.RedisPassword = getEnv("REDIS_PASSWORD", "")
	cfg.RedisDB = getEnvInt("REDIS_DB", 0)

	cfg.JWTSecret = getEnv("JWT_SECRET", "change-me")
	cfg.JWTIssuer = getEnv("JWT_ISSUER", "auth-service")

	cfg.UpstreamTimeout = getDuration("UPSTREAM_TIMEOUT", 10*time.Second)
	cfg.UpstreamRetryAttempts = getEnvInt("UPSTREAM_RETRY_ATTEMPTS", 3)
	cfg.UpstreamRetryBackoff = getDuration("UPSTREAM_RETRY_BACKOFF", 100*time.Millisecond)
	cfg.CircuitBreakerFailures = getEnvInt("CIRCUIT_BREAKER_FAILURES", 5)
	cfg.CircuitBreakerOpenFor = getDuration("CIRCUIT_BREAKER_OPEN_FOR", 30*time.Second)

	cfg.RateLimitPerMinute = getEnvInt("RATE_LIMIT_PER_MINUTE", 100)
	cfg.RateLimitWindow = getDuration("RATE_LIMIT_WINDOW", time.Minute)

	cfg.TLSEnabled = getEnvBool("TLS_ENABLED", false)
	cfg.TLSCertFile = getEnv("TLS_CERT_FILE", "")
	cfg.TLSKeyFile = getEnv("TLS_KEY_FILE", "")

	cfg.TrustProxyHeaders = getEnvBool("TRUST_PROXY_HEADERS", false)
	cfg.RequireHTTPS = getEnvBool("REQUIRE_HTTPS", false)
	cfg.DemoSimulationEnabled = getEnvBool("DEMO_SIMULATION_ENABLED", false)
	cfg.DemoSimulationPath = getEnv("DEMO_SIMULATION_PATH", "/api/v1/feed")
	cfg.DemoLatency = getDuration("DEMO_LATENCY", 0)
	cfg.DemoFailureRate = getEnvFloat("DEMO_FAILURE_RATE", 0)
	validate(cfg)
	return cfg
}

func validate(cfg Config) {
	if strings.TrimSpace(cfg.Port) == "" {
		log.Fatal("PORT is required")
	}
	if cfg.RateLimitPerMinute <= 0 {
		log.Fatal("RATE_LIMIT_PER_MINUTE must be greater than 0")
	}

	if cfg.RateLimitWindow <= 0 {
		log.Fatal("RATE_LIMIT_WINDOW must be greater than 0")
	}
	if cfg.UpstreamRetryAttempts <= 0 {
		log.Fatal("UPSTREAM_RETRY_ATTEMPTS must be greater than 0")
	}
	if cfg.UpstreamRetryBackoff <= 0 {
		log.Fatal("UPSTREAM_RETRY_BACKOFF must be greater than 0")
	}
	if cfg.CircuitBreakerFailures <= 0 {
		log.Fatal("CIRCUIT_BREAKER_FAILURES must be greater than 0")
	}
	if cfg.CircuitBreakerOpenFor <= 0 {
		log.Fatal("CIRCUIT_BREAKER_OPEN_FOR must be greater than 0")
	}

	if cfg.TLSEnabled {
		if strings.TrimSpace(cfg.TLSCertFile) == "" {
			log.Fatal("TLS_CERT_FILE is required when TLS_ENABLED=true")
		}
		if strings.TrimSpace(cfg.TLSKeyFile) == "" {
			log.Fatal("TLS_KEY_FILE is required when TLS_ENABLED=true")
		}
	}

	if cfg.RequireHTTPS && !cfg.TLSEnabled && !cfg.TrustProxyHeaders {
		log.Fatal("REQUIRE_HTTPS=true requires either TLS_ENABLED=true or TRUST_PROXY_HEADERS=true")
	}
	if cfg.DemoFailureRate < 0 || cfg.DemoFailureRate > 1 {
		log.Fatal("DEMO_FAILURE_RATE must be between 0 and 1")
	}
	if cfg.DemoSimulationEnabled && strings.TrimSpace(cfg.DemoSimulationPath) == "" {
		log.Fatal("DEMO_SIMULATION_PATH is required when DEMO_SIMULATION_ENABLED=true")
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

func getEnvFloat(key string, fallback float64) float64 {
	raw := getEnv(key, strconv.FormatFloat(fallback, 'f', -1, 64))
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fallback
	}
	return value
}

func getEnvBool(key string, fallback bool) bool {
	raw := strings.ToLower(strings.TrimSpace(getEnv(key, strconv.FormatBool(fallback))))
	switch raw {
	case "true", "1", "yes", "y", "on":
		return true
	case "false", "0", "no", "n", "off":
		return false
	default:
		return fallback
	}
}
