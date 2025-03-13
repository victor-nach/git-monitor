package utils

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/victor-nach/git-monitor/internal/domain/models"
	"github.com/victor-nach/git-monitor/internal/http/errors"

	"go.uber.org/zap"
)

func ExtractRepoInfo(c *gin.Context) models.RepoInfo {
	owner := c.Param("owner")
	repo := c.Param("repo")

	if owner == "" || owner == ":owner" {
		owner = ""
	}
	if repo == "" || repo == ":repo" {
		repo = ""
	}

	return models.RepoInfo{
		Owner: owner,
		Name:  repo,
	}
}

func ExtractLimit(c *gin.Context) int {
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			return l
		}
	}
	return limit
}

func ExtractPaginationReq(c *gin.Context) models.PaginationReq {
	return models.PaginationReq{
		Limit:  ExtractLimit(c),
		Cursor: c.Query("cursor"),
	}
}

func ExtractTime(c *gin.Context, key string) (*time.Time, error) {
	timeStr := c.Query(key)
	if timeStr == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, timeStr); 
	if err != nil {
		return nil, errors.InvalidTimeFormat(key)
	}

	return &t, nil
}

func WithRepoInfo(log *zap.Logger, repoInfo models.RepoInfo) *zap.Logger {
	return log.With(
		zap.String("repo_owner", repoInfo.Owner),
		zap.String("repo_name", repoInfo.Name),
	)
}
