package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/victor-nach/git-monitor/internal/http/errors"
	"github.com/victor-nach/git-monitor/internal/http/models"
	"github.com/victor-nach/git-monitor/internal/http/utils"
	"go.uber.org/zap"
)

func (h *Handler) TriggerTask(c *gin.Context) {
	log := h.log.With(zap.String("method", "TriggerTask"))
	repoInfo, err := GetRepoInfo(c.Request.Context())
	if err != nil {
		h.log.Error("failed to retrieve repository info from context", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.ErrInternalServer)
		return
	}
	log = utils.WithRepoInfo(log, repoInfo)
	
	log.Info("handling trigger task API request")

	taskID, err := h.taskSvc.TriggerTask(c.Request.Context(), repoInfo, nil)
	if  err != nil {
		log.Error("failed to trigger task", zap.Error(err))
		status, httpErr := errors.MapError(err)
		c.JSON(status, httpErr)
		return
	}

	resp := models.APIResponse{
		Status:  models.SuccessStatus,
		Message: "Task triggered successfully",
		Data:    models.TaskResponse{TaskID: taskID},
	}

	log.Info("task triggered successfully", zap.String("task_id", taskID))
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListTasks(c *gin.Context) {
	log := h.log.With(zap.String("method", "ListTasks"))
	log.Info("handling list tasks API request")

	tasks, err := h.taskSvc.List(c.Request.Context())
	if err != nil {
		log.Error("failed to list tasks", zap.Error(err))
		status, httpErr := errors.MapError(err)
		c.JSON(status, httpErr)
		return
	}

	resp := models.APIResponse{
		Status:  models.SuccessStatus,
		Message: "Tasks retrieved successfully",
		Data:    tasks,
	}

	log.Info("tasks retrieved successfully")
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetTask(c *gin.Context) {
	log := h.log.With(zap.String("method", "GetTask"))

	taskID := c.Param("id")
	if taskID == "" {
		err := errors.ErrInputValidation("task id is required")
		log.Error("failed to get task", zap.Error(err))
		c.JSON(http.StatusBadRequest, err)
		return
	}
	log = log.With(zap.String("task_id", taskID))
	log.Info("handling get task API request")

	task, err := h.taskSvc.GetTask(c.Request.Context(), taskID)
	if err != nil {
		log.Error("failed to get task", zap.Error(err))
		status, httpErr := errors.MapError(err)
		c.JSON(status, httpErr)
		return
	}

	resp := models.APIResponse{
		Status:  models.SuccessStatus,
		Message: "Task retrieved successfully",
		Data:    task,
	}

	log.Info("task retrieved successfully")
	c.JSON(http.StatusOK, resp)
}
