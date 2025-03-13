package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

const (
	AppEnvProduction  = "production"
	AppEnvDevelopment = "development"
)

type Config struct {
	Port             string        `envconfig:"PORT" default:"8080"`                                       // Port to run the server on
	AppEnv           string        `envconfig:"APP_ENV" default:"development"`                             // Application environment (production, development)
	GithubToken      string        `envconfig:"GITHUB_TOKEN" required:"true"`                              // GitHub API token (required)
	DBFileName       string        `envconfig:"DB_FILE_NAME" default:"app.db"`                             // Database file name
	RateLimit        int           `envconfig:"RATE_LIMIT_RPS" default:"5"`                                // Rate limit in requests per second
	QueueSize        int           `envconfig:"QUEUE_SIZE" default:"100"`                                  // Size of the job queue
	WorkerSize       int           `envconfig:"WORKER_SIZE" default:"2"`                                   // Number of worker goroutines
	RabbitMQURL      string        `envconfig:"RABBITMQ_URL" default:"amqp://guest:guest@localhost:5672/"` // RabbitMQ connection URL
	GithubBatchSize  int           `envconfig:"GITHUB_BATCH_SIZE" default:"100"`                           // Number of items to process in a batch from GitHub
	ScheduleInterval time.Duration `envconfig:"SCHEDULE_INTERVAL_MINUTES" default:"60m"`                   // Interval for scheduled tasks (e.g - "60m", "1h")
}

func Load(log *zap.Logger) (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Info("No .env file found, using system environment variables")
	}

	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate required fields
	if cfg.GithubToken == "" {
		return nil, fmt.Errorf("GitHub token is required")
	}
	if cfg.RateLimit <= 0 {
		return nil, fmt.Errorf("rate limit must be a positive integer")
	}
	if cfg.QueueSize <= 0 {
		return nil, fmt.Errorf("queue size must be a positive integer")
	}
	if cfg.WorkerSize <= 0 {
		return nil, fmt.Errorf("worker size must be a positive integer")
	}
	if cfg.GithubBatchSize <= 0 {
		return nil, fmt.Errorf("GitHub batch size must be a positive integer")
	}

	log.Info("Configuration loaded",
		zap.String("port", cfg.Port),
		zap.String("app_env", cfg.AppEnv),
		zap.String("db_file_name", cfg.DBFileName),
		zap.Int("rate_limit", cfg.RateLimit),
		zap.Int("queue_size", cfg.QueueSize),
		zap.Int("worker_size", cfg.WorkerSize),
		zap.String("rabbitmq_url", cfg.RabbitMQURL),
		zap.Int("github_batch_size", cfg.GithubBatchSize),
		zap.Duration("schedule_interval", cfg.ScheduleInterval),
	)

	return &cfg, nil
}
