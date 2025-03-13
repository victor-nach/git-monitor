package dto

import "time"

type (
	GitHubRepositoryResponse struct {
		ID            int       `json:"id"`
		Owner         Owner     `json:"owner"`
		Name          string    `json:"name"`
		Description   string    `json:"description"`
		URL           string    `json:"html_url"`
		Language      string    `json:"language"`
		ForksCount    int       `json:"forks_count"`
		StarsCount    int       `json:"stargazers_count"`
		OpenIssues    int       `json:"open_issues_count"`
		WatchersCount int       `json:"watchers_count"`
		CreatedAt     time.Time `json:"created_at"`
		UpdatedAt     time.Time `json:"updated_at"`
	}

	Owner struct {
		Login string `json:"login"`
	}

	GitHubCommitResponse struct {
		SHA     string `json:"sha"`
		Commit  Commit `json:"commit"`
		HTMLURL string `json:"html_url"`
	}

	Commit struct {
		Message string `json:"message"`
		Author  Author `json:"author"`
	}

	Author struct {
		Name  string    `json:"name"`
		Email string    `json:"email"`
		Date  time.Time `json:"date"`
	}
)
