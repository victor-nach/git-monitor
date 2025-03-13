package saver

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/victor-nach/git-monitor/internal/domain/events"
	"github.com/victor-nach/git-monitor/internal/domain/models"
	"go.uber.org/zap"
)

type (
	worker struct {
		log         *zap.Logger
		commitSvc   commitService
		repoSvc     repoService
		eventBus    eventBus
		workerCount int
	}

	commitService interface {
		CreateBatch(ctx context.Context, commits []models.Commit) error
	}

	repoService interface {
		UpdateTrackingInfo(ctx context.Context, repoInfo models.RepoInfo, lastFetchedCommitTime time.Time) error
	}

	eventBus interface {
		Publish(ctx context.Context, topic string, message interface{}) error
		Subscribe(ctx context.Context, topic string, handler func(message []byte) error) error
	}
)

func New(log *zap.Logger, commitSvc commitService, repoSvc repoService, eventBus eventBus, workerCount int) *worker {
	log = log.With(zap.String("worker", "saver"))

	return &worker{
		log:         log,
		commitSvc:   commitSvc,
		eventBus:    eventBus,
		workerCount: workerCount,
		repoSvc:     repoSvc,
	}
}

func (w *worker) Subscribe(ctx context.Context) error {
	log := w.log.With(zap.String("method", "Subscribe"))

	log.Info("subscribing to save commit events")

	for i := 0; i < w.workerCount; i++ {
		go func(workerID int) {
			log := log.With(zap.Int("worker_id", workerID))
			log.Info("saver worker subscribing to save commit events")

			err := w.eventBus.Subscribe(ctx, events.SaveCommitEventTopic, w.handleEvent)
			if err != nil {
				log.Error("subscription error", zap.Error(err))
			}
		}(i)
	}
	return nil
}

func (w *worker) handleEvent(msg []byte) error {
	var event events.SaveCommitEvent
	if err := json.Unmarshal(msg, &event); err != nil {
		return fmt.Errorf("failed to unmarshal save commit event: %w", err)
	}
	return w.handleSaveCommitEvent(context.Background(), event)
}

func (w *worker) handleSaveCommitEvent(ctx context.Context, event events.SaveCommitEvent) error {
	log := w.log.With(
		zap.String("method", "handleSaveCommitEvent"),
		zap.String("taskID", event.TaskID),
		zap.Int("commitCount", len(event.Commits)),
	)

	log.Info("received save commit event")

	if len(event.Commits) == 0 {
		log.Info("no commits to save")
		return nil
	}

	batchLatestCommitTime := event.Commits[0].Date
	batchOldestCommitTime := event.Commits[len(event.Commits)-1].Date

	log.Info("saving commits", zap.Time("batchLatestCommitTime", batchLatestCommitTime), zap.Time("batchOldestCommitTime", batchOldestCommitTime))

	if err := w.repoSvc.UpdateTrackingInfo(ctx, event.RepoInfo, batchLatestCommitTime); err != nil {
		log.Error("failed to update repo tracking info", zap.Error(err))
		return fmt.Errorf("failed to update repo tracking info: %w", err)
	}

	if err := w.commitSvc.CreateBatch(ctx, event.Commits); err != nil {
		log.Error("failed to save commits", zap.Error(err))
		return fmt.Errorf("failed to save commits: %w", err)
	}

	log.Info("commits saved successfully")
	return nil
}
