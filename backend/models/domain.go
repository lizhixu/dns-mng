package models

import "time"

// Domain represents a DNS domain from any provider
type Domain struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	UnicodeName string `json:"unicode_name,omitempty"`
	State       string `json:"state,omitempty"`
	Group       string `json:"group,omitempty"`
	IPv4Address string `json:"ipv4_address,omitempty"`
	IPv6Address string `json:"ipv6_address,omitempty"`
	TTL         int    `json:"ttl,omitempty"`
	CreatedOn   string `json:"created_on,omitempty"`
	UpdatedOn   string `json:"updated_on,omitempty"`
	AccountID   int64  `json:"account_id,omitempty"`
	AccountName string `json:"account_name,omitempty"`
	// Cached fields from domain_cache
	RenewalDate string `json:"renewal_date,omitempty"`
	RenewalURL  string `json:"renewal_url,omitempty"`
	CacheSynced bool   `json:"cache_synced,omitempty"`
	// UsesDNSHEDNS indicates whether the domain uses DNSHE's own DNS resolution.
	// Only set for DNSHE-account domains. nil for non-DNSHE domains.
	UsesDNSHEDNS *bool `json:"uses_dnshe_dns,omitempty"`
}

// DomainCache represents cached domain data with renewal info
type DomainCache struct {
	ID                int64      `json:"id"`
	UserID            int64      `json:"user_id"`
	AccountID         int64      `json:"account_id"`
	DomainID          string     `json:"domain_id"`
	DomainName        string     `json:"domain_name"`
	RenewalDate       string     `json:"renewal_date,omitempty"`
	RenewalURL        string     `json:"renewal_url,omitempty"`
	// RenewalManual indicates renewal_date/renewal_url were manually edited by
	// the user and must not be overwritten by provider data during sync.
	RenewalManual     bool       `json:"renewal_manual,omitempty"`
	UsesDNSHEDNS      bool       `json:"uses_dnshe_dns"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty"`
	LastSyncAt        *time.Time `json:"last_sync_at,omitempty"`
	ProviderUpdatedOn *time.Time `json:"provider_updated_on,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// UpdateDomainCacheRequest is the request body for updating domain cache
type UpdateDomainCacheRequest struct {
	RenewalDate      string `json:"renewal_date"`
	RenewalURL       string `json:"renewal_url"`
	NotifyDaysBefore int    `json:"notify_days_before"`
	NotifyEnabled    bool   `json:"notify_enabled"`
	UsesDNSHEDNS     *bool  `json:"uses_dnshe_dns,omitempty"`
	// RenewalManual controls the manual-edit lock flag. nil leaves it unchanged
	// (used by sync paths); 0/1 explicitly clears/sets it (used by the manual
	// edit endpoint).
	RenewalManual    *int   `json:"renewal_manual,omitempty"`
}

// BatchCacheItem represents a single item in batch cache operations
type BatchCacheItem struct {
	AccountID     int64  `json:"account_id"`
	DomainID      string `json:"domain_id"`
	DomainName    string `json:"domain_name"`
	RenewalDate   string `json:"renewal_date"`
	RenewalURL    string `json:"renewal_url"`
	// RenewalManual is the manual-edit lock flag written by BatchUpsertCache.
	// nil leaves the existing value; 0/1 explicitly sets it. BatchUpdateDomainCache
	// computes this automatically, matching UpdateDomainCache behaviour.
	RenewalManual *int `json:"renewal_manual,omitempty"`
}

// BatchCacheRequest is the request body for batch updating domain cache
type BatchCacheRequest struct {
	Items []BatchCacheItem `json:"items"`
}

// BatchCacheDeleteItem represents a single item in batch delete operations
type BatchCacheDeleteItem struct {
	AccountID   int64  `json:"account_id"`
	AccountName string `json:"account_name,omitempty"`
	DomainID    string `json:"domain_id"`
	DomainName  string `json:"domain_name,omitempty"`
}

// BatchCacheDeleteRequest is the request body for batch deleting domain cache
type BatchCacheDeleteRequest struct {
	Items []BatchCacheDeleteItem `json:"items"`
}

// CacheStats represents statistics about cached domains
type CacheStats struct {
	TotalCached     int `json:"total_cached"`
	WithRenewalDate int `json:"with_renewal_date"`
	PermanentFree   int `json:"permanent_free"`
	WithRenewalURL  int `json:"with_renewal_url"`
}

// BatchSoftDeleteRequest is the request body for batch soft deleting domains
type BatchSoftDeleteRequest struct {
	Items []BatchCacheDeleteItem `json:"items"`
}

// BatchRestoreRequest is the request body for batch restoring domains
type BatchRestoreRequest struct {
	Items []BatchCacheDeleteItem `json:"items"`
}

// RefreshDomainsResponse represents the response from refresh domains API
type RefreshDomainsResponse struct {
	Domains         []Domain               `json:"domains"`
	DomainsToDelete []BatchCacheDeleteItem `json:"domains_to_delete"`
	RestoredDomains []string               `json:"restored_domains"`
	CacheTimestamp  string                 `json:"cache_timestamp"`
	HasChanges      bool                   `json:"has_changes"`
}

// DNSHEAutoRenewConfig represents the DNSHE auto-renew configuration for a user
type DNSHEAutoRenewConfig struct {
	UserID      int64      `json:"user_id"`
	Enabled     bool       `json:"enabled"`
	DaysBefore  int        `json:"days_before"`
	LastRunAt   *time.Time `json:"last_run_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// UpdateDNSHEAutoRenewConfigRequest is the request body for updating DNSHE auto-renew config
type UpdateDNSHEAutoRenewConfigRequest struct {
	Enabled    *bool `json:"enabled,omitempty"`
	DaysBefore *int  `json:"days_before,omitempty"`
}

// ResolveToCloudflareRequest is the request body for resolving a DNSHE domain to Cloudflare
type ResolveToCloudflareRequest struct {
	CloudflareAccountID int64 `json:"cloudflare_account_id" binding:"required"`
}

// ResolveToCloudflareResult is the result of resolving a DNSHE domain to Cloudflare
type ResolveToCloudflareResult struct {
	DomainName  string   `json:"domain_name"`
	ZoneName    string   `json:"zone_name"`
	NameServers []string `json:"name_servers"`
}
