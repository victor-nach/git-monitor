package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/victor-nach/git-monitor/internal/http/errors"
	"github.com/victor-nach/git-monitor/internal/http/models"
	"github.com/victor-nach/git-monitor/internal/http/utils"
	"go.uber.org/zap"
)

func (h *Handler) AddTrackedRepository(c *gin.Context) {
	log := h.log.With(zap.String("method", "AddTrackedRepository"))

	repoInfo, err := GetRepoInfo(c.Request.Context())
	if err != nil {
		h.log.Error("failed to retrieve repository info from context", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.ErrInternalServer)
		return
	}
	log = utils.WithRepoInfo(log, repoInfo)
	since, err := utils.ExtractTime(c, "since")
	if err != nil {
		log.Error("failed to extract time", zap.Error(err))
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if since != nil {
		log = log.With(zap.Time("since", *since))
	}

	log.Info("handling add tracked repository API request")

	repo, taskID, err := h.repoSvc.Create(c.Request.Context(), repoInfo, since)
	if err != nil {
		log.Error("failed to add tracked repository", zap.Error(err))
		status, httpErr := errors.MapError(err)
		httpErr.WithMessage("failed to add tracked repository")
		c.JSON(status, httpErr)
		return
	}

	repoResp := models.APIResponse{
		Status:  models.SuccessStatus,
		Message: "Repository added successfully",
		Data: models.RepositoryResponse{
			TaskID:     taskID,
			Repository: repo,
		},
	}

	log.Info("Repository added successfully", zap.String("taskID", taskID))
	c.JSON(http.StatusOK, repoResp)
}

func (h *Handler) ListTrackedRepositories(c *gin.Context) {
	log := h.log.With(zap.String("method", "ListTrackedRepositories"))

	log.Info("handling list tracked repositories API request")

	repos, err := h.repoSvc.List(c.Request.Context())
	if err != nil {
		log.Error("failed to list tracked repository", zap.Error(err))
		status, httpErr := errors.MapError(err)
		httpErr.WithMessage("failed to list tracked repository")
		c.JSON(status, httpErr)
		return
	}

	resp := models.APIResponse{
		Status:  models.SuccessStatus,
		Message: "Tracked repositories listed successfully",
		Data:    repos,
	}

	log.Info("Tracked repositories listed successfully", zap.Int("count", len(repos)))
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ResetRepo(c *gin.Context) {
	log := h.log.With(zap.String("method", "ResetRepo"))

	repoInfo, err := GetRepoInfo(c.Request.Context())
	if err != nil {
		h.log.Error("failed to retrieve repository info from context", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.ErrInternalServer)
		return
	}
	log = utils.WithRepoInfo(log, repoInfo)
	since, err := utils.ExtractTime(c, "since")
	if err != nil {
		log.Error("failed to extract time", zap.Error(err))
		c.JSON(http.StatusBadRequest, err)
		return
	}

	taskID, err := h.repoSvc.Reset(c.Request.Context(), repoInfo, since)
	if err != nil {
		log.Error("failed to reset repository", zap.Error(err))
		status, httpErr := errors.MapError(err)
		httpErr.WithMessage("failed to reset repository")
		c.JSON(status, httpErr)
		return
	}

	resp := models.APIResponse{
		Status:  models.SuccessStatus,
		Message: "Repository reset successfully",
		Data: models.TaskResponse{
			TaskID: taskID,
		},
	}

	log.Info("Repository reset successfully")
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) UpdateRepoStatus(c *gin.Context) {
	log := h.log.With(zap.String("method", "UpdateRepoStatus"))

	status := c.Query("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, errors.ErrInputValidation("tatus query param is required"))
		return
	}
	repoInfo, err := GetRepoInfo(c.Request.Context())
	if err != nil {
		h.log.Error("failed to retrieve repository info from context", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.ErrInternalServer)
		return
	}
	log = utils.WithRepoInfo(log, repoInfo)

	log = log.With(zap.String("status", status))

	log.Info("handling update repository status API request")

	isActive := new(bool)
	switch {
	case status == models.ActiveStatus:
		*isActive = true

	case status == models.InactiveStatus:
		*isActive = false

	default:
		err := errors.ErrInputValidation("a valid status query param is required")
		log.Error("failed to update repository status", zap.Error(err))
		c.JSON(http.StatusBadRequest, err)
		return
	}

	log.Info("handling update repository status API request")

	if err := h.repoSvc.UpdateStatus(c.Request.Context(), repoInfo, isActive); err != nil {
		log.Error("failed to update repository status", zap.Error(err))
		status, httpErr := errors.MapError(err)
		httpErr.WithMessage("failed to update repository status")
		c.JSON(status, httpErr)
		return
	}

	resp := models.APIResponse{
		Status:  models.SuccessStatus,
		Message: "Repository status updated successfully",
	}

	log.Info("Repository status updated successfully")
	c.JSON(http.StatusOK, resp)
}
