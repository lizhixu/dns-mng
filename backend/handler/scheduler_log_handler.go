package handler

import (
	"net/http"
	"strconv"

	"dns-mng/service"

	"github.com/gin-gonic/gin"
)

type SchedulerLogHandler struct {
	schedulerLogService *service.SchedulerLogService
	schedulerService    *service.SchedulerService
}

func NewSchedulerLogHandler(schedulerLogService *service.SchedulerLogService, schedulerService *service.SchedulerService) *SchedulerLogHandler {
	return &SchedulerLogHandler{
		schedulerLogService: schedulerLogService,
		schedulerService:    schedulerService,
	}
}

func (h *SchedulerLogHandler) GetSchedulerLogs(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)

	pageSizeStr := c.DefaultQuery("page_size", "20")
	pageSize, _ := strconv.Atoi(pageSizeStr)

	response, err := h.schedulerLogService.GetSchedulerLogs(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get scheduler logs"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *SchedulerLogHandler) GetSchedulerLogsByTask(c *gin.Context) {
	taskName := c.Param("taskName")
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)

	pageSizeStr := c.DefaultQuery("page_size", "20")
	pageSize, _ := strconv.Atoi(pageSizeStr)

	response, err := h.schedulerLogService.GetSchedulerLogsByTask(taskName, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get scheduler logs"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *SchedulerLogHandler) TriggerManualCheck(c *gin.Context) {
	h.schedulerService.TriggerManualCheck()
	c.JSON(http.StatusOK, gin.H{"message": "Manual check triggered successfully"})
}
