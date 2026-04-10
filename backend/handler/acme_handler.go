package handler

import (
	"net/http"

	"dns-mng/middleware"
	"dns-mng/models"
	"dns-mng/service"

	"github.com/gin-gonic/gin"
)

type AcmeHandler struct {
	acmeService *service.AcmeService
}

func NewAcmeHandler(acmeService *service.AcmeService) *AcmeHandler {
	return &AcmeHandler{acmeService: acmeService}
}

// POST /api/acme/dns01/present
func (h *AcmeHandler) Present(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.AcmeDNS01Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.acmeService.Present(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// POST /api/acme/dns01/cleanup
func (h *AcmeHandler) Cleanup(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.AcmeDNS01Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.acmeService.Cleanup(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

