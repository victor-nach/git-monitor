package commit

import (
	"context"
	"fmt"

	"github.com/victor-nach/git-monitor/internal/domain/models"
)

type (
	service struct {
		commitStore commitStore
	}

	commitStore interface {
		GetTopAuthors(ctx context.Context, RepoInfo models.RepoInfo, limit int) ([]models.AuthorStats, error)
		List(ctx context.Context, RepoInfo models.RepoInfo, pagination models.PaginationReq) ([]models.Commit, string, error)
		CreateBatch(ctx context.Context, commits []models.Commit) error
	}
)

func New(commitStore commitStore) *service {
	return &service{
		commitStore: commitStore,
	}
}

func (s *service) GetTopAuthors(ctx context.Context, RepoInfo models.RepoInfo, limit int) ([]models.AuthorStats, error) {
	stats, err := s.commitStore.GetTopAuthors(ctx, RepoInfo, limit)
	if err != nil {
		return []models.AuthorStats{}, fmt.Errorf("error retriving author stats %w", err)
	}

	return stats, nil
}

func (s *service) List(ctx context.Context, RepoInfo models.RepoInfo, pagination models.PaginationReq) ([]models.Commit, string, error) {
	commits, cursor, err := s.commitStore.List(ctx, RepoInfo, pagination)
	if err != nil {
		return []models.Commit{}, "", fmt.Errorf("error retriving author stats %w", err)
	}

	return commits, cursor, nil
}

func (s *service) CreateBatch(ctx context.Context, commits []models.Commit) error {
	return s.commitStore.CreateBatch(ctx, commits)
}
