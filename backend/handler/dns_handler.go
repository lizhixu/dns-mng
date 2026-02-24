package handler

import (
	"net/http"

	"dns-mng/middleware"
	"dns-mng/models"
	"dns-mng/service"

	"github.com/gin-gonic/gin"
)

type DNSHandler struct {
	dnsService *service.DNSService
}

func NewDNSHandler(dnsService *service.DNSService) *DNSHandler {
	return &DNSHandler{dnsService: dnsService}
}

func (h *DNSHandler) ListAllDomains(c *gin.Context) {
	userID := middleware.GetUserID(c)

	domains, err := h.dnsService.ListAllDomains(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if domains == nil {
		domains = []models.Domain{}
	}
	c.JSON(http.StatusOK, domains)
}

func (h *DNSHandler) ListDomains(c *gin.Context) {
	userID := middleware.GetUserID(c)
	accountID, err := middleware.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	domains, err := h.dnsService.ListDomains(c.Request.Context(), userID, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if domains == nil {
		domains = []models.Domain{}
	}
	c.JSON(http.StatusOK, domains)
}

func (h *DNSHandler) GetDomain(c *gin.Context) {
	userID := middleware.GetUserID(c)
	accountID, err := middleware.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}
	domainID := c.Param("domainId")

	domain, err := h.dnsService.GetDomain(c.Request.Context(), userID, accountID, domainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, domain)
}

func (h *DNSHandler) ListRecords(c *gin.Context) {
	userID := middleware.GetUserID(c)
	accountID, err := middleware.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}
	domainID := c.Param("domainId")

	records, err := h.dnsService.ListRecords(c.Request.Context(), userID, accountID, domainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if records == nil {
		records = []models.Record{}
	}
	c.JSON(http.StatusOK, records)
}

func (h *DNSHandler) CreateRecord(c *gin.Context) {
	userID := middleware.GetUserID(c)
	accountID, err := middleware.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}
	domainID := c.Param("domainId")

	var req models.CreateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record, err := h.dnsService.CreateRecord(c.Request.Context(), userID, accountID, domainID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, record)
}

func (h *DNSHandler) UpdateRecord(c *gin.Context) {
	userID := middleware.GetUserID(c)
	accountID, err := middleware.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}
	domainID := c.Param("domainId")
	recordID := c.Param("recordId")

	var req models.UpdateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record, err := h.dnsService.UpdateRecord(c.Request.Context(), userID, accountID, domainID, recordID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, record)
}

func (h *DNSHandler) DeleteRecord(c *gin.Context) {
	userID := middleware.GetUserID(c)
	accountID, err := middleware.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}
	domainID := c.Param("domainId")
	recordID := c.Param("recordId")

	if err := h.dnsService.DeleteRecord(c.Request.Context(), userID, accountID, domainID, recordID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "record deleted"})
}
