package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"time"

	"log"

	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/victor-nach/git-monitor/config"

	dErrors "github.com/victor-nach/git-monitor/internal/domain/errors"
	"github.com/victor-nach/git-monitor/internal/domain/models"
	"github.com/victor-nach/git-monitor/pkg/logger"
	"github.com/victor-nach/git-monitor/pkg/migrator"
	"github.com/victor-nach/git-monitor/pkg/utils"

	_ "modernc.org/sqlite"
)

var (
	db      *gorm.DB
	sqlDB   *sql.DB
	testCtx = context.Background()
)

func TestMain(m *testing.M) {
	logr, err := logger.New(config.AppEnvDevelopment)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logr.Sync()

	sqlDB, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %v", err)
		return
	}

	db, err = gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to initialize GORM: %v", err)
		return
	}

	projectRoot, err := utils.GetBasePath()
	if err != nil {
		log.Fatalf("failed to get project root: %v", err)
	}

	mf := filepath.Join(projectRoot, "migrations")
	mf = filepath.ToSlash(mf)

	if err := migrator.Migrate(testCtx, sqlDB, mf, logr); err != nil {
		sqlDB.Close()
		log.Fatalf("failed to apply migrations: %v", err)
		return
	}

	code := m.Run()

	sqlDB.Close()
	os.Exit(code)
}

func TestRepoStore_Create(t *testing.T) {
	repoStore := &repoStore{db: db}
	testRepo := models.Repository{
		Name:  "test-repo",
		Owner: "victor",
	}

	err := repoStore.Create(testCtx, testRepo)
	assert.NoError(t, err, "should create repository without error")

	// verify creation
	var count int64
	db.Model(&models.Repository{}).Where("name = ? AND owner = ?", "test-repo", "victor").Count(&count)
	assert.Equal(t, int64(1), count, "should have exactly one repository")
}

func TestRepoStore_Get(t *testing.T) {
	repoStore := &repoStore{db: db}
	// setup: create a repo first
	repo := models.Repository{ID: "123456C", Name: "get-repo", Owner: "tester", RepoID: 1234}
	db.Create(&repo)

	result, err := repoStore.Get(testCtx, models.RepoInfo{Name: "get-repo", Owner: "tester"})
	assert.NoError(t, err)
	assert.Equal(t, "get-repo", result.Name)
	assert.Equal(t, "tester", result.Owner)

	// test non-existent repo
	_, err = repoStore.Get(testCtx, models.RepoInfo{Name: "unknown", Owner: "tester"})
	assert.ErrorIs(t, err, dErrors.ErrRepositoryNotFound)
}

func TestRepoStore_Reset(t *testing.T) {
	repoStore := &repoStore{db: db}

	testRepo := models.Repository{
		ID:                      uuid.NewString(),
		Name:                    "reset-repo",
		Owner:                   "tester",
		RepoID:                  12347,
		CommitTrackingStartTime: time.Time{},
		LastFetchedAt:           nil,
	}
	db.Create(&testRepo)
	db.Create(&models.Commit{RepoName: "reset-repo", RepoOwner: "tester"})

	startTime := time.Now()
	err := repoStore.Reset(testCtx, models.RepoInfo{Name: "reset-repo", Owner: "tester"}, &startTime)
	assert.NoError(t, err)

	// validate reset
	var commitsCount int64
	db.Model(&models.Commit{}).Where("repo_name = ? AND repo_owner = ?", "reset-repo", "tester").Count(&commitsCount)
	assert.Equal(t, int64(0), commitsCount, "commits should be deleted")

	var updatedRepo models.Repository
	db.Where("name = ? AND owner = ?", "reset-repo", "tester").First(&updatedRepo)
	assert.NotNil(t, updatedRepo.CommitTrackingStartTime)
	assert.Nil(t, updatedRepo.LastFetchedAt)
}

func TestCommitStore_GetTopAuthors(t *testing.T) {
	commitStore := &commitStore{db: db}

	// Setup
	commits := []models.Commit{
		{ID: "12334sfjh", SHA: "asdasf", Author: "author1", RepoName: "repo1", RepoOwner: "owner1"},
		{ID: "12334asbg", SHA: "asdafj", Author: "author1", RepoName: "repo1", RepoOwner: "owner1"},
		{ID: "12334sdfg", SHA: "asdalhj", Author: "author2", RepoName: "repo1", RepoOwner: "owner1"},
	}
	db.Create(&commits)

	stats, err := commitStore.GetTopAuthors(testCtx, models.RepoInfo{Name: "repo1", Owner: "owner1"}, 1)
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, "author1", stats[0].Author)
	assert.Equal(t, int(2), stats[0].Commits)
}

func TestCommitStore_List(t *testing.T) {
	commitStore := &commitStore{db: db}

	// Setup
	date1 := time.Now().Add(-2 * time.Hour)
	date2 := time.Now().Add(-1 * time.Hour)
	commits := []models.Commit{
		{ID: "123456A", SHA: "hash1", Date: date1, RepoName: "repo-list", RepoOwner: "owner-list"},
		{ID: "12345D", SHA: "hash2", Date: date2, RepoName: "repo-list", RepoOwner: "owner-list"},
	}
	db.Create(&commits)

	fetchedCommits, nextCursor, err := commitStore.List(testCtx, models.RepoInfo{Name: "repo-list", Owner: "owner-list"}, models.PaginationReq{Limit: 1})
	assert.NoError(t, err)
	assert.Len(t, fetchedCommits, 1)
	assert.NotEmpty(t, nextCursor)
}

func TestCommitStore_CreateBatch(t *testing.T) {
	commitStore := &commitStore{db: db}

	commits := []models.Commit{
		{ID: uuid.NewString(), SHA: "hash11", RepoName: "repo-batch", RepoOwner: "owner-batch"},
		{ID: uuid.NewString(), SHA: "hash22", RepoName: "repo-batch", RepoOwner: "owner-batch"},
	}

	err := commitStore.CreateBatch(testCtx, commits)
	assert.NoError(t, err)

	var count int64
	db.Model(&models.Commit{}).Where("repo_name = ? AND repo_owner = ?", "repo-batch", "owner-batch").Count(&count)
	assert.Equal(t, int64(2), count)

	// Test idempotency
	err = commitStore.CreateBatch(testCtx, commits)
	assert.NoError(t, err)
	db.Model(&models.Commit{}).Where("repo_name = ? AND repo_owner = ?", "repo-batch", "owner-batch").Count(&count)
	assert.Equal(t, int64(2), count, "should still have only 2 commits due to OnConflict")
}
