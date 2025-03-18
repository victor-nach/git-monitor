package main

import (
	"context"
	"database/sql"
	ilog "log"
	"os"
	"path/filepath"
	"time"

	"github.com/victor-nach/git-monitor/config"
	"github.com/victor-nach/git-monitor/internal/db"
	"github.com/victor-nach/git-monitor/internal/db/store"
	"github.com/victor-nach/git-monitor/internal/domain/services/commit"
	"github.com/victor-nach/git-monitor/pkg/github"

	"github.com/victor-nach/git-monitor/internal/domain/services/repository"
	"github.com/victor-nach/git-monitor/internal/domain/services/task"
	"github.com/victor-nach/git-monitor/internal/http/handlers"
	"github.com/victor-nach/git-monitor/internal/http/server"
	"github.com/victor-nach/git-monitor/internal/scheduler"
	"github.com/victor-nach/git-monitor/internal/worker/fetcher"
	"github.com/victor-nach/git-monitor/internal/worker/saver"
	"github.com/victor-nach/git-monitor/pkg/eventbus"
	"github.com/victor-nach/git-monitor/pkg/githubclient"
	"github.com/victor-nach/git-monitor/pkg/logger"
	"github.com/victor-nach/git-monitor/pkg/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	log := InitLogger()
	defer log.Sync()

	log.Info("starting application...")

	cfg := loadConfiguration(log)

	gormDB, sqlDB := initDatabase(log, cfg)
	defer sqlDB.Close()

	run(log, cfg, gormDB)
}

func run(log *zap.Logger, cfg *config.Config, gormDB *gorm.DB) {
	db := store.New(gormDB)
	repoStore := db.NewRepoStore()
	commitStore := db.NewCommitStore()
	taskStore := db.NewTaskStore()

	eventBus := eventbus.NewInMemoryEventBus(log, cfg.GetQueueBufferSize())
	defer eventBus.Close()
	gitClient := githubclient.New(cfg.GetGithubToken(), log)

	githubSvc := github.New(log, gitClient, cfg.GetGithubBatchSize())
	tasksSvc := task.New(taskStore, repoStore, eventBus)
	repoSvc := repository.New(repoStore, tasksSvc, githubSvc)
	commitSvc := commit.New(commitStore)

	ctx := context.Background()
	schedulerSvc := scheduler.New(log, tasksSvc, cfg.GetScheduleInterval())
	go schedulerSvc.Start(ctx)

	fetcherWorker := fetcher.New(log, githubSvc, tasksSvc, eventBus, cfg.GetWorkerSize())
	if err := fetcherWorker.Subscribe(ctx); err != nil {
		log.Fatal("failed to subscribe fetcher worker", zap.Error(err))
	}

	saverWorker := saver.New(log, commitSvc, repoSvc, eventBus, cfg.GetWorkerSize())
	if err := saverWorker.Subscribe(ctx); err != nil {
		log.Fatal("failed to subscribe saver worker", zap.Error(err))
	}

	handlers := handlers.New(log, repoSvc, commitSvc, tasksSvc)
	server.Run(log, handlers, cfg.GetPort())
}

func InitLogger() *zap.Logger {
	appEnv, ok := os.LookupEnv("APP_ENV")
	if !ok {
		appEnv = config.AppEnvDevelopment
	}

	log, err := logger.New(appEnv)
	if err != nil {
		ilog.Fatalf("failed to initialize logger: %v", err)
	}
	return log
}

func loadConfiguration(log *zap.Logger) *config.Config {
	cfg, err := config.Load(log)
	if err != nil {
		log.Fatal("failed to load configuration", zap.Error(err))
	}

	return cfg
}

func initDatabase(log *zap.Logger, cfg *config.Config) (*gorm.DB, *sql.DB) {
	projectRoot, err := utils.GetBasePath()
	if err != nil {
		log.Fatal("failed to get project root", zap.Error(err))
	}
	dbf := filepath.Join(projectRoot, "data", cfg.GetDBFileName())
	mf := filepath.Join(projectRoot, "migrations")
	mf = filepath.ToSlash(mf)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	gormDB, sqlDB, err := db.New(ctx, log, dbf, mf)
	if err != nil {
		log.Fatal("failed to initialize database", zap.Error(err))
	}
	return gormDB, sqlDB
}
