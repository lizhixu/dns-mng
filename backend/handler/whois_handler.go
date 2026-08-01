package handler

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"dns-mng/middleware"
	"dns-mng/models"
	"dns-mng/service"

	"github.com/gin-gonic/gin"
)

// WHOISHandler exposes WHOIS lookup configuration and the lookup endpoint.
// The API key is stored per-user in the database and returned in plaintext to
// the owning user (they configure and view it themselves, same pattern as the
// DDNS token); lookups are proxied through the backend so the key is used to
// call WhoisJSON.com server-side and is never embedded in browser-facing logic.
type WHOISHandler struct {
	whoisService *service.WHOISService
	logService   *service.LogService
}

// NewWHOISHandler constructs a WHOISHandler.
func NewWHOISHandler(whoisService *service.WHOISService, logService *service.LogService) *WHOISHandler {
	return &WHOISHandler{
		whoisService: whoisService,
		logService:   logService,
	}
}

// GetConfig returns the user's WHOIS lookup configuration (with the API key
// cleared) or `{"configured": false}` when no row exists.
func (h *WHOISHandler) GetConfig(c *gin.Context) {
	userID := middleware.GetUserID(c)

	config, err := h.whoisService.GetConfig(userID)
	if err != nil {
		log.Printf("Failed to get WHOIS config for user_id=%d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get WHOIS configuration"})
		return
	}

	if config == nil {
		c.JSON(http.StatusOK, gin.H{"configured": false})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateConfig creates or updates the user's WHOIS lookup configuration.
// An empty api_key in the request means "keep the existing key".
func (h *WHOISHandler) UpdateConfig(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req models.UpdateWHOISConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := h.whoisService.UpsertConfig(userID, &req)
	if err != nil {
		if errors.Is(err, service.ErrWHOISAPIKeyRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "API key is required for initial configuration"})
			return
		}
		log.Printf("Failed to update WHOIS config for user_id=%d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update WHOIS configuration"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// Query performs a WHOIS lookup for the domain given by the "domain" query
// parameter and returns the structured result.
func (h *WHOISHandler) Query(c *gin.Context) {
	userID := middleware.GetUserID(c)

	domain := strings.TrimSpace(c.Query("domain"))
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain is required"})
		return
	}
	if len(domain) > 255 || !strings.Contains(domain, ".") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain"})
		return
	}

	result, err := h.whoisService.Query(userID, domain)
	if err != nil {
		if errors.Is(err, service.ErrWHOISAPIKeyNotConfigured) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "WHOIS API Key is not configured"})
			return
		}
		log.Printf("Failed to query whois for user_id=%d domain=%s: %v", userID, domain, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to query WHOIS information"})
		return
	}

	c.JSON(http.StatusOK, result)
}
