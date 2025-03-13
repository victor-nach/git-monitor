package fetcher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/victor-nach/git-monitor/internal/domain/events"
	"github.com/victor-nach/git-monitor/internal/domain/models"
	"github.com/victor-nach/git-monitor/internal/http/utils"
	"go.uber.org/zap"
)

type (
	worker struct {
		log           *zap.Logger
		githubService githubService
		taskService   taskService
		eventBus      eventBus
		workerCount   int
	}

	githubService interface {
		GetCommitsStream(ctx context.Context, request models.GetCommitsStreamRequest) models.GetCommitsStreamResponse
	}

	taskService interface {
		UpdateStatus(ctx context.Context, taskID string, status string, errMsg *string) error
	}

	eventBus interface {
		Publish(ctx context.Context, topic string, message interface{}) error
		Subscribe(ctx context.Context, topic string, handler func(message []byte) error) error
	}
)

func New(log *zap.Logger, githubService githubService, taskService taskService, eventBus eventBus, workerCount int) *worker {
	log = log.With(zap.String("worker", "fetcher"))

	return &worker{
		log:           log,
		githubService: githubService,
		eventBus:      eventBus,
		taskService:   taskService,
		workerCount:   workerCount,
	}
}

func (w *worker) Subscribe(ctx context.Context) error {
	log := w.log.With(zap.String("method", "Subscribe"))

	log.Info("subscribing to fetch commit events")

	for i := 0; i < w.workerCount; i++ {
		go func(workerID int) {
			log := log.With(zap.Int("worker_id", workerID))
			log.Info("fetcher worker subscribing to fetch commit events")

			err := w.eventBus.Subscribe(ctx, events.FetchCommitEventTopic, w.handleEvent)
			if err != nil {
				log.Error("subscription error", zap.Error(err))
			}
		}(i)
	}
	return nil
}

func (w *worker) handleEvent(msg []byte) error {
	var event events.FetchCommitEvent
	if err := json.Unmarshal(msg, &event); err != nil {
		return fmt.Errorf("failed to unmarshal fetch commit event: %w", err)
	}
	return w.handleFetchCommitEvent(context.Background(), event)
}

func (w *worker) handleFetchCommitEvent(ctx context.Context, event events.FetchCommitEvent) error {
	log := w.log.With(zap.String("method", "handleFetchCommitEvent"),
		zap.String("taskID", event.TaskID),
		zap.String("repoID", event.RepoID),
	)
	log = utils.WithRepoInfo(log, event.RepoInfo)

	log.Info("received fetch commit event")

	req := models.GetCommitsStreamRequest{
		RepoID:   event.RepoID,
		RepoInfo: event.RepoInfo,
		Since:    &event.Since,
	}

	resp := w.githubService.GetCommitsStream(ctx, req)

	for {
		select {
		case commits, ok := <-resp.DataChan:
			if !ok {
				log.Info("commit data channel closed")
				goto COMPLETE
			}

			log.Info("fetched a batch of commits", zap.Int("commitCount", len(commits)))

			// Publish the save commit event
			saveEvent := events.SaveCommitEvent{
				RepoInfo: event.RepoInfo,
				TaskID:  event.TaskID,
				Commits: commits,
			}
			if err := w.eventBus.Publish(ctx, events.SaveCommitEventTopic, saveEvent); err != nil {
				log.Error("failed to publish save commit event", zap.Error(err))
				return fmt.Errorf("failed to publish save commit event: %w", err)
			}

		case err, ok := <-resp.ErrChan:
			if !ok {
				log.Info("error channel closed")
				goto COMPLETE
			}
			log.Error("error fetching commit batch", zap.Error(err))
			return fmt.Errorf("error fetching commit batch: %w", err)

		case <-resp.DoneChan:
			log.Info("commit streaming completed")
			goto COMPLETE

		case <-ctx.Done():
			log.Info("context cancelled, aborting fetch commit event", zap.Error(ctx.Err()))
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		}
	}

COMPLETE:
	if err := w.taskService.UpdateStatus(ctx, event.TaskID, models.TaskStatusCompleted, nil); err != nil {
		log.Error("failed to update job status", zap.Error(err))
		return fmt.Errorf("failed to update job status: %w", err)
	}

	log.Info("fetch commit event completed successfully")
	return nil
}
