package models

import (
	"github.com/victor-nach/git-monitor/internal/domain/models"
)

const (
	SuccessStatus = "success"
	ErrorStatus   = "error"

	ActiveStatus   = "active"
	InactiveStatus = "inactive"
)

type (
	APIResponse struct {
		Status     string      `json:"status"`
		Message    string      `json:"message"`
		Pagination *Pagination `json:"pagination,omitempty"`
		Data       interface{} `json:"data,omitempty"`
	}

	Pagination struct {
		NextCursor     string `json:"next_cursor,omitempty"`
		PreviousCursor string `json:"previous_cursor,omitempty"`
	}

	ListTasksResponse struct {
		Tasks []models.Task `json:"tasks"`
	}

	TaskResponse struct {
		TaskID  string `json:"task_id"`
	}

	CommitsResponse struct {
		Commits []models.Commit `json:"commits"`
	}

	TopAuthorsResponse struct {
		Authors []models.AuthorStats `json:"authors"`
	}

	RepositoryResponse struct {
		TaskID     string            `json:"task_id"`
		Repository models.Repository `json:"repository"`
	}
		
)
