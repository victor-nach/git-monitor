package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	dErrors "github.com/victor-nach/git-monitor/internal/domain/errors"
	"github.com/victor-nach/git-monitor/internal/domain/models"
	"gorm.io/gorm"
)

type repoStore struct {
	db *gorm.DB
}

func (s *store) NewRepoStore() *repoStore {
	return &repoStore{
		db: s.db,
	}
}

func (s *repoStore) Get(ctx context.Context, RepoInfo models.RepoInfo) (models.Repository, error) {
	var repo models.Repository

	err := s.db.WithContext(ctx).
		Where("name = ? AND owner = ?", RepoInfo.Name, RepoInfo.Owner).
		First(&repo).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Repository{}, dErrors.ErrRepositoryNotFound
		}
		return models.Repository{}, err
	}

	return repo, nil
}

func (s *repoStore) List(ctx context.Context) ([]models.Repository, error) {
	var repos []models.Repository

	err := s.db.WithContext(ctx).Find(&repos).Error
	if err != nil {
		return nil, err
	}

	return repos, nil
}

func (s *repoStore) Create(ctx context.Context, repo models.Repository) error {
	if err := s.db.WithContext(ctx).Create(&repo).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return dErrors.ErrDuplicateRepository
		}
		return err
	}

	return nil
}

func (s *repoStore) CheckExists(ctx context.Context, RepoInfo models.RepoInfo) (bool, error) {
	var count int64

	err := s.db.WithContext(ctx).
		Model(&models.Repository{}).
		Where("name = ? AND owner = ?", RepoInfo.Name, RepoInfo.Owner).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *repoStore) Reset(ctx context.Context, RepoInfo models.RepoInfo, startTime *time.Time) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Where("repo_name = ? AND repo_owner = ?", RepoInfo.Name, RepoInfo.Owner).
			Delete(&models.Commit{}).Error; err != nil {
			return err
		}

		updates := map[string]interface{}{
			"commit_tracking_start_time": startTime,
			"last_fetched_at":            nil,
			"is_synced_to_start_time":    false,
			"updated_at":                 time.Now(),
		}

		if err := tx.
			Model(&models.Repository{}).
			Where("name = ? AND owner = ?", RepoInfo.Name, RepoInfo.Owner).
			Updates(updates).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to reset repository: %w", err)
	}

	return nil
}

func (s *repoStore) UpdateStatus(ctx context.Context, RepoInfo models.RepoInfo, isActive *bool) error {
	if isActive == nil {
		return dErrors.ErrInvalidInput
	}

	updates := map[string]interface{}{
		"is_active":  *isActive,
		"updated_at": time.Now(),
	}

	err := s.db.WithContext(ctx).
		Model(&models.Repository{}).
		Where("name = ? AND owner = ?", RepoInfo.Name, RepoInfo.Owner).
		Updates(updates).Error

	if err != nil {
		return fmt.Errorf("failed to update repository status: %w", err)
	}

	return nil
}

func (s *repoStore) UpdateTrackingInfo(ctx context.Context, repoInfo models.RepoInfo, lastFetchedCommitTime time.Time) error {
	var commitCount int64
	if err := s.db.Model(&models.Commit{}).
		Where("repo_name = ? AND repo_owner = ?", repoInfo.Name, repoInfo.Owner).
		Count(&commitCount).Error; err != nil {
		return fmt.Errorf("failed to count commits: %w", err)
	}

	if commitCount != 0 {
		return nil
	}

	updates := map[string]interface{}{
		"last_fetched_commit_time": lastFetchedCommitTime,
		"last_fetched_at":            time.Now(),
		"updated_at":                 time.Now(),
	}

	result := s.db.Model(&models.Repository{}).
		Where("name = ? AND owner = ?", repoInfo.Name, repoInfo.Owner).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update repository: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return dErrors.ErrRepositoryNotFound
	}

	return nil
}
