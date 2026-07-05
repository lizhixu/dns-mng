package handler

import (
	"dns-mng/middleware"
	"dns-mng/models"
	"dns-mng/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CFOptimizeHandler handles CF CDN optimization requests
type CFOptimizeHandler struct {
	cfOptimizeService *service.CFOptimizeService
}

func NewCFOptimizeHandler(cfOptimizeService *service.CFOptimizeService) *CFOptimizeHandler {
	return &CFOptimizeHandler{cfOptimizeService: cfOptimizeService}
}

// Create handles one-click CDN optimization
func (h *CFOptimizeHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req models.CreateCFOptimizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := h.cfOptimizeService.Create(c.Request.Context(), userID, req.AccountID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// List returns all CF optimize configs for the current user
func (h *CFOptimizeHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)

	configs, err := h.cfOptimizeService.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if configs == nil {
		configs = []models.CFOptimize{}
	}

	c.JSON(http.StatusOK, configs)
}

// Refresh updates the status of a custom hostname from Cloudflare
func (h *CFOptimizeHandler) Refresh(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	config, err := h.cfOptimizeService.Refresh(c.Request.Context(), userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// Delete removes a CF optimize config
func (h *CFOptimizeHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Default: cleanup records
	cleanup := c.Query("cleanup") != "false"

	if err := h.cfOptimizeService.Delete(c.Request.Context(), userID, id, cleanup); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
