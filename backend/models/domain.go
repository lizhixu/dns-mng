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
	RenewalDate   string `json:"renewal_date,omitempty"`
	RenewalURL    string `json:"renewal_url,omitempty"`
	CacheSynced   bool   `json:"cache_synced,omitempty"`
}

// DomainCache represents cached domain data with renewal info
type DomainCache struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	AccountID   int64     `json:"account_id"`
	DomainID    string    `json:"domain_id"`
	DomainName  string    `json:"domain_name"`
	RenewalDate string    `json:"renewal_date,omitempty"`
	RenewalURL  string    `json:"renewal_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UpdateDomainCacheRequest is the request body for updating domain cache
type UpdateDomainCacheRequest struct {
	RenewalDate       string `json:"renewal_date"`
	RenewalURL        string `json:"renewal_url"`
	NotifyDaysBefore  int    `json:"notify_days_before"`
	NotifyEnabled     bool   `json:"notify_enabled"`
}

// BatchCacheItem represents a single item in batch cache operations
type BatchCacheItem struct {
	AccountID   int64  `json:"account_id"`
	DomainID    string `json:"domain_id"`
	DomainName  string `json:"domain_name"`
	RenewalDate string `json:"renewal_date"`
	RenewalURL  string `json:"renewal_url"`
}

// BatchCacheRequest is the request body for batch updating domain cache
type BatchCacheRequest struct {
	Items []BatchCacheItem `json:"items"`
}

// BatchCacheDeleteItem represents a single item in batch delete operations
type BatchCacheDeleteItem struct {
	AccountID int64  `json:"account_id"`
	DomainID  string `json:"domain_id"`
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
