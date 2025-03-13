package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/victor-nach/git-monitor/internal/domain/models"
	"go.uber.org/zap"
)

func GetBasePath() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get current file path")
	}

	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return "", fmt.Errorf("project root not found")
		}
		
		dir = parentDir
	}
}

func WithRepoInfo(log *zap.Logger, repoInfo models.RepoInfo) *zap.Logger {
	return log.With(
		zap.String("repo_owner", repoInfo.Owner),
		zap.String("repo_name", repoInfo.Name),
	)
}
