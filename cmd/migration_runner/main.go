package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/victor-nach/git-monitor/internal/config"
	"github.com/victor-nach/git-monitor/internal/db"
	"github.com/victor-nach/git-monitor/pkg/logger"
	"github.com/victor-nach/git-monitor/pkg/utils"
	_ "modernc.org/sqlite"
)

func main() {
	appEnv, ok := os.LookupEnv("APP_ENV")
	if !ok {
		appEnv = config.AppEnvDevelopment
	}

	logr, err := logger.New(appEnv)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logr.Sync()

	cfg, err := config.Load(logr)
	if err != nil {
		logr.Fatal("failed to load configuration", zap.Error(err))
	}

	projectRoot, err := utils.GetBasePath()
	if err != nil {
		log.Fatalf("failed to get project root: %v", err)
	}

	// Set up the database file path
	dbf := filepath.Join(projectRoot, "data", cfg.DBFileName)
	mf := filepath.Join(projectRoot, "migrations")
	mf = filepath.ToSlash(mf)

	ctx := context.Background()
	_, sqlDB, err := db.New(ctx, logr, dbf, mf)
	if err != nil {
		logr.Fatal("failed to initialize database", zap.Error(err))
	}
	defer sqlDB.Close()

	logr.Info("migrations applied successfully")
}
