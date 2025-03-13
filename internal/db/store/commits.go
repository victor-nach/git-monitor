package store

import (
	"context"
	"fmt"
	"time"

	"github.com/victor-nach/git-monitor/internal/domain/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type commitStore struct {
	db *gorm.DB
}

func (s *store) NewCommitStore() *commitStore {
	return &commitStore{
		db: s.db,
	}
}

func (s *commitStore) GetTopAuthors(ctx context.Context, RepoInfo models.RepoInfo, limit int) ([]models.AuthorStats, error) {
	var authorStats []models.AuthorStats

	err := s.db.WithContext(ctx).
		Model(&models.Commit{}).
		Select("author, COUNT(*) as commits").
		Where("repo_name = ? AND repo_owner = ?", RepoInfo.Name, RepoInfo.Owner).
		Group("author").
		Order("commits DESC").
		Limit(limit).
		Find(&authorStats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch top authors: %w", err)
	}

	return authorStats, nil
}

func (s *commitStore) List(ctx context.Context, RepoInfo models.RepoInfo, pagination models.PaginationReq) ([]models.Commit, string, error) {
	var commits []models.Commit

	query := s.db.WithContext(ctx).
		Where("repo_name = ? AND repo_owner = ?", RepoInfo.Name, RepoInfo.Owner).
		Order("date DESC")

	if pagination.Cursor != "" {
		query = query.Where("date < ?", pagination.Cursor)
	}

	err := query.Limit(pagination.Limit).Find(&commits).Error
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch commits: %w", err)
	}

	var nextCursor string
	if len(commits) > 0 {
		nextCursor = commits[len(commits)-1].Date.Format(time.RFC3339)
	}

	return commits, nextCursor, nil
}

func (s *commitStore) CreateBatch(ctx context.Context, commits []models.Commit) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&commits).Error; err != nil {
			return fmt.Errorf("failed to insert commits: %w", err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create commit batch: %w", err)
	}

	return nil
}
