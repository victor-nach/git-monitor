package github

import (
	"context"
	"sync"
	"time"

	"github.com/victor-nach/git-monitor/internal/domain/errors"
	"github.com/victor-nach/git-monitor/internal/domain/models"
	"github.com/victor-nach/git-monitor/pkg/utils"
	"github.com/victor-nach/git-monitor/pkg/githubclient/dto"
	"go.uber.org/zap"
)

type (
	service struct {
		log       *zap.Logger
		client    githubClient
		batchSize int
	}

	githubClient interface {
		GetRepository(ctx context.Context, owner, repoName string) (dto.GitHubRepositoryResponse, error)
		GetCommits(ctx context.Context, owner, repoName string, from, to time.Time, batchSize int, page int) ([]dto.GitHubCommitResponse, error)
	}
)

func New(log *zap.Logger, client githubClient, batchSize int) *service {
	return &service{
		log:       log,
		client:    client,
		batchSize: batchSize,
	}
}

func (s *service) GetRepository(ctx context.Context, RepoInfo models.RepoInfo) (models.Repository, error) {
	log := s.log.With(
		zap.String("owner", RepoInfo.Name),
		zap.String("repoName", RepoInfo.Name))

	log.Info("getting repository from github")

	repo, err := s.client.GetRepository(ctx, RepoInfo.Owner, RepoInfo.Name)
	if err != nil {
		log.Error("failed to get repository", zap.Error(err))
		return models.Repository{}, err
	}

	newRepoID := models.NewUUIDWithPrefix(models.RepoPrefix)
	log = log.With(zap.String("repoID", newRepoID))

	log.Info("successfully retrieved repository from github")

	return mapToRepository(newRepoID, repo), nil
}

func (s *service) GetCommitsStream(ctx context.Context, request models.GetCommitsStreamRequest) models.GetCommitsStreamResponse {
	log := s.log.With(zap.String("method", "GetCommitsStream"))
	log = utils.WithRepoInfo(log, request.RepoInfo)

	log.Info("getting commits stream from github",
		zap.String("repo_owner", request.RepoInfo.Owner),
		zap.String("repo_name", request.RepoInfo.Name),
		zap.Any("since", request.Since),
		zap.Any("until", request.Until),
		zap.Int("batch_size", s.batchSize),
	)

	dataChan := make(chan []models.Commit)
	errChan := make(chan error)
	doneChan := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer func() {
			// Signal completion.
			doneChan <- struct{}{}
		}()

		untilVal := s.determineUntil(request)
		sinceVal := s.determineSince(request)
		s.streamCommits(ctx, log, request.RepoInfo, request.RepoID, sinceVal, untilVal, dataChan, errChan)
	}()

	go func() {
		wg.Wait()
		close(dataChan)
		close(errChan)
		close(doneChan)
		log.Info("closed data, error and done channels")
	}()

	return models.GetCommitsStreamResponse{
		DataChan: dataChan,
		ErrChan:  errChan,
		DoneChan: doneChan,
	}
}

func (s *service) determineUntil(request models.GetCommitsStreamRequest) time.Time {
	if request.Until != nil {
		return *request.Until
	}
	return time.Now()
}

func (s *service) determineSince(request models.GetCommitsStreamRequest) time.Time {
	if request.Since != nil {
		return *request.Since
	}
	return time.Time{}
}

func (s *service) streamCommits(ctx context.Context, log *zap.Logger, repoInfo models.RepoInfo, repoID string, since, until time.Time, dataChan chan<- []models.Commit, errChan chan<- error) {
	currentPage := 1

	for {
		select {
		case <-ctx.Done():
			log.Info("context done, stopping commit stream", zap.Error(ctx.Err()))
			return
		default:
			log.Info("fetching commits batch",
				zap.Int("page", currentPage),
				zap.Any("since", since),
				zap.Any("until", until),
			)

			commitsDTOs, err := s.client.GetCommits(ctx, repoInfo.Owner, repoInfo.Name, since, until, s.batchSize, currentPage)
			if err != nil {
				batchErr := errors.NewBatchError(nil, s.batchSize, err)
				errChan <- batchErr
				return
			}

			if len(commitsDTOs) == 0 {
				log.Info("no more commits to fetch from github")
				return
			}

			log.Info("retrieved commits batch from github",
				zap.Int("count", len(commitsDTOs)),
				zap.Any("first_commit_date", commitsDTOs[0].Commit.Author.Date),
				zap.Any("last_commit_date", commitsDTOs[len(commitsDTOs)-1].Commit.Author.Date),
			)

			domainCommits := mapCommits(repoInfo, repoID, commitsDTOs)
			dataChan <- domainCommits

			if len(commitsDTOs) < s.batchSize {
				log.Info("successfully retrieved all commit streams from github")
				return
			}

			currentPage++
		}
	}
}

func mapToRepository(id string, dto dto.GitHubRepositoryResponse) models.Repository {
	return models.Repository{
		ID:                      id,
		RepoID:                  dto.ID,
		Name:                    dto.Name,
		Owner:                   dto.Owner.Login,
		Description:             dto.Description,
		URL:                     dto.URL,
		Language:                dto.Language,
		ForksCount:              dto.ForksCount,
		StarsCount:              dto.StarsCount,
		OpenIssues:              dto.OpenIssues,
		WatchersCount:           dto.WatchersCount,
		IsSyncedToStartTime:     false,
		IsActive:                true,
		LastFetchedAt:           nil,
		RepoCreatedAt:           dto.CreatedAt,
		RepoUpdatedAt:           dto.UpdatedAt,
		CreatedAt:               time.Now(),
		UpdatedAt:               nil,
	}
}

func mapCommits(repoInfo models.RepoInfo, repoID string, dtos []dto.GitHubCommitResponse) []models.Commit {
	commits := make([]models.Commit, len(dtos))
	for i, v := range dtos {
		id := models.NewUUIDWithPrefix(models.CommitPrefix)
		commits[i] = mapToCommit(id, repoInfo, repoID, v)
	}
	return commits
}

func mapToCommit(id string, repoInfo models.RepoInfo, repoID string, dto dto.GitHubCommitResponse) models.Commit {
	return models.Commit{
		ID:           id,
		SHA:          dto.SHA,
		RepositoryID: repoID,
		RepoName:     repoInfo.Name,
		RepoOwner:   repoInfo.Owner,
		Message:      dto.Commit.Message,
		URL:          dto.HTMLURL,
		Author:       dto.Commit.Author.Name,
		AuthorEmail:  dto.Commit.Author.Email,
		Date:         dto.Commit.Author.Date,
		CreatedAt:    time.Now(),
		UpdatedAt:    nil,
	}
}
