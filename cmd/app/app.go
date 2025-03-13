package main

// import (
// 	"context"
// 	"database/sql"
// 	ilog "log"
// 	"os"
// 	"path/filepath"
// 	"time"

// 	"github.com/victor-nach/git-monitor/internal/config"
// 	"github.com/victor-nach/git-monitor/internal/db"
// 	"github.com/victor-nach/git-monitor/internal/db/store"
// 	"github.com/victor-nach/git-monitor/internal/domain/services/commit"
// 	"github.com/victor-nach/git-monitor/internal/domain/services/github"
// 	"github.com/victor-nach/git-monitor/internal/domain/services/repository"
// 	"github.com/victor-nach/git-monitor/internal/domain/services/task"
// 	"github.com/victor-nach/git-monitor/internal/http/handlers"
// 	"github.com/victor-nach/git-monitor/internal/http/server"
// 	"github.com/victor-nach/git-monitor/internal/scheduler"
// 	"github.com/victor-nach/git-monitor/internal/worker/fetcher"
// 	"github.com/victor-nach/git-monitor/internal/worker/saver"
// 	"github.com/victor-nach/git-monitor/pkg/eventbus"
// 	"github.com/victor-nach/git-monitor/pkg/githubclient"
// 	"github.com/victor-nach/git-monitor/pkg/logger"
// 	"github.com/victor-nach/git-monitor/pkg/utils"
// 	"go.uber.org/zap"
// 	"gorm.io/gorm"
// )

// func run(log *zap.Logger, cfg *config.Config, gormDB *gorm.DB, eventBus *eventbus.RabbitMQEventBus) {
// 	db := store.New(gormDB)
// 	repoStore := db.NewRepoStore()
// 	commitStore := db.NewCommitStore()
// 	taskStore := db.NewTaskStore()

// 	gitClient := githubclient.New(cfg.GithubToken)

// 	githubSvc := github.New(log, gitClient, cfg.GithubBatchSize)
// 	tasksSvc := task.New(taskStore, repoStore, eventBus)
// 	repoSvc := repository.New(repoStore, tasksSvc, githubSvc)
// 	commitSvc := commit.New(commitStore)

// 	ctx := context.Background()
// 	schedulerSvc := scheduler.New(log, tasksSvc, cfg.ScheduleInterval)
// 	schedulerSvc.Start(ctx)

// 	fetcherWorker := fetcher.New(log, githubSvc, tasksSvc, eventBus, cfg.WorkerSize)
// 	if err := fetcherWorker.Subscribe(ctx); err != nil {
// 		log.Fatal("failed to subscribe fetcher worker", zap.Error(err))
// 	}

// 	saverWorker := saver.New(log, commitSvc, eventBus, cfg.WorkerSize)
// 	if err := saverWorker.Subscribe(ctx); err != nil {
// 		log.Fatal("failed to subscribe saver worker", zap.Error(err))
// 	}

// 	handlers := handlers.New(log, repoSvc, commitSvc, tasksSvc)
// 	server.Run(log, handlers, cfg)
// }


// func InitLogger() *zap.Logger {
// 	appEnv, ok := os.LookupEnv("APP_ENV")
// 	if !ok {
// 		appEnv = config.AppEnvDevelopment
// 	}

// 	log, err := logger.New(appEnv)
// 	if err != nil {
// 		ilog.Fatalf("failed to initialize logger: %v", err)
// 	}
// 	return log
// }

// func loadConfiguration(log *zap.Logger) *config.Config {
// 	cfg, err := config.Load(log)
// 	if err != nil {
// 		log.Fatal("failed to load configuration", zap.Error(err))
// 	}

// 	return cfg
// }

// func initDatabase(log *zap.Logger, cfg *config.Config) (*gorm.DB, *sql.DB) {
// 	projectRoot, err := utils.GetBasePath()
// 	if err != nil {
// 		log.Fatal("failed to get project root", zap.Error(err))
// 	}
// 	dbf := filepath.Join(projectRoot, "data", cfg.DBFileName)
// 	mf := filepath.Join(projectRoot, "migrations")
// 	mf = filepath.ToSlash(mf)

// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	gormDB, sqlDB, err := db.New(ctx, log, dbf, mf)
// 	if err != nil {
// 		log.Fatal("failed to initialize database", zap.Error(err))
// 	}
// 	return gormDB, sqlDB
// }

// func initMessageQueue(log *zap.Logger, cfg *config.Config) *eventbus.RabbitMQEventBus {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	eventBus, err := eventbus.NewRabbitMQEventBus(ctx, cfg.RabbitMQURL, log)
// 	if err != nil {
// 		log.Fatal("failed to initialize message queue", zap.Error(err))
// 	}
// 	return eventBus
// }
