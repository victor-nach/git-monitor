package handlers

import (
	"context"
	// "strconv"
	"time"

	// "github.com/gin-gonic/gin"
	"github.com/victor-nach/git-monitor/internal/domain/models"
	"go.uber.org/zap"
)

type (
	Handler struct {
		log       *zap.Logger
		repoSvc   repoSvc
		commitSvc commitSvc
		taskSvc   taskSvc
	}

	repoSvc interface {
		Create(ctx context.Context, RepoInfo models.RepoInfo, since *time.Time) (models.Repository, string, error)
		List(ctx context.Context) ([]models.Repository, error)
		Reset(ctx context.Context, RepoInfo models.RepoInfo, startTime *time.Time) (string, error)
		UpdateStatus(ctx context.Context, RepoInfo models.RepoInfo, isActive *bool) error
	}

	taskSvc interface {
		GetTask(ctx context.Context, id string) (models.Task, error)
		List(ctx context.Context) ([]models.Task, error)
		TriggerTask(ctx context.Context, RepoInfo models.RepoInfo, since *time.Time)  (string, error)
	}

	commitSvc interface {
		GetTopAuthors(ctx context.Context, RepoInfo models.RepoInfo, limit int) ([]models.AuthorStats, error)
		List(ctx context.Context, RepoInfo models.RepoInfo, pagination models.PaginationReq) ([]models.Commit, string, error)
	}
)

func New(log *zap.Logger, repoSvc repoSvc, commitSvc commitSvc, taskSvc taskSvc) *Handler {
	return &Handler{
		log:       log,
		repoSvc:   repoSvc,
		commitSvc: commitSvc,
		taskSvc:   taskSvc,
	}
}
