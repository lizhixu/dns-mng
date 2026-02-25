package handler

import (
	"net/http"
	"strconv"

	"dns-mng/service"

	"github.com/gin-gonic/gin"
)

type LogHandler struct {
	logService *service.LogService
}

func NewLogHandler(logService *service.LogService) *LogHandler {
	return &LogHandler{logService: logService}
}

func (h *LogHandler) GetLogs(c *gin.Context) {
	userID := c.GetInt64("user_id")
	
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	
	logs, err := h.logService.GetLogs(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logs"})
		return
	}
	
	c.JSON(http.StatusOK, logs)
}
