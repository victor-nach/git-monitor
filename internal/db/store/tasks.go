package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/victor-nach/git-monitor/internal/domain/models"
	"gorm.io/gorm"
)

type taskStore struct {
	db *gorm.DB
}

func (s *store) NewTaskStore() *taskStore {
	return &taskStore{
		db: s.db,
	}
}

func (s *taskStore) Get(ctx context.Context, taskID string) (models.Task, error) {
	var task models.Task
	err := s.db.WithContext(ctx).
		Where("id = ?", taskID).
		First(&task).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Task{}, fmt.Errorf("task not found: %w", err)
		}
		return models.Task{}, err
	}
	return task, nil
}

func (s *taskStore) Create(ctx context.Context, task models.Task) error {
	return s.db.WithContext(ctx).Create(&task).Error
}

func (s *taskStore) List(ctx context.Context) ([]models.Task, error) {
	var tasks []models.Task
	err := s.db.WithContext(ctx).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *taskStore) UpdateStatus(ctx context.Context, taskID string, status string, errMsg *string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if errMsg != nil {
		updates["error_message"] = *errMsg
	}
	if status == models.TaskStatusCompleted {
		updates["completed_at"] = time.Now()
	}
	result := s.db.Model(&models.Task{}).
		Where("id = ?", taskID).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update task status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
