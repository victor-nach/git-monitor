package scheduler

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type (
	service struct {
		log *zap.Logger
		tasksService tasksService
		interval time.Duration
	}

	tasksService interface {
		StartTasks(ctx context.Context) error
	}
)

func New(log *zap.Logger, tasksService tasksService, interval time.Duration) *service {
	log = log.With(zap.String("service", "scheduler"))

	return &service{
		log: 		log,
		tasksService: tasksService,
		interval: interval,
	}
}

func (s *service) Start(ctx context.Context) {
	log := s.log.With(zap.String("method", "Start"))

	log.Info("starting scheduler")

	log.Info("starting tasks")
	if err := s.tasksService.StartTasks(ctx); err != nil {
		log.Error("error starting tasks", zap.Error(err))
	}
	
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("scheduler stopped")
			return

		case <-ticker.C:
			log.Info("starting tasks")
			if err := s.tasksService.StartTasks(ctx); err != nil {
				log.Error("error starting tasks", zap.Error(err))
			}
		}
	}
}
