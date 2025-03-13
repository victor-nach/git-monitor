package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/victor-nach/git-monitor/internal/http/errors"
	"github.com/victor-nach/git-monitor/internal/http/models"
	"github.com/victor-nach/git-monitor/internal/http/utils"
	"go.uber.org/zap"
)

func (h *Handler) GetTopCommitAuthors(c *gin.Context) {
	log := h.log.With(zap.String("method", "GetTopCommitAuthors"))

	repoInfo, err := GetRepoInfo(c.Request.Context())
	if err != nil {
		h.log.Error("failed to retrieve repository info from context", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.ErrInternalServer)
		return
	}
	log = utils.WithRepoInfo(log, repoInfo)
	limit := utils.ExtractLimit(c)
	log = log.With(zap.Int("limit", limit))

	log.Info("handling Get top commit authors API request")

	stats, err := h.commitSvc.GetTopAuthors(c.Request.Context(), repoInfo, limit)
	if err != nil {
		log.Error("failed to get top commit authors", zap.Error(err))
		status, httpErr := errors.MapError(err)
		c.JSON(status, httpErr)
		return
	}

	log.Info("Top commit authors retrieved successfully", zap.Int("count", len(stats)))

	resp := models.APIResponse{
		Status:  models.SuccessStatus,
		Message: "Top commit authors retrieved successfully",
		Data:    stats,
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListCommits(c *gin.Context) {
	log := h.log.With(zap.String("method", "ListCommits"))

	repoInfo, err := GetRepoInfo(c.Request.Context())
	if err != nil {
		h.log.Error("failed to retrieve repository info from context", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.ErrInternalServer)
		return
	}
	log = utils.WithRepoInfo(log, repoInfo)

	paginationRq := utils.ExtractPaginationReq(c)
	log = log.With(
		zap.Int("limit", paginationRq.Limit),
		zap.String("cursor", paginationRq.Cursor))

	log.Info("handling List commits API request")

	commits, cursor, err := h.commitSvc.List(c.Request.Context(), repoInfo, paginationRq)
	if err != nil {
		log.Error("failed to list commits", zap.Error(err))
		status, httpErr := errors.MapError(err)
		c.JSON(status, httpErr)
		return
	}

	paginationResp := models.Pagination{
		NextCursor:     cursor,
		PreviousCursor: paginationRq.Cursor,
	}
	resp := models.APIResponse{
		Status:     models.SuccessStatus,
		Message:    "Commits listed successfully",
		Pagination: &paginationResp,
		Data:       commits,
	}

	log.Info("Commits listed successfully", zap.Int("count", len(commits)))

	c.JSON(http.StatusOK, resp)
}
