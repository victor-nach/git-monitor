package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/victor-nach/git-monitor/pkg/migrator"
	_ "modernc.org/sqlite"
)

type DBConfig struct {
	DBFilePath      string
	MigrationsPath  string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	BusyTimeout     int
	JournalMode     string
}

func New(ctx context.Context, log *zap.Logger, dbPath, migrationsPath string, opts ...func(*DBConfig)) (*gorm.DB, *sql.DB, error) {
	cfg := DBConfig{
		DBFilePath:      dbPath,
		MigrationsPath:  migrationsPath,
		MaxOpenConns:    2,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
		BusyTimeout:     5000,
		JournalMode:     "WAL",
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	if _, err := os.Stat(filepath.Dir(cfg.DBFilePath)); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("database directory does not exist: %w", err)
	}

	dsn := fmt.Sprintf(
		"file:%s?cache=shared&_busy_timeout=%d&_journal_mode=%s",
		cfg.DBFilePath,
		cfg.BusyTimeout,
		cfg.JournalMode,
	)

	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open SQL database: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to ping database: %w", err)
	}

	gormDB, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	if err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to initialize GORM: %w", err)
	}

	migrationsURL := cfg.MigrationsPath
	if err := migrator.Migrate(ctx, sqlDB, migrationsURL, log); err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Info("database initialized successfully")

	return gormDB, sqlDB, nil
}
