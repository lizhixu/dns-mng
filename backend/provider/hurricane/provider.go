package hurricane

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dns-mng/models"
)

// Provider implements the DNSProvider interface for Hurricane Electric (dns.he.net)
type Provider struct {
	client *Client
}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Name() string {
	return "hurricane"
}

func (p *Provider) DisplayName() string {
	return "Hurricane Electric (dns.he.net)"
}

func (p *Provider) WebsiteURL() string {
	return "https://dns.he.net"
}

func (p *Provider) DefaultTTL() int {
	return 300
}

// parseAPIKey splits the API key format "username,password"
func parseAPIKey(apiKey string) (string, string, error) {
	parts := strings.Split(apiKey, ",")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid API key format, expected: username,password")
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func (p *Provider) ListDomains(ctx context.Context, apiKey string) ([]models.Domain, error) {
	username, password, err := parseAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	client := NewClient(username, password)
	zones, err := client.ListZones(ctx)
	if err != nil {
		return nil, err
	}

	domains := make([]models.Domain, 0, len(zones))
	for _, zone := range zones {
		domains = append(domains, models.Domain{
			ID:        zone.ID,
			Name:      zone.Name,
			State:     zone.Status,
			UpdatedOn: time.Now().Format(time.RFC3339),
		})
	}

	return domains, nil
}

func (p *Provider) GetDomain(ctx context.Context, apiKey string, domainID string) (*models.Domain, error) {
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
	username, password, err := parseAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	// Get domain info
	domain, err := p.GetDomain(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}

	client := NewClient(username, password)
	heRecords, err := client.ListRecords(ctx, domainID)
	if err != nil {
		return nil, err
	}

	records := make([]models.Record, 0, len(heRecords))
	for _, r := range heRecords {
		// Extract node name from full record name
		nodeName := r.Name
		if strings.HasSuffix(r.Name, "."+domain.Name) {
			nodeName = strings.TrimSuffix(r.Name, "."+domain.Name)
		} else if r.Name == domain.Name {
			nodeName = ""
		}

		nodeName = strings.TrimSuffix(nodeName, ".")

		// For reverse DNS, if the name is just the IP part, keep it as is
		// HE usually shows the full name, e.g., 1.0.0.127.in-addr.arpa

		records = append(records, models.Record{
			ID:         r.ID,
			DomainID:   domainID,
			DomainName: domain.Name,
			NodeName:   nodeName,
			RecordType: r.Type,
			TTL:        r.TTL,
			State:      true,
			Content:    r.Content,
			Priority:   r.Priority,
			UpdatedOn:  time.Now().Format(time.RFC3339),
		})
	}

	return records, nil
}

func (p *Provider) CreateRecord(ctx context.Context, apiKey string, domainID string, record *models.Record) (*models.Record, error) {
	username, password, err := parseAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	// Get domain info
	domain, err := p.GetDomain(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}

	client := NewClient(username, password)

	ttl := record.TTL
	if ttl == 0 {
		ttl = 86400 // Default TTL for HE
	}

	heRecord := Record{
		ZoneID:   domainID,
		Name:     record.NodeName,
		Type:     record.RecordType,
		Content:  record.Content,
		TTL:      ttl,
		Priority: record.Priority,
	}

	newRecordID, err := client.CreateRecord(ctx, domainID, heRecord)
	if err != nil {
		return nil, err
	}

	// Return the created record with the actual ID
	return &models.Record{
		ID:         newRecordID,
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
	username, password, err := parseAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	client := NewClient(username, password)

	ttl := record.TTL
	if ttl == 0 {
		ttl = 86400
	}

	heRecord := Record{
		ID:       record.ID,
		ZoneID:   domainID,
		Name:     record.NodeName,
		Type:     record.RecordType,
		Content:  record.Content,
		TTL:      ttl,
		Priority: record.Priority,
	}

	// Pass domain name for building full record name
	err = client.UpdateRecord(ctx, domainID, record.DomainName, heRecord)
	if err != nil {
		return nil, err
	}

	record.UpdatedOn = time.Now().Format(time.RFC3339)
	return record, nil
}

func (p *Provider) DeleteRecord(ctx context.Context, apiKey string, domainID string, recordID string) error {
	username, password, err := parseAPIKey(apiKey)
	if err != nil {
		return err
	}

	client := NewClient(username, password)
	return client.DeleteRecord(ctx, domainID, recordID)
}
