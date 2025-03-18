package db

import (
	"context"
	"database/sql"
	"path/filepath"

	"log"

	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/victor-nach/git-monitor/config"
	"github.com/victor-nach/git-monitor/pkg/logger"
	"github.com/victor-nach/git-monitor/pkg/migrator"
	"github.com/victor-nach/git-monitor/pkg/utils"

	_ "modernc.org/sqlite"
)

var (
	db      *gorm.DB
	sqlDB   *sql.DB
	testCtx = context.Background()
)

func TestMain(m *testing.M) {
	logr, err := logger.New(config.AppEnvDevelopment)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logr.Sync()

	sqlDB, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %v", err)
		return
	}

	db, err = gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to initialize GORM: %v", err)
		return
	}

	projectRoot, err := utils.GetBasePath()
	if err != nil {
		log.Fatalf("failed to get project root: %v", err)
	}

	mf := filepath.Join(projectRoot, "migrations")
	mf = filepath.ToSlash(mf)

	if err := migrator.Migrate(testCtx, sqlDB, mf, logr); err != nil {
		sqlDB.Close()
		log.Fatalf("failed to apply migrations: %v", err)
		return
	}

	code := m.Run()

	sqlDB.Close()
	os.Exit(code)
}
