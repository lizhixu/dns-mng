package provider

import (
	"context"
	"dns-mng/models"
)

// DNSProvider defines the interface any DNS platform must implement.
// To add a new platform (e.g. Cloudflare), implement this interface
// and register it in the registry.
type DNSProvider interface {
	// Name returns the provider identifier (e.g. "dynu", "cloudflare")
	Name() string

	// DisplayName returns human-readable provider name
	DisplayName() string

	// WebsiteURL returns the provider's website URL
	WebsiteURL() string

	// ListDomains returns all domains for the given API key
	ListDomains(ctx context.Context, apiKey string) ([]models.Domain, error)

	// GetDomain returns details of a specific domain
	GetDomain(ctx context.Context, apiKey string, domainID string) (*models.Domain, error)

	// ListRecords returns all DNS records for a domain
	ListRecords(ctx context.Context, apiKey string, domainID string) ([]models.Record, error)

	// CreateRecord creates a new DNS record
	CreateRecord(ctx context.Context, apiKey string, domainID string, record *models.Record) (*models.Record, error)

	// UpdateRecord updates an existing DNS record
	UpdateRecord(ctx context.Context, apiKey string, domainID string, record *models.Record) (*models.Record, error)

	// DeleteRecord deletes a DNS record
	DeleteRecord(ctx context.Context, apiKey string, domainID string, recordID string) error
}

// ProviderInfo describes a registered provider for the UI
type ProviderInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	WebsiteURL  string `json:"website_url"`
}
