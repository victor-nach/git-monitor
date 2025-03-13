package events

import (
	"time"

	"github.com/victor-nach/git-monitor/internal/domain/models"
)

var (
	FetchCommitEventTopic = "fetch_commit_event"
	SaveCommitEventTopic  = "save_commit_event"
)

type (
	FetchCommitEvent struct {
		models.RepoInfo
		TaskID string
		RepoID string
		Since  time.Time
	}

	SaveCommitEvent struct {
		models.RepoInfo
		TaskID  string
		Commits []models.Commit
	}
)
