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

	RedisHost             string
	RedisPort             string
	KafkaBrokers          string
	KafkaTopicPostCreated string
	KafkaTopicFollowed    string
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

	cfg.RedisHost = getEnv("REDIS_HOST", "localhost")
	cfg.RedisPort = getEnv("REDIS_PORT", "6379")
	cfg.KafkaBrokers = strings.TrimSpace(getEnv("KAFKA_BROKERS", ""))
	cfg.KafkaTopicPostCreated = getEnv("KAFKA_TOPIC_POST_CREATED", "post.created")
	cfg.KafkaTopicFollowed = getEnv("KAFKA_TOPIC_USER_FOLLOWED", "user.followed")

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
