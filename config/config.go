package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const (
	AppEnvProduction  = "production"
	AppEnvDevelopment = "development"
)

type Config struct {
	port             string
	appEnv           string
	githubToken      string
	dbFileName       string
	queueBufferSize  int
	workerSize       int
	rabbitMQURL      string
	githubBatchSize  int
	scheduleInterval time.Duration
}

func Load(log *zap.Logger) (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Info("No .env file found, using system environment variables")
	}

	cfg := &Config{
		port:             getEnv("PORT", "8080"),
		appEnv:           getEnv("APP_ENV", "development"),
		githubToken:      getEnv("GITHUB_TOKEN", ""),
		dbFileName:       getEnv("DB_FILE_NAME", "app.db"),
		queueBufferSize:  getEnvAsInt("QUEUE_BUFFER_SIZE", 100),
		workerSize:       getEnvAsInt("WORKER_SIZE", 2),
		rabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		githubBatchSize:  getEnvAsInt("GITHUB_BATCH_SIZE", 100),
		scheduleInterval: getEnvAsDuration("SCHEDULE_INTERVAL_MINUTES", 60*time.Minute),
	}

	// Validate required fields
	if cfg.githubToken == "" {
		return nil, fmt.Errorf("GitHub token is required")
	}
	if cfg.workerSize <= 0 {
		return nil, fmt.Errorf("worker size must be a positive integer")
	}
	if cfg.githubBatchSize <= 0 {
		return nil, fmt.Errorf("GitHub batch size must be a positive integer")
	}
	if cfg.queueBufferSize <= 0 {
		return nil, fmt.Errorf("queue buffer size must be a positive integer")
	}

	log.Info("Configuration loaded",
		zap.String("port", cfg.port),
		zap.String("app_env", cfg.appEnv),
		zap.String("db_file_name", cfg.dbFileName),
		zap.Int("worker_size", cfg.workerSize),
		zap.String("rabbitmq_url", cfg.rabbitMQURL),
		zap.Int("github_batch_size", cfg.githubBatchSize),
		zap.Duration("schedule_interval", cfg.scheduleInterval),
		zap.Int("queue_buffer_size", cfg.queueBufferSize),
	)

	return cfg, nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	strValue := getEnv(key, "")
	if strValue == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(strValue)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	strValue := getEnv(key, "")
	if strValue == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(strValue)
	if err != nil {
		return defaultValue
	}
	return value
}

// Getter methods (unchanged)
func (c *Config) GetPort() string {
	return c.port
}

func (c *Config) GetAppEnv() string {
	return c.appEnv
}

func (c *Config) GetGithubToken() string {
	return c.githubToken
}

func (c *Config) GetDBFileName() string {
	return c.dbFileName
}

func (c *Config) GetQueueBufferSize() int {
	return c.queueBufferSize
}

func (c *Config) GetWorkerSize() int {
	return c.workerSize
}

func (c *Config) GetRabbitMQURL() string {
	return c.rabbitMQURL
}

func (c *Config) GetGithubBatchSize() int {
	return c.githubBatchSize
}

func (c *Config) GetScheduleInterval() time.Duration {
	return c.scheduleInterval
}
