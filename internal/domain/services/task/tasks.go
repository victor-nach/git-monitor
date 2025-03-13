package task

import (
	"context"
	"fmt"
	"time"

	"github.com/victor-nach/git-monitor/internal/domain/events"
	"github.com/victor-nach/git-monitor/internal/domain/models"
)

type (
	service struct {
		taskStore taskStore
		repoStore repoStore
		publisher publisher
	}

	taskStore interface {
		Get(ctx context.Context, taskID string) (models.Task, error)
		Create(ctx context.Context, task models.Task) error
		List(ctx context.Context) ([]models.Task, error)
		UpdateStatus(ctx context.Context, taskID string, status string, errMsg *string) error
	}

	repoStore interface {
		Get(ctx context.Context, RepoInfo models.RepoInfo) (models.Repository, error)
		List(ctx context.Context) ([]models.Repository, error)
	}

	publisher interface {
		Publish(ctx context.Context, topic string, message interface{}) error
	}
)

func New(taskStore taskStore, repoStore repoStore, publisher publisher) *service {
	return &service{
		taskStore: taskStore,
		repoStore: repoStore,
		publisher: publisher,
	}
}

func (s *service) StartTasks(ctx context.Context) error {
	activeRepos, err := s.repoStore.List(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving active repos %w", err)
	}

	for _, repo := range activeRepos {
		if !repo.IsActive {
			continue
		}
		if _, err := s.handleTask(ctx, repo, repo.LastFetchedCommitTime); err != nil {
			return fmt.Errorf("error starting task %w", err)
		}
	}
	return nil
}

func (s *service) TriggerTask(ctx context.Context, RepoInfo models.RepoInfo, since *time.Time) (string, error) {
	repo, err := s.repoStore.Get(ctx, RepoInfo)
	if err != nil {
		return "", fmt.Errorf("error retrieving active repos %w", err)
	}
	return s.handleTask(ctx, repo, since)
}

func (s *service) handleTask(ctx context.Context, repo models.Repository, since *time.Time) (string, error) {
	task := models.Task{
		ID:            models.NewUUIDWithPrefix(models.TaskPrefix),
		RepoName:      repo.Name,
		RepositoryID:  repo.ID,
		RepoOwner:     repo.Owner,
		Status:        models.TaskStatusPending,
		CreatedAt:     time.Now(),
	}
	if err := s.taskStore.Create(ctx, task); err != nil {
		return "", fmt.Errorf("failed to create task")
	}

	event := events.FetchCommitEvent{
		TaskID: task.ID,
		RepoInfo: models.RepoInfo{
			Owner: repo.Owner,
			Name:  repo.Name,
		},
		RepoID:        repo.ID,
		Since: repo.CommitTrackingStartTime,
	}
	if since != nil {
		event.Since = *since
	}
	if err := s.publisher.Publish(ctx, events.FetchCommitEventTopic, event); err != nil {
		return "", fmt.Errorf("failed to publish event")
	}

	return task.ID, nil
}

func (s *service) List(ctx context.Context) ([]models.Task, error) {
	return s.taskStore.List(ctx)
}

func (s *service) GetTask(ctx context.Context, taskID string) (models.Task, error) {
	return s.taskStore.Get(ctx, taskID)
}

func (s *service) UpdateStatus(ctx context.Context, taskID string, status string, errMsg *string) error {
	return s.taskStore.UpdateStatus(ctx, taskID, status, errMsg)
}
