package dnshe

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"dns-mng/models"
)

type Provider struct {
	client *Client
}

func NewProvider() *Provider {
	return &Provider{
		client: NewClient(),
	}
}

// New is an alias for NewProvider
func New() *Provider {
	return NewProvider()
}

func (p *Provider) Name() string {
	return "dnshe"
}

func (p *Provider) DisplayName() string {
	return "DNSHE"
}

func (p *Provider) WebsiteURL() string {
	return "https://www.dnshe.com"
}

func (p *Provider) DefaultTTL() int {
	return 600
}

// parseAPIKey splits the API key format "apikey,apisecret"
func parseAPIKey(apiKey string) (string, string, error) {
	parts := strings.Split(apiKey, ",")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid API key format, expected: apikey,apisecret")
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func (p *Provider) ListDomains(ctx context.Context, apiKey string) ([]models.Domain, error) {
	key, secret, err := parseAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.ListSubdomains(ctx, key, secret)
	if err != nil {
		return nil, err
	}

	domains := make([]models.Domain, 0, len(resp.Subdomains))
	for _, sub := range resp.Subdomains {
		domains = append(domains, models.Domain{
			ID:          strconv.Itoa(sub.ID),
			Name:        sub.FullDomain,
			UnicodeName: sub.FullDomain,
			State:       sub.Status,
			CreatedOn:   sub.CreatedAt,
			UpdatedOn:   sub.UpdatedAt,
		})
	}

	return domains, nil
}

func (p *Provider) GetDomain(ctx context.Context, apiKey string, domainID string) (*models.Domain, error) {
	// For DNSHE, we can get domain info from the list
	domains, err := p.ListDomains(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	for _, domain := range domains {
		if domain.ID == domainID {
			return &domain, nil
		}
	}

	return nil, fmt.Errorf("domain not found: %s", domainID)
}

func (p *Provider) ListRecords(ctx context.Context, apiKey string, domainID string) ([]models.Record, error) {
	key, secret, err := parseAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	id, err := strconv.Atoi(domainID)
	if err != nil {
		return nil, fmt.Errorf("invalid domain ID: %w", err)
	}

	// Get domain info
	domain, err := p.GetDomain(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}

	// Get DNS records
	resp, err := p.client.ListDNSRecords(ctx, key, secret, id)
	if err != nil {
		return nil, err
	}

	records := make([]models.Record, 0, len(resp.Records))
	for _, r := range resp.Records {
		// Extract node name from full record name
		nodeName := strings.TrimSuffix(r.Name, "."+domain.Name)
		nodeName = strings.TrimSuffix(nodeName, ".")
		if nodeName == domain.Name || nodeName == strings.TrimSuffix(domain.Name, ".") {
			nodeName = ""
		}

		priority := 0
		if r.Priority != nil {
			priority = *r.Priority
		}

		records = append(records, models.Record{
			ID:         strconv.Itoa(r.ID),
			DomainID:   domainID,
			DomainName: domain.Name,
			NodeName:   nodeName,
			RecordType: r.Type,
			TTL:        r.TTL,
			State:      r.Status == "active",
			Content:    r.Content,
			Priority:   priority,
			UpdatedOn:  r.CreatedAt,
		})
	}

	return records, nil
}

func (p *Provider) CreateRecord(ctx context.Context, apiKey string, domainID string, record *models.Record) (*models.Record, error) {
	key, secret, err := parseAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	id, err := strconv.Atoi(domainID)
	if err != nil {
		return nil, fmt.Errorf("invalid domain ID: %w", err)
	}

	ttl := record.TTL
	if ttl == 0 {
		ttl = 600 // Default TTL
	}

	var priority *int
	if record.RecordType == "MX" || record.RecordType == "SRV" {
		p := record.Priority
		if p == 0 {
			p = 10 // Default priority
		}
		priority = &p
	}

	err = p.client.CreateDNSRecord(ctx, key, secret, id, record.NodeName, record.RecordType, record.Content, ttl, priority)
	if err != nil {
		return nil, err
	}

	// Get domain info for domain name
	domain, err := p.GetDomain(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}

	// Return the created record
	return &models.Record{
		ID:         "new",
		DomainID:   domainID,
		DomainName: domain.Name,
		NodeName:   record.NodeName,
		RecordType: record.RecordType,
		TTL:        ttl,
		State:      true,
		Content:    record.Content,
		Priority:   record.Priority,
		UpdatedOn:  time.Now().Format(time.RFC3339),
	}, nil
}

func (p *Provider) UpdateRecord(ctx context.Context, apiKey string, domainID string, record *models.Record) (*models.Record, error) {
	key, secret, err := parseAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	// Get the old record first
	oldRecord, err := p.GetRecord(ctx, apiKey, domainID, record.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get old record: %w", err)
	}

	// Check if record has actually changed
	if oldRecord.NodeName == record.NodeName &&
		oldRecord.RecordType == record.RecordType &&
		oldRecord.Content == record.Content &&
		oldRecord.TTL == record.TTL &&
		oldRecord.Priority == record.Priority {
		// No changes, return the original record
		return record, nil
	}

	// If node name or type changed, we need to delete and recreate
	if oldRecord.NodeName != record.NodeName || oldRecord.RecordType != record.RecordType {
		// Delete the old record
		err = p.DeleteRecord(ctx, apiKey, domainID, record.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete old record: %w", err)
		}

		// Create a new record with the updated values
		newRecord := &models.Record{
			DomainID:   domainID,
			NodeName:   record.NodeName,
			RecordType: record.RecordType,
			TTL:        record.TTL,
			Content:    record.Content,
			Priority:   record.Priority,
		}

		createdRecord, err := p.CreateRecord(ctx, apiKey, domainID, newRecord)
		if err != nil {
			// Try to restore old record if create failed
			restoreRecord := &models.Record{
				DomainID:   domainID,
				NodeName:   oldRecord.NodeName,
				RecordType: oldRecord.RecordType,
				TTL:        oldRecord.TTL,
				Content:    oldRecord.Content,
				Priority:   oldRecord.Priority,
			}
			_, restoreErr := p.CreateRecord(ctx, apiKey, domainID, restoreRecord)
			if restoreErr != nil {
				return nil, fmt.Errorf("failed to create new record: %w, and failed to restore old record: %v", err, restoreErr)
			}
			return nil, fmt.Errorf("failed to create new record: %w (old record has been restored)", err)
		}

		return createdRecord, nil
	}

	// Only content/TTL/priority changed, use update API
	recordID, err := strconv.Atoi(record.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid record ID: %w", err)
	}

	ttl := record.TTL
	if ttl == 0 {
		ttl = 600
	}

	var priority *int
	if record.RecordType == "MX" || record.RecordType == "SRV" {
		p := record.Priority
		if p == 0 {
			p = 10
		}
		priority = &p
	}

	err = p.client.UpdateDNSRecord(ctx, key, secret, recordID, record.Content, ttl, priority)
	if err != nil {
		return nil, err
	}

	// Set update time
	record.UpdatedOn = time.Now().Format(time.RFC3339)

	return record, nil
}

func (p *Provider) DeleteRecord(ctx context.Context, apiKey string, domainID string, recordID string) error {
	key, secret, err := parseAPIKey(apiKey)
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(recordID)
	if err != nil {
		return fmt.Errorf("invalid record ID: %w", err)
	}

	return p.client.DeleteDNSRecord(ctx, key, secret, id)
}

func (p *Provider) GetRecord(ctx context.Context, apiKey string, domainID string, recordID string) (*models.Record, error) {
	records, err := p.ListRecords(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		if record.ID == recordID {
			return &record, nil
		}
	}

	return nil, fmt.Errorf("record not found: %s", recordID)
}
