package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/victor-nach/git-monitor/internal/domain/models"
	"github.com/victor-nach/git-monitor/internal/http/errors"
	"github.com/victor-nach/git-monitor/internal/http/utils"
	"go.uber.org/zap"
)

type contextKey string

const repoInfoKey contextKey = "repoInfo"

// RepoInfoMiddleware validates the repository info and adds it to the context
func (h *Handler) RepoInfoMiddleware(c *gin.Context) {
	log := h.log.With(zap.String("method", "RepoInfoMiddleware"))

	repoInfo := utils.ExtractRepoInfo(c)

	log = log.With(
		zap.String("repo_owner", repoInfo.Owner),
		zap.String("repo_name", repoInfo.Name),
	)

	log.Info("validating repository info")

	if repoInfo.Owner == "" || repoInfo.Name == "" {
		log.Error("invalid repository info: owner or name is missing")
		c.JSON(http.StatusBadRequest, errors.ErrMissingRepoInfo)
		c.Abort()
		return
	}

	ctx := context.WithValue(c.Request.Context(), repoInfoKey, repoInfo)
	c.Request = c.Request.WithContext(ctx)

	log.Info("repository info validated and added to context")
	c.Next()
}

func GetRepoInfo(ctx context.Context) (models.RepoInfo, error) {
	repoInfo, ok := ctx.Value(repoInfoKey).(models.RepoInfo)
	if !ok {
		return models.RepoInfo{}, errors.ErrRepoInfoNotFound
	}
	return repoInfo, nil
}
