package githubclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	rerrors "errors"

	"github.com/victor-nach/git-monitor/internal/domain/errors"
	"github.com/victor-nach/git-monitor/pkg/githubclient/dto"
	"go.uber.org/zap"
)

const (
	defaultBaseURL    = "https://api.github.com"
	defaultTimeout    = 45 * time.Second
	defaultRetryCount = 3
	defaultRetryDelay = 2 * time.Second
)

type (
	HTTPClient interface {
		Do(req *http.Request) (*http.Response, error)
	}

	Config struct {
		BaseURL    *string
		HTTPClient HTTPClient
		RetryCount *int
		RetryDelay *time.Duration
		Timeout    *time.Duration
	}

	client struct {
		token      string
		baseURL    string
		httpClient HTTPClient
		retryCount int
		retryDelay time.Duration
		log        *zap.Logger
	}
)

// New creates a new GitHub client
func New(token string, logger *zap.Logger, cfg ...*Config) *client {
	client := &client{
		token:      token,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: defaultTimeout},
		retryCount: defaultRetryCount,
		retryDelay: defaultRetryDelay,
		log:        logger,
	}

	if len(cfg) > 0 {
		client.applyConfig(cfg[0])
	}

	return client
}

func (c *client) applyConfig(cfg *Config) {
	if cfg == nil {
		return
	}
	if cfg.BaseURL != nil {
		c.baseURL = *cfg.BaseURL
	}
	if cfg.HTTPClient != nil {
		c.httpClient = cfg.HTTPClient
	}
	if cfg.RetryCount != nil {
		c.retryCount = *cfg.RetryCount
	}
	if cfg.RetryDelay != nil {
		c.retryDelay = *cfg.RetryDelay
	}
	if cfg.Timeout != nil {
		c.httpClient.(*http.Client).Timeout = *cfg.Timeout
	}
}

// GetRepository fetches repository details from GitHub
func (c *client) GetRepository(ctx context.Context, owner, repo string) (dto.GitHubRepositoryResponse, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", c.baseURL, owner, repo)
	var repoResponse dto.GitHubRepositoryResponse
	if err := c.doWithRetry(ctx, url, &repoResponse); err != nil {
		return dto.GitHubRepositoryResponse{}, err
	}
	return repoResponse, nil
}

// GetCommits fetches commits for a repository from GitHub
func (c *client) GetCommits(ctx context.Context, owner, repoName string, from, to time.Time, batchSize int, page int) ([]dto.GitHubCommitResponse, error) {
	log := c.log.With(
		zap.String("method", "GetCommits"),
		zap.String("owner", owner),
		 zap.String("repo", repoName),
		 zap.Time("from", from),
		 zap.Time("to", to),
		 zap.Int("batchSize", batchSize),
		 zap.Int("page", page),
		)

	log.Info("fetching commits from github")

	url := fmt.Sprintf("%s/repos/%s/%s/commits", c.baseURL, owner, repoName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	if !from.IsZero() {
		q.Add("since", from.Format(time.RFC3339))
	}
	q.Add("until", to.Format(time.RFC3339))
	q.Add("per_page", strconv.Itoa(batchSize))
	if page > 0 {
		q.Add("page", strconv.Itoa(page))
	}
	req.URL.RawQuery = q.Encode()

	var commits []dto.GitHubCommitResponse
	if err := c.doWithRetry(ctx, req.URL.String(), &commits); err != nil {
		return nil, err
	}

	log.Info("fetched commits from github", zap.Int("count", len(commits)))

	return commits, nil
}

func (c *client) doWithRetry(ctx context.Context, url string, result interface{}) error {
	var err error
	for i := 0; i < c.retryCount; i++ {
		err = c.do(ctx, url, result)
		if err == nil {
			return nil
		}

		if rerrors.Is(err, errors.ErrRateLimitExceeded) || rerrors.Is(err, errors.ErrInternalServer) {
			c.log.Warn("retrying request", zap.String("url", url), zap.Error(err), zap.Int("attempt", i+1))
			time.Sleep(c.retryDelay)
			continue
		}

		return err
	}
	return fmt.Errorf("request failed after %d retries: %w", c.retryCount, err)
}

func (c *client) do(ctx context.Context, url string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return errors.ErrInvalidResponse
		}
		return nil
	case http.StatusNotFound:
		return errors.ErrRepositoryNotFound
	case http.StatusUnauthorized:
		return errors.ErrUnauthorized
	case http.StatusForbidden:
		return errors.ErrRateLimitExceeded
	case http.StatusInternalServerError:
		return errors.ErrInternalServer
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
