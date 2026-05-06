package handler

import (
	"net/http"

	"dns-mng/middleware"
	"dns-mng/models"
	"dns-mng/service"

	"github.com/gin-gonic/gin"
)

type DDNSTokenHandler struct {
	ddnsTokenService *service.DDNSTokenService
	logService       *service.LogService
}

func NewDDNSTokenHandler(ddnsTokenService *service.DDNSTokenService, logService *service.LogService) *DDNSTokenHandler {
	return &DDNSTokenHandler{
		ddnsTokenService: ddnsTokenService,
		logService:       logService,
	}
}

// GetToken gets the DDNS token for current user
func (h *DDNSTokenHandler) GetToken(c *gin.Context) {
	userID := middleware.GetUserID(c)

	token, err := h.ddnsTokenService.GetToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get token: " + err.Error()})
		return
	}

	if token == nil {
		c.JSON(http.StatusOK, gin.H{
			"has_token": false,
			"token":     nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"has_token": true,
		"token":     token,
	})
}

// UpdateToken creates or updates the DDNS token for current user
func (h *DDNSTokenHandler) UpdateToken(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req models.UpdateDDNSTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if token exists
	existing, err := h.ddnsTokenService.GetToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get token"})
		return
	}

	var token *models.DDNSToken
	if existing == nil {
		// Create new token
		token, err = h.ddnsTokenService.CreateToken(userID, req.Token)
	} else {
		// Update existing token
		token, err = h.ddnsTokenService.UpdateToken(userID, req.Enabled, req.Token)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update token"})
		return
	}

	c.JSON(http.StatusOK, token)
}

// DeleteToken deletes the DDNS token for current user
func (h *DDNSTokenHandler) DeleteToken(c *gin.Context) {
	userID := middleware.GetUserID(c)

	// Check if token exists
	existing, err := h.ddnsTokenService.GetToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get token"})
		return
	}

	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
		return
	}

	err = h.ddnsTokenService.DeleteToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "token deleted successfully"})
}
