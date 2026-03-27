package ndjp

import (
	"context"
	"fmt"
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

// New is an alias for NewProvider for consistency with other providers
func New() *Provider {
	return NewProvider()
}

func (p *Provider) Name() string {
	return "ndjp"
}

func (p *Provider) DisplayName() string {
	return "NDJP NET"
}

func (p *Provider) WebsiteURL() string {
	return "https://manage.ndjp.net"
}

func (p *Provider) ListDomains(ctx context.Context, apiKey string) ([]models.Domain, error) {
	resp, err := p.client.ListDomains(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	// API returns array of subdomain strings like ["nancybrandy", "ps1"]
	// We need to append .ndjp.net to create FQDN
	domains := make([]models.Domain, 0, len(resp.Data))
	for _, subdomain := range resp.Data {
		fqdn := subdomain + ".ndjp.net"
		domains = append(domains, models.Domain{
			ID:          subdomain, // Use subdomain as ID
			Name:        fqdn,
			UnicodeName: fqdn,
			State:       "Active",
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
	records, err := p.client.ListRecords(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}

	domain, err := p.GetDomain(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}

	result := make([]models.Record, 0, len(records))
	for i, r := range records {
		recordID := fmt.Sprintf("%s-%s-%s-%d", r.Type, r.Name, r.Content, i)

		ttl := r.TTL
		if ttl == 0 {
			ttl = 300
		}

		// Handle node name - API returns FQDN like "ps1.ndjp.net."
		nodeName := r.Name
		// Remove trailing dot
		nodeName = strings.TrimSuffix(nodeName, ".")
		
		// If it's the domain itself, use empty string (root)
		if nodeName == domain.Name {
			nodeName = ""
		} else if strings.HasSuffix(nodeName, "."+domain.Name) {
			// Remove domain suffix to get subdomain
			nodeName = strings.TrimSuffix(nodeName, "."+domain.Name)
		}

		result = append(result, models.Record{
			ID:         recordID,
			DomainID:   domainID,
			DomainName: domain.Name,
			NodeName:   nodeName,
			RecordType: r.Type,
			TTL:        ttl,
			State:      true,
			Content:    r.Content,
			UpdatedOn:  time.Now().Format(time.RFC3339),
		})
	}

	return result, nil
}

func (p *Provider) CreateRecord(ctx context.Context, apiKey string, domainID string, record *models.Record) (*models.Record, error) {
	ndjpRecord := Record{
		Type:    record.RecordType,
		Name:    record.NodeName,
		Content: record.Content,
		TTL:     record.TTL,
	}

	if ndjpRecord.Name == "" {
		ndjpRecord.Name = "@"
	}

	err := p.client.AddRecord(ctx, apiKey, domainID, ndjpRecord)
	if err != nil {
		return nil, err
	}

	recordID := fmt.Sprintf("%s-%s-%s-new", record.RecordType, record.NodeName, record.Content)
	return &models.Record{
		ID:         recordID,
		DomainID:   domainID,
		DomainName: record.DomainName,
		NodeName:   record.NodeName,
		RecordType: record.RecordType,
		TTL:        record.TTL,
		State:      true,
		Content:    record.Content,
		Priority:   record.Priority,
		UpdatedOn:  time.Now().Format(time.RFC3339),
	}, nil
}

func (p *Provider) UpdateRecord(ctx context.Context, apiKey string, domainID string, record *models.Record) (*models.Record, error) {
	ndjpRecord := Record{
		Type:    record.RecordType,
		Name:    record.NodeName,
		Content: record.Content,
		TTL:     record.TTL,
	}

	if ndjpRecord.Name == "" {
		ndjpRecord.Name = "@"
	}

	err := p.client.UpdateRecord(ctx, apiKey, domainID, ndjpRecord)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (p *Provider) DeleteRecord(ctx context.Context, apiKey string, domainID string, recordID string) error {
	parts := strings.Split(recordID, "-")
	if len(parts) < 3 {
		return fmt.Errorf("invalid record ID format: %s", recordID)
	}

	recordType := parts[0]
	name := parts[1]
	content := strings.Join(parts[2:len(parts)-1], "-")

	if name == "" {
		name = "@"
	}

	return p.client.DeleteRecord(ctx, apiKey, domainID, recordType, name, content)
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
