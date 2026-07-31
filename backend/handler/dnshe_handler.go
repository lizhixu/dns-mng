package handler

import (
	"net/http"

	"dns-mng/middleware"
	"dns-mng/models"
	"dns-mng/service"

	"github.com/gin-gonic/gin"
)

type DNSHEHandler struct {
	dnsheService          *service.DNSHEService
	dnsheAutoRenewService *service.DNSHEAutoRenewService
	logService            *service.LogService
	schedulerLogService   *service.SchedulerLogService
}

func NewDNSHEHandler(dnsheService *service.DNSHEService, dnsheAutoRenewService *service.DNSHEAutoRenewService, logService *service.LogService, schedulerLogService *service.SchedulerLogService) *DNSHEHandler {
	return &DNSHEHandler{
		dnsheService:          dnsheService,
		dnsheAutoRenewService: dnsheAutoRenewService,
		logService:            logService,
		schedulerLogService:   schedulerLogService,
	}
}

// ListAccounts returns all DNSHE accounts for the current user.
func (h *DNSHEHandler) ListAccounts(c *gin.Context) {
	userID := middleware.GetUserID(c)
	accounts, err := h.dnsheService.ListAccounts(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if accounts == nil {
		accounts = []models.Account{}
	}
	c.JSON(http.StatusOK, gin.H{"accounts": accounts})
}

// GetQuota returns the quota for a DNSHE account.
func (h *DNSHEHandler) GetQuota(c *gin.Context) {
	userID := middleware.GetUserID(c)
	accountID, err := middleware.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}
	quota, err := h.dnsheService.GetQuota(c.Request.Context(), userID, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, quota)
}

type dnsheSubdomainRequest struct {
	Subdomain string `json:"subdomain" binding:"required"`
	Rootdomain string `json:"rootdomain" binding:"required"`
}

// RegisterSubdomain registers a new subdomain.
func (h *DNSHEHandler) RegisterSubdomain(c *gin.Context) {
	userID := middleware.GetUserID(c)
	accountID, err := middleware.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}
	var req dnsheSubdomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.dnsheService.RegisterSubdomain(c.Request.Context(), userID, accountID, req.Subdomain, req.Rootdomain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

type dnsheSubdomainIDRequest struct {
	SubdomainID int `json:"subdomain_id" binding:"required"`
}

// DeleteSubdomain deletes a subdomain.
func (h *DNSHEHandler) DeleteSubdomain(c *gin.Context) {
	userID := middleware.GetUserID(c)
	accountID, err := middleware.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}
	var req dnsheSubdomainIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.dnsheService.DeleteSubdomain(c.Request.Context(), userID, accountID, req.SubdomainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

type dnsheSetResolutionRequest struct {
	UsesDNSHEDNS bool `json:"uses_dnshe_dns"`
}

// SetDomainResolution updates the uses_dnshe_dns flag for a domain.
func (h *DNSHEHandler) SetDomainResolution(c *gin.Context) {
	userID := middleware.GetUserID(c)
	accountID, err := middleware.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}
	domainID := c.Param("domainId")
	if domainID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}
	var req dnsheSetResolutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.dnsheService.SetDomainResolution(userID, accountID, domainID, req.UsesDNSHEDNS); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "resolution updated"})
}

// ResolveToCloudflare switches a DNSHE domain's NS records to Cloudflare.
func (h *DNSHEHandler) ResolveToCloudflare(c *gin.Context) {
	userID := middleware.GetUserID(c)
	accountID, err := middleware.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}
	domainID := c.Param("domainId")
	if domainID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}
	var req models.ResolveToCloudflareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.dnsheService.ResolveToCloudflare(c.Request.Context(), userID, accountID, domainID, req.CloudflareAccountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetAutoRenewConfig returns the auto-renew config for the current user.
func (h *DNSHEHandler) GetAutoRenewConfig(c *gin.Context) {
	userID := middleware.GetUserID(c)
	cfg, err := h.dnsheAutoRenewService.GetConfig(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// UpdateAutoRenewConfig updates the auto-renew config for the current user.
func (h *DNSHEHandler) UpdateAutoRenewConfig(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.UpdateDNSHEAutoRenewConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cfg, err := h.dnsheAutoRenewService.UpdateConfig(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// TriggerAutoRenew manually triggers the auto-renew job for the current user.
func (h *DNSHEHandler) TriggerAutoRenew(c *gin.Context) {
	userID := middleware.GetUserID(c)
	result, err := h.dnsheAutoRenewService.TriggerRunForUser(c.Request.Context(), userID, h.schedulerLogService)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.dnsheAutoRenewService.UpdateLastRunAt(userID)
	c.JSON(http.StatusOK, result)
}