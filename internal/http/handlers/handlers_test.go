package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/victor-nach/git-monitor/internal/domain/models"
	"github.com/victor-nach/git-monitor/internal/http/handlers/mocks"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestAddTrackedRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepoSvc := mocks.NewMockrepoSvc(ctrl)
	mockCommitSvc := mocks.NewMockcommitSvc(ctrl)
	mockTaskSvc := mocks.NewMocktaskSvc(ctrl)

	log := zap.NewNop()
	h := New(log, mockRepoSvc, mockCommitSvc, mockTaskSvc)

	repoInfo := models.RepoInfo{Name: "test-repo", Owner: "owner"}

	mockRepoSvc.EXPECT().Create(gomock.Any(), repoInfo, gomock.Any()).Return(models.Repository{Name: "test-repo"}, "task-id", nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/repos", nil)
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), repoInfoKey, repoInfo))

	h.AddTrackedRepository(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Repository added successfully")
}

func TestListTrackedRepositories(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepoSvc := mocks.NewMockrepoSvc(ctrl)
	mockCommitSvc := mocks.NewMockcommitSvc(ctrl)
	mockTaskSvc := mocks.NewMocktaskSvc(ctrl)

	log := zap.NewNop()
	h := New(log, mockRepoSvc, mockCommitSvc, mockTaskSvc)

	repos := []models.Repository{
		{Name: "repo1", Owner: "owner1"},
		{Name: "repo2", Owner: "owner2"},
	}

	mockRepoSvc.EXPECT().List(gomock.Any()).Return(repos, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/repos", nil)

	h.ListTrackedRepositories(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Tracked repositories listed successfully")
	assert.Contains(t, w.Body.String(), "repo1")
	assert.Contains(t, w.Body.String(), "repo2")
}

func TestResetRepo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepoSvc := mocks.NewMockrepoSvc(ctrl)
	mockCommitSvc := mocks.NewMockcommitSvc(ctrl)
	mockTaskSvc := mocks.NewMocktaskSvc(ctrl)

	log := zap.NewNop()
	h := New(log, mockRepoSvc, mockCommitSvc, mockTaskSvc)

	repoInfo := models.RepoInfo{Name: "test-repo", Owner: "owner"}
	taskID := "task-id"

	mockRepoSvc.EXPECT().Reset(gomock.Any(), repoInfo, gomock.Any()).Return(taskID, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/repos/reset", nil)
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), repoInfoKey, repoInfo))

	h.ResetRepo(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Repository reset successfully")
	assert.Contains(t, w.Body.String(), taskID)
}

func TestGetTopCommitAuthors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepoSvc := mocks.NewMockrepoSvc(ctrl)
	mockCommitSvc := mocks.NewMockcommitSvc(ctrl)
	mockTaskSvc := mocks.NewMocktaskSvc(ctrl)

	log := zap.NewNop()
	h := New(log, mockRepoSvc, mockCommitSvc, mockTaskSvc)

	repoInfo := models.RepoInfo{Name: "test-repo", Owner: "owner"}
	limit := 5
	stats := []models.AuthorStats{
		{Author: "user1", Commits: 10},
		{Author: "user2", Commits: 8},
	}

	mockCommitSvc.EXPECT().GetTopAuthors(gomock.Any(), repoInfo, limit).Return(stats, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/repos/top-authors?limit=5", nil)
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), repoInfoKey, repoInfo))

	h.GetTopCommitAuthors(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Top commit authors retrieved successfully")
	assert.Contains(t, w.Body.String(), "user1")
	assert.Contains(t, w.Body.String(), "user2")
}

func TestListCommits(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepoSvc := mocks.NewMockrepoSvc(ctrl)
	mockCommitSvc := mocks.NewMockcommitSvc(ctrl)
	mockTaskSvc := mocks.NewMocktaskSvc(ctrl)

	log := zap.NewNop()
	h := New(log, mockRepoSvc, mockCommitSvc, mockTaskSvc)

	repoInfo := models.RepoInfo{Name: "test-repo", Owner: "owner"}
	paginationReq := models.PaginationReq{Limit: 10, Cursor: "cursor1"}
	commits := []models.Commit{
		{SHA: "abc123", Message: "Initial commit"},
		{SHA: "def456", Message: "Add feature"},
	}
	nextCursor := "cursor2"

	mockCommitSvc.EXPECT().List(gomock.Any(), repoInfo, paginationReq).Return(commits, nextCursor, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/repos/commits?limit=10&cursor=cursor1", nil)
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), repoInfoKey, repoInfo))

	h.ListCommits(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Commits listed successfully")
	assert.Contains(t, w.Body.String(), "abc123")
	assert.Contains(t, w.Body.String(), "def456")
	assert.Contains(t, w.Body.String(), "cursor2") // Verify next cursor
}
