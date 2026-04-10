package handler

import (
	"net/http"

	"dns-mng/middleware"
	"dns-mng/models"
	"dns-mng/service"

	"github.com/gin-gonic/gin"
)

type DomainCacheHandler struct {
	dnsService *service.DNSService
	logService *service.LogService
}

func NewDomainCacheHandler(dnsService *service.DNSService, logService *service.LogService) *DomainCacheHandler {
	return &DomainCacheHandler{
		dnsService: dnsService,
		logService: logService,
	}
}

func (h *DomainCacheHandler) UpdateDomainCache(c *gin.Context) {
	userID := middleware.GetUserID(c)
	accountID, err := middleware.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}
	domainID := c.Param("domainId")

	var req models.UpdateDomainCacheRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get domain name from provider
	domain, err := h.dnsService.GetDomain(c.Request.Context(), userID, accountID, domainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	domainName := domain.Name
	if domain.UnicodeName != "" {
		domainName = domain.UnicodeName
	}

	updatedDomain, err := h.dnsService.UpdateDomainCache(c.Request.Context(), userID, accountID, domainID, domainName, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log operation
	h.logService.CreateLog(userID, "update", "domain_cache", domainID, map[string]interface{}{
		"domain":       domainName,
		"renewal_date":  req.RenewalDate,
		"renewal_url":   req.RenewalURL,
		"account":       accountID,
	}, c.ClientIP())

	c.JSON(http.StatusOK, updatedDomain)
}

func (h *DomainCacheHandler) BatchUpdateDomainCache(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req models.BatchCacheRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no items provided"})
		return
	}

	err := h.dnsService.BatchUpdateDomainCache(c.Request.Context(), userID, req.Items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log operation
	h.logService.CreateLog(userID, "batch_update", "domain_cache", "", map[string]interface{}{
		"count": len(req.Items),
	}, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "batch update successful", "count": len(req.Items)})
}

func (h *DomainCacheHandler) BatchDeleteDomainCache(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req models.BatchCacheDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no items provided"})
		return
	}

	err := h.dnsService.BatchDeleteDomainCache(c.Request.Context(), userID, req.Items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log operation
	h.logService.CreateLog(userID, "batch_delete", "domain_cache", "", map[string]interface{}{
		"count": len(req.Items),
	}, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "batch delete successful", "count": len(req.Items)})
}

func (h *DomainCacheHandler) GetCacheStats(c *gin.Context) {
	userID := middleware.GetUserID(c)

	stats, err := h.dnsService.GetCacheStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
