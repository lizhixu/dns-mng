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
	logService *service.LogService
}

func NewDNSHandler(dnsService *service.DNSService, logService *service.LogService) *DNSHandler {
	return &DNSHandler{
		dnsService: dnsService,
		logService: logService,
	}
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

	// Log operation
	h.logService.CreateLog(userID, "create", "record", record.ID, map[string]interface{}{
		"domain":      domainID,
		"node_name":   record.NodeName,
		"record_type": record.RecordType,
		"content":     record.Content,
		"ttl":         record.TTL,
		"account":     accountID,
	}, c.ClientIP())

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

	// Get old record for comparison
	oldRecords, _ := h.dnsService.ListRecords(c.Request.Context(), userID, accountID, domainID)
	var oldRecord *models.Record
	for _, r := range oldRecords {
		if r.ID == recordID {
			oldRecord = &r
			break
		}
	}

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

	// Log operation with before/after comparison
	logDetails := map[string]interface{}{
		"domain":      domainID,
		"node_name":   record.NodeName,
		"record_type": record.RecordType,
		"account":     accountID,
	}
	if oldRecord != nil {
		changes := make(map[string]interface{})
		if oldRecord.Content != record.Content {
			changes["content"] = map[string]string{"old": oldRecord.Content, "new": record.Content}
		}
		if oldRecord.TTL != record.TTL {
			changes["ttl"] = map[string]int{"old": oldRecord.TTL, "new": record.TTL}
		}
		if oldRecord.State != record.State {
			changes["state"] = map[string]bool{"old": oldRecord.State, "new": record.State}
		}
		if len(changes) > 0 {
			logDetails["changes"] = changes
		}
	}
	h.logService.CreateLog(userID, "update", "record", recordID, logDetails, c.ClientIP())

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

	// Get record details before deletion
	records, _ := h.dnsService.ListRecords(c.Request.Context(), userID, accountID, domainID)
	var recordDetails map[string]interface{}
	for _, r := range records {
		if r.ID == recordID {
			recordDetails = map[string]interface{}{
				"domain":      domainID,
				"node_name":   r.NodeName,
				"record_type": r.RecordType,
				"content":     r.Content,
				"ttl":         r.TTL,
				"account":     accountID,
			}
			break
		}
	}

	if err := h.dnsService.DeleteRecord(c.Request.Context(), userID, accountID, domainID, recordID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log operation
	if recordDetails == nil {
		recordDetails = map[string]interface{}{
			"domain":  domainID,
			"account": accountID,
		}
	}
	h.logService.CreateLog(userID, "delete", "record", recordID, recordDetails, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "record deleted"})
}
