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

    DBHost             string
    DBPort             string
    DBName             string
    DBUser             string
    DBPassword         string
    DBSSLMode          string
    KafkaBrokers       string
    KafkaTopicFollowed string

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


    cfg.DBHost = getEnv("DB_HOST", "localhost")
    cfg.DBPort = getEnv("DB_PORT", "5432")
    cfg.DBName = getEnv("DB_NAME", "users_db")
    cfg.DBUser = getEnv("DB_USER", "postgres")
    cfg.DBPassword = getEnv("DB_PASSWORD", "postgres")
    cfg.DBSSLMode = getEnv("DB_SSLMODE", "disable")
    cfg.KafkaBrokers = getEnv("KAFKA_BROKERS", "localhost:9092")
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
