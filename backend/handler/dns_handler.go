package handler

import (
	"fmt"
	"net/http"
	"time"

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

	// Get the latest cache sync time from domains
	domainCacheService := service.NewDomainCacheService()
	caches, err := domainCacheService.GetCacheByUser(userID)
	var latestSyncTime *time.Time
	if err == nil && len(caches) > 0 {
		for _, cache := range caches {
			if cache.LastSyncAt != nil {
				if latestSyncTime == nil || cache.LastSyncAt.After(*latestSyncTime) {
					latestSyncTime = cache.LastSyncAt
				}
			}
		}
	}

	response := gin.H{
		"domains": domains,
	}
	// Only include cache_timestamp if we have a valid sync time
	if latestSyncTime != nil {
		response["cache_timestamp"] = latestSyncTime.Format(time.RFC3339)
	}

	c.JSON(http.StatusOK, response)
}

func (h *DNSHandler) RefreshAllDomains(c *gin.Context) {
	userID := middleware.GetUserID(c)

	// Get soft deleted domains before refresh
	domainCacheService := service.NewDomainCacheService()
	softDeletedDomains, _ := domainCacheService.GetSoftDeletedDomains(userID)
	softDeletedMap := make(map[string]bool)
	for _, d := range softDeletedDomains {
		key := fmt.Sprintf("%d:%s", d.AccountID, d.DomainID)
		softDeletedMap[key] = true
	}

	// Fetch domains from providers
	domains, err := h.dnsService.ListAllDomainsFromProvider(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if domains == nil {
		domains = []models.Domain{}
	}

	// Check for domains to delete and domains to restore
	providerDomainMap := make(map[string]bool)
	for _, d := range domains {
		key := fmt.Sprintf("%d:%s", d.AccountID, d.ID)
		providerDomainMap[key] = true
	}

	// Get all cached domains (excluding soft deleted for comparison)
	allCaches, _ := domainCacheService.GetCacheByUser(userID)
	
	// Get account names for display
	accountService := service.NewAccountService()
	accounts, _ := accountService.List(userID)
	accountMap := make(map[int64]string)
	for _, acc := range accounts {
		accountMap[acc.ID] = acc.Name
	}
	
	var domainsToDelete []models.BatchCacheDeleteItem
	for _, cache := range allCaches {
		key := fmt.Sprintf("%d:%s", cache.AccountID, cache.DomainID)
		if !providerDomainMap[key] {
			domainsToDelete = append(domainsToDelete, models.BatchCacheDeleteItem{
				AccountID:   cache.AccountID,
				AccountName: accountMap[cache.AccountID],
				DomainID:    cache.DomainID,
				DomainName:  cache.DomainName,
			})
		}
	}

	// Check for restored domains (soft deleted but now exist in provider)
	var restoredDomains []string
	for _, d := range domains {
		key := fmt.Sprintf("%d:%s", d.AccountID, d.ID)
		if softDeletedMap[key] {
			// Auto restore
			domainCacheService.BatchRestoreCache(userID, []models.BatchCacheDeleteItem{
				{AccountID: d.AccountID, DomainID: d.ID},
			})
			restoredDomains = append(restoredDomains, d.Name)
		}
	}

	// Update last sync time for all domains
	syncTime := time.Now()
	for _, d := range domains {
		var updatedOn *time.Time
		if d.UpdatedOn != "" {
			if t, err := time.Parse(time.RFC3339, d.UpdatedOn); err == nil {
				updatedOn = &t
			} else if t, err := time.Parse("2006-01-02T15:04:05Z", d.UpdatedOn); err == nil {
				updatedOn = &t
			} else if t, err := time.Parse("2006-01-02", d.UpdatedOn); err == nil {
				updatedOn = &t
			}
		}
		domainCacheService.UpdateLastSyncTime(userID, d.AccountID, d.ID, updatedOn)
	}

	response := models.RefreshDomainsResponse{
		Domains:         domains,
		DomainsToDelete: domainsToDelete,
		RestoredDomains: restoredDomains,
		CacheTimestamp:  syncTime.Format(time.RFC3339),
		HasChanges:      len(domainsToDelete) > 0 || len(restoredDomains) > 0,
	}

	c.JSON(http.StatusOK, response)
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
	// Get the latest cache sync time from domains
	var latestSyncTime *time.Time
	domainCacheService := service.NewDomainCacheService()
	caches, err := domainCacheService.GetCacheByUser(userID)
	if err == nil && len(caches) > 0 {
		for _, cache := range caches {
			if cache.AccountID == accountID && cache.LastSyncAt != nil {
				if latestSyncTime == nil || cache.LastSyncAt.After(*latestSyncTime) {
					latestSyncTime = cache.LastSyncAt
				}
			}
		}
	}
	
	response := gin.H{
		"domains": domains,
	}
	// Only include cache_timestamp if we have a valid sync time
	if latestSyncTime != nil {
		response["cache_timestamp"] = latestSyncTime.Format(time.RFC3339)
	}
	
	c.JSON(http.StatusOK, response)
}

func (h *DNSHandler) RefreshDomains(c *gin.Context) {
	userID := middleware.GetUserID(c)
	accountID, err := middleware.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	// Get soft deleted domains before refresh
	domainCacheService := service.NewDomainCacheService()
	softDeletedDomains, _ := domainCacheService.GetSoftDeletedDomains(userID)
	softDeletedMap := make(map[string]bool)
	for _, d := range softDeletedDomains {
		if d.AccountID == accountID {
			key := fmt.Sprintf("%d:%s", d.AccountID, d.DomainID)
			softDeletedMap[key] = true
		}
	}

	domains, _, err := h.dnsService.ListDomainsFromProvider(c.Request.Context(), userID, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if domains == nil {
		domains = []models.Domain{}
	}

	// Check for domains to delete and domains to restore
	providerDomainMap := make(map[string]bool)
	for _, d := range domains {
		key := fmt.Sprintf("%d:%s", d.AccountID, d.ID)
		providerDomainMap[key] = true
	}

	// Get cached domains for this account to compare
	allCaches, _ := domainCacheService.GetCacheByUser(userID)

	var domainsToDelete []models.BatchCacheDeleteItem
	for _, cache := range allCaches {
		if cache.AccountID != accountID {
			continue
		}
		key := fmt.Sprintf("%d:%s", cache.AccountID, cache.DomainID)
		if !providerDomainMap[key] {
			domainsToDelete = append(domainsToDelete, models.BatchCacheDeleteItem{
				AccountID:   cache.AccountID,
				DomainID:    cache.DomainID,
				DomainName:  cache.DomainName,
			})
		}
	}

	// Check for restored domains (soft deleted but now exist in provider)
	var restoredDomains []string
	for _, d := range domains {
		key := fmt.Sprintf("%d:%s", d.AccountID, d.ID)
		if softDeletedMap[key] {
			domainCacheService.BatchRestoreCache(userID, []models.BatchCacheDeleteItem{
				{AccountID: d.AccountID, DomainID: d.ID},
			})
			restoredDomains = append(restoredDomains, d.Name)
		}
	}

	// Update last sync time for all domains
	syncTime := time.Now()
	for _, d := range domains {
		var updatedOn *time.Time
		if d.UpdatedOn != "" {
			if t, err := time.Parse(time.RFC3339, d.UpdatedOn); err == nil {
				updatedOn = &t
			} else if t, err := time.Parse("2006-01-02T15:04:05Z", d.UpdatedOn); err == nil {
				updatedOn = &t
			} else if t, err := time.Parse("2006-01-02", d.UpdatedOn); err == nil {
				updatedOn = &t
			}
		}
		domainCacheService.UpdateLastSyncTime(userID, d.AccountID, d.ID, updatedOn)
	}

	response := models.RefreshDomainsResponse{
		Domains:         domains,
		DomainsToDelete: domainsToDelete,
		RestoredDomains: restoredDomains,
		CacheTimestamp:  syncTime.Format(time.RFC3339),
		HasChanges:      len(domainsToDelete) > 0 || len(restoredDomains) > 0,
	}

	c.JSON(http.StatusOK, response)
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

	// Get domain name for logging - prefer from record, fallback to GetDomain
	domainName := record.DomainName
	if domainName == "" {
		domain, _ := h.dnsService.GetDomain(c.Request.Context(), userID, accountID, domainID)
		if domain != nil {
			domainName = domain.Name
			if domain.UnicodeName != "" {
				domainName = domain.UnicodeName
			}
		} else {
			domainName = domainID
		}
	}

	// Log operation
	h.logService.CreateLog(userID, "create", "record", record.ID, map[string]interface{}{
		"domain":    domainName,
		"type":      record.RecordType,
		"node_name": record.NodeName,
		"content":   record.Content,
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

	// Get domain name for logging - prefer from record, fallback to GetDomain
	domainName := record.DomainName
	if domainName == "" {
		domain, _ := h.dnsService.GetDomain(c.Request.Context(), userID, accountID, domainID)
		if domain != nil {
			domainName = domain.Name
			if domain.UnicodeName != "" {
				domainName = domain.UnicodeName
			}
		} else {
			domainName = domainID
		}
	}

	// Log operation with changes
	logDetails := map[string]interface{}{
		"domain":    domainName,
		"type":      record.RecordType,
		"node_name": record.NodeName,
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
	var domainName string
	for _, r := range records {
		if r.ID == recordID {
			recordDetails = map[string]interface{}{
				"type":      r.RecordType,
				"node_name": r.NodeName,
				"content":   r.Content,
			}
			// Get domain name from record
			domainName = r.DomainName
			break
		}
	}

	// Fallback to GetDomain if domain name not found in record
	if domainName == "" {
		domain, _ := h.dnsService.GetDomain(c.Request.Context(), userID, accountID, domainID)
		if domain != nil {
			domainName = domain.Name
			if domain.UnicodeName != "" {
				domainName = domain.UnicodeName
			}
		} else {
			domainName = domainID
		}
	}

	if err := h.dnsService.DeleteRecord(c.Request.Context(), userID, accountID, domainID, recordID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log operation
	if recordDetails == nil {
		recordDetails = map[string]interface{}{}
	}
	recordDetails["domain"] = domainName
	h.logService.CreateLog(userID, "delete", "record", recordID, recordDetails, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "record deleted"})
}
