package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	RepoPrefix   = "repo"
	TaskPrefix   = "task"
	CommitPrefix = "commit"
)

func NewUUIDWithPrefix(prefix string) string {
	id := strings.ReplaceAll(uuid.NewString(), "-", "")
	return fmt.Sprintf("%s-%s", prefix, id)
}

const (
	TaskStatusPending    = "pending"
	TaskStatusInProgress = "in_progress"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"
)

type (
	Repository struct {
		ID                      string     `json:"id"`
		RepoID                  int        `json:"repo_id"`
		Name                    string     `json:"name"`
		Owner                   string     `json:"owner"`
		Description             string     `json:"description"`
		URL                     string     `json:"url"`
		Language                string     `json:"language"`
		ForksCount              int        `json:"forks_count"`
		StarsCount              int        `json:"stars_count"`
		OpenIssues              int        `json:"open_issues"`
		WatchersCount           int        `json:"watchers_count"`
		IsSyncedToStartTime     bool       `json:"is_synced_to_start_time"`
		IsActive                bool       `json:"is_active"`
		CommitTrackingStartTime time.Time  `json:"commit_tracking_start_time"`
		LastFetchedAt           *time.Time `json:"last_fetched_at"`
		LastFetchedCommitTime   *time.Time `json:"last_fetched_commit_time"`
		RepoCreatedAt           time.Time  `json:"repo_created_at"`
		RepoUpdatedAt           time.Time  `json:"repo_updated_at"`
		CreatedAt               time.Time  `json:"created_at"`
		UpdatedAt               *time.Time `json:"updated_at"`
	}

	Commit struct {
		ID           string     `json:"id"`
		SHA          string     `json:"sha"`
		RepositoryID string     `json:"repository_id"`
		RepoName     string     `json:"repo_name"`
		RepoOwner    string     `json:"repo_owner"`
		Message      string     `json:"message"`
		Author       string     `json:"author"`
		AuthorEmail  string     `json:"author_email"`
		Date         time.Time  `json:"date"`
		URL          string     `json:"url"`
		CreatedAt    time.Time  `json:"created_at"`
		UpdatedAt    *time.Time `json:"updated_at"`
	}

	BatchDetail struct {
		BatchID      int       `json:"batch_id"`
		StartTime    time.Time `json:"start_time"`
		EndTime      time.Time `json:"end_time"`
		Status       string    `json:"status"`
		ErrorMessage string    `json:"error_message"`
	}

	Task struct {
		ID           string     `json:"id"`
		RepositoryID string     `json:"repository_id"`
		RepoName     string     `json:"repo_name"`
		RepoOwner    string     `json:"repo_owner"`
		Status       string     `json:"status"`
		CompletedAt  *time.Time `json:"completed_at"`
		ErrorMessage string     `json:"error_message"`
		CreatedAt    time.Time  `json:"created_at"`
		UpdatedAt    *time.Time `json:"updated_at"`
	}

	PaginationReq struct {
		Cursor string `json:"cursor"`
		Limit  int    `json:"limit"`
	}

	CommitsResponse struct {
		Pagination Pagination `json:"pagination"` // Pagination information
		Commits    []Commit   `json:"commits"`    // List of commits

	}

	Pagination struct {
		TotalCount     int64  `json:"total_count"`     // Total number of commits in the repository
		NextCursor     string `json:"next_cursor"`     // Cursor for the next page
		PreviousCursor string `json:"previous_cursor"` // Cursor for the previous page
		Limit          int    `json:"limit"`           // Number of commits per page
	}

	CommitResponse struct {
		Commits []Commit `json:"commits"`
		Next    string   `json:"next"`
	}

	AuthorStats struct {
		Author  string `json:"author"`
		Commits int    `json:"commits"`
	}

	RepoInfo struct {
		Name  string `json:"name"`
		Owner string `json:"owner"`
	}

	GetCommitsStreamRequest struct {
		RepoID   string     `json:"repo_id"`
		RepoInfo RepoInfo   `json:"repo_info"`
		Since    *time.Time `json:"since"`
		Until    *time.Time `json:"until"`
	}

	GetCommitsStreamResponse struct {
		DataChan <-chan []Commit
		ErrChan  <-chan error
		DoneChan <-chan struct{}
	}
)
