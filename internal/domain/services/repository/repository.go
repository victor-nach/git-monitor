package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/victor-nach/git-monitor/internal/domain/errors"
	"github.com/victor-nach/git-monitor/internal/domain/models"
)

type (
	service struct {
		repoStore     repoStore
		taskSvc       taskSvc
		githubService githubService
	}

	repoStore interface {
		Create(ctx context.Context, repo models.Repository) error
		CheckExists(ctx context.Context, RepoInfo models.RepoInfo) (bool, error)
		List(ctx context.Context) ([]models.Repository, error)
		Reset(ctx context.Context, RepoInfo models.RepoInfo, startTime *time.Time) error
		UpdateStatus(ctx context.Context, RepoInfo models.RepoInfo, isActive *bool) error
		UpdateTrackingInfo(ctx context.Context, repoInfo models.RepoInfo, lastFetchedCommitTime time.Time) error
	}

	taskSvc interface {
		TriggerTask(ctx context.Context, RepoInfo models.RepoInfo, since *time.Time) (string, error)
	}

	githubService interface {
		GetRepository(ctx context.Context, RepoInfo models.RepoInfo) (models.Repository, error)
	}
)

func New(repoStore repoStore, taskSvc taskSvc, githubService githubService) *service {
	return &service{
		repoStore:     repoStore,
		taskSvc:       taskSvc,
		githubService: githubService,
	}
}

// AddTrackedRepository adds a new repository to the list of tracked repositories
// and starts the fetch events task
func (s *service) Create(ctx context.Context, RepoInfo models.RepoInfo, since *time.Time) (models.Repository, string, error) {
	exists, err := s.repoStore.CheckExists(ctx, RepoInfo)
	if err != nil {
		return models.Repository{}, "", fmt.Errorf("error validating repo name %w", err)
	}
	if exists {
		return models.Repository{}, "", errors.ErrDuplicateRepository
	}

	newRepo, err := s.githubService.GetRepository(ctx, RepoInfo)
	if err != nil {
		return models.Repository{}, "", fmt.Errorf("error getting git repo from github %w", err)
	}

	commitStartTrackingTime := newRepo.CreatedAt
	if since != nil {
		commitStartTrackingTime = *since
	}
	newRepo.CommitTrackingStartTime = commitStartTrackingTime
	if err := s.repoStore.Create(ctx, newRepo); err != nil {
		return models.Repository{}, "", fmt.Errorf("error creating repository %w", err)
	}

	taskID, err := s.taskSvc.TriggerTask(ctx, RepoInfo, &commitStartTrackingTime)
	if  err != nil {
		return models.Repository{}, "", fmt.Errorf("error sending task event %w", err)
	}

	return newRepo, taskID, nil
}

func (s *service) List(ctx context.Context) ([]models.Repository, error) {
	repos, err := s.repoStore.List(ctx)
	if err != nil {
		return []models.Repository{}, fmt.Errorf("error listing repositories %w", err)
	}

	return repos, nil
}

func (s *service) Reset(ctx context.Context, RepoInfo models.RepoInfo, since *time.Time) (string, error) {
	if err := s.CheckExists(ctx, RepoInfo); err != nil {
		return "", fmt.Errorf("error validating repo name %w", err)
	}

	if err := s.repoStore.Reset(ctx, RepoInfo, since); err != nil {
		return "", fmt.Errorf("error resetting repository %w", err)
	}

	taskID, err := s.taskSvc.TriggerTask(ctx, RepoInfo, since)
	if  err != nil {
		return "", fmt.Errorf("error sending task event %w", err)
	}

	return taskID, nil
}

func (s *service) UpdateStatus(ctx context.Context, RepoInfo models.RepoInfo, isActive *bool) error {
	if err := s.CheckExists(ctx, RepoInfo); err != nil {
		return fmt.Errorf("error validating repo name %w", err)
	}

	if err := s.repoStore.UpdateStatus(ctx, RepoInfo, isActive); err != nil {
		return fmt.Errorf("error updating repository status %w", err)
	}

	return nil
}

func (s *service) CheckExists(ctx context.Context, RepoInfo models.RepoInfo) error {
	exists, err := s.repoStore.CheckExists(ctx, RepoInfo)
	if err != nil {
		return fmt.Errorf("error validating repo name %w", err)
	}
	if !exists {
		return errors.ErrTrackedRepositoryNotFound
	}

	return nil
}

func (s *service) UpdateTrackingInfo(ctx context.Context, repoInfo models.RepoInfo, lastFetchedCommitTime time.Time) error {
	if err := s.repoStore.UpdateTrackingInfo(ctx, repoInfo, lastFetchedCommitTime); err != nil {
		return fmt.Errorf("failed to update repository commit tracking: %w", err)
	}
	return nil
}
