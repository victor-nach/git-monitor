package githubclient

import (
	"context"
	ierrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/victor-nach/git-monitor/internal/domain/errors"
	"github.com/victor-nach/git-monitor/pkg/githubclient/dto"
	"go.uber.org/zap"
)

func TestGetRepository(t *testing.T) {
	successResp := `{
		"id": 1296269,
		"owner": {
			"login": "octocat"
		},
		"name": "Hello-World",
		"description": "This your first repo!",
		"html_url": "https://github.com/octocat/Hello-World",
		"language": "JavaScript",
		"forks_count": 500,
		"stargazers_count": 1000,
		"open_issues_count": 10,
		"watchers_count": 1000,
		"created_at": "2011-01-26T19:01:12Z",
		"updated_at": "2023-10-01T14:42:30Z"
	}`

	expectedRepository := dto.GitHubRepositoryResponse{
		ID:            1296269,
		Owner:         dto.Owner{Login: "octocat"},
		Name:          "Hello-World",
		Description:   "This your first repo!",
		URL:           "https://github.com/octocat/Hello-World",
		Language:      "JavaScript",
		ForksCount:    500,
		StarsCount:    1000,
		OpenIssues:    10,
		WatchersCount: 1000,
		CreatedAt:     time.Date(2011, 1, 26, 19, 1, 12, 0, time.UTC),
		UpdatedAt:     time.Date(2023, 10, 1, 14, 42, 30, 0, time.UTC),
	}

	tests := []struct {
		name           string
		responseStatus int
		responseBody   string
		expectedResult dto.GitHubRepositoryResponse
		expectedError  error
	}{
		{
			name:           "Success - Repository found",
			responseStatus: http.StatusOK,
			responseBody:   successResp,
			expectedResult: expectedRepository,
			expectedError:  nil,
		},
		{
			name:           "Error - Repository not found",
			responseStatus: http.StatusNotFound,
			responseBody:   `{}`,
			expectedResult: dto.GitHubRepositoryResponse{},
			expectedError:  errors.ErrRepositoryNotFound,
		},
		{
			name:           "Error - Unauthorized",
			responseStatus: http.StatusUnauthorized,
			responseBody:   `{}`,
			expectedResult: dto.GitHubRepositoryResponse{},
			expectedError:  errors.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			cfg := &Config{
				BaseURL: &server.URL,
			}
			client := New("test-token",  zap.NewNop(), cfg)

			result, err := client.GetRepository(context.Background(), "octocat", "Hello-World")

			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestGetCommits(t *testing.T) {

	successResp := `[
		{
			"sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			"commit": {
				"message": "Fix bug in authentication module",
				"author": {
					"name": "John Doe",
					"email": "john@example.com",
					"date": "2023-10-01T12:00:00Z"
				}
			},
			"html_url": "https://github.com/octocat/Hello-World/commit/6dcb09b5b57875f334f61aebed695e2e4193db5e"
		}
	]`

	expectedCommits := []dto.GitHubCommitResponse{
		{
			SHA: "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			Commit: dto.Commit{
				Message: "Fix bug in authentication module",
				Author: dto.Author{
					Name:  "John Doe",
					Email: "john@example.com",
					Date:  time.Date(2023, 10, 1, 12, 0, 0, 0, time.UTC),
				},
			},
			HTMLURL: "https://github.com/octocat/Hello-World/commit/6dcb09b5b57875f334f61aebed695e2e4193db5e",
		},
	}

	tests := []struct {
		name           string
		responseStatus int
		responseBody   string
		expectedResult []dto.GitHubCommitResponse
		expectedError  error
	}{
		{
			name:           "Success - Commits found",
			responseStatus: http.StatusOK,
			responseBody:   successResp,
			expectedResult: expectedCommits,
			expectedError:  nil,
		},
		{
			name:           "Error - Repository not found",
			responseStatus: http.StatusNotFound,
			responseBody:   `{}`,
			expectedResult: nil,
			expectedError:  errors.ErrRepositoryNotFound,
		},
		{
			name:           "Error - Rate limit exceeded",
			responseStatus: http.StatusForbidden,
			responseBody:   `{}`,
			expectedResult: nil,
			expectedError:  errors.ErrRateLimitExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			cfg := &Config{
				BaseURL: &server.URL,
			}
			client := New("test-token", zap.NewNop(), cfg)

			startTime := time.Now().AddDate(0, -1, 0)
			endTime := time.Now()
			result, err := client.GetCommits(context.Background(), "octocat", "Hello-World", startTime, endTime, 10, 1)

			// require.Equal(t, tt.expectedError, err)
			require.True(t, ierrors.Is(err, tt.expectedError))
			require.Equal(t, tt.expectedResult, result)
		})
	}
}
