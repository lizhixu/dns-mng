package handler

import (
	"net/http"
	"strings"

	"dns-mng/models"
	"dns-mng/service"

	"github.com/gin-gonic/gin"
)

type DDNSHandler struct {
	dnsService       *service.DNSService
	accountService   *service.AccountService
	logService       *service.LogService
	ddnsTokenService *service.DDNSTokenService
}

func NewDDNSHandler(dnsService *service.DNSService, accountService *service.AccountService, logService *service.LogService, ddnsTokenService *service.DDNSTokenService) *DDNSHandler {
	return &DDNSHandler{
		dnsService:       dnsService,
		accountService:   accountService,
		logService:       logService,
		ddnsTokenService: ddnsTokenService,
	}
}

// UpdateDDNS updates a DNS record with the client's IP address
// Compatible with DuckDNS API format
// Query parameters:
// - token: authentication token (user-level)
// - domains: required, comma-separated domain names to update
// - ip: optional IPv4 address (if not provided, uses client IP)
// - ipv6: optional IPv6 address
func (h *DDNSHandler) UpdateDDNS(c *gin.Context) {
	tokenValue := c.Query("token")
	if tokenValue == "" {
		c.String(http.StatusBadRequest, "KO")
		return
	}

	// Get token from database (user-level)
	token, err := h.ddnsTokenService.GetTokenByValue(tokenValue)
	if err != nil {
		c.String(http.StatusInternalServerError, "KO")
		return
	}

	if token == nil {
		c.String(http.StatusUnauthorized, "KO")
		return
	}

	if !token.Enabled {
		c.String(http.StatusForbidden, "KO")
		return
	}

	// Parse domains parameter
	domainsParam := c.Query("domains")
	if domainsParam == "" {
		c.String(http.StatusBadRequest, "KO")
		return
	}

	domains := strings.Split(domainsParam, ",")
	if len(domains) == 0 || domains[0] == "" {
		c.String(http.StatusBadRequest, "KO")
		return
	}

	// Get IP address
	ip := c.Query("ip")
	ipv6 := c.Query("ipv6")

	// If no IP provided, use client IP
	if ip == "" && ipv6 == "" {
		clientIP := c.ClientIP()
		if strings.Contains(clientIP, ":") {
			ipv6 = clientIP
		} else {
			ip = clientIP
		}
	}

	if ip == "" && ipv6 == "" {
		c.String(http.StatusBadRequest, "KO")
		return
	}

	userID := token.UserID

	// Track updated domains and records for logging
	updatedDomains := []string{}
	updatedRecords := []map[string]interface{}{}
	skippedDomains := []string{}

	// Update each domain
	for _, domainName := range domains {
		domainName = strings.TrimSpace(domainName)
		if domainName == "" {
			continue
		}

		domainUpdated := false

		// Find the domain and update its records
		// Get all accounts for this user
		accounts, err := h.accountService.List(userID)
		if err != nil {
			continue
		}

		for _, account := range accounts {
			domainList, err := h.dnsService.ListDomains(c.Request.Context(), userID, account.ID)
			if err != nil {
				continue
			}

			for _, domain := range domainList {
				if domain.Name == domainName || domain.Name == domainName+"." {
					// Found the domain, update its records
					records, err := h.dnsService.ListRecords(c.Request.Context(), userID, account.ID, domain.ID)
					if err != nil {
						continue
					}

					for _, record := range records {
						if record.RecordType == "A" && ip != "" && record.Content != ip {
							_, err = h.dnsService.UpdateRecord(
								c.Request.Context(),
								userID,
								account.ID,
								domain.ID,
								record.ID,
								&models.UpdateRecordRequest{
									NodeName:   record.NodeName,
									RecordType: record.RecordType,
									TTL:        record.TTL,
									State:      &record.State,
									Content:    ip,
									Priority:   record.Priority,
								},
							)
							if err == nil {
								domainUpdated = true
								updatedRecords = append(updatedRecords, map[string]interface{}{
									"domain":    domain.Name,
									"type":      "A",
									"node_name": record.NodeName,
									"old_ip":    record.Content,
									"new_ip":    ip,
								})
							}
						} else if record.RecordType == "AAAA" && ipv6 != "" && record.Content != ipv6 {
							_, err = h.dnsService.UpdateRecord(
								c.Request.Context(),
								userID,
								account.ID,
								domain.ID,
								record.ID,
								&models.UpdateRecordRequest{
									NodeName:   record.NodeName,
									RecordType: record.RecordType,
									TTL:        record.TTL,
									State:      &record.State,
									Content:    ipv6,
									Priority:   record.Priority,
							},
							)
							if err == nil {
								domainUpdated = true
								updatedRecords = append(updatedRecords, map[string]interface{}{
									"domain":    domain.Name,
									"type":      "AAAA",
									"node_name": record.NodeName,
									"old_ip":    record.Content,
									"new_ip":    ipv6,
								})
							}
						}
					}
				}
			}
		}

		if domainUpdated {
			updatedDomains = append(updatedDomains, domainName)
		} else {
			skippedDomains = append(skippedDomains, domainName)
		}
	}

	// Update last used timestamp
	h.ddnsTokenService.UpdateLastUsed(tokenValue, ip)

	c.String(http.StatusOK, "OK")
}
