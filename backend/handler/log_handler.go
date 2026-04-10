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
	
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	
	pageSizeStr := c.DefaultQuery("page_size", "20")
	pageSize, _ := strconv.Atoi(pageSizeStr)
	
	response, err := h.logService.GetLogs(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logs"})
		return
	}
	
	c.JSON(http.StatusOK, response)
}
