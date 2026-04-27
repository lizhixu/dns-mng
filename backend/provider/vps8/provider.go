package vps8

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

// New is an alias for NewProvider for consistency with other providers
func New() *Provider {
	return NewProvider()
}

func (p *Provider) Name() string {
	return "vps8"
}

func (p *Provider) DisplayName() string {
	return "VPS8 DNS"
}

func (p *Provider) WebsiteURL() string {
	return "https://vps8.zz.cd"
}

func (p *Provider) DefaultTTL() int {
	return 100
}

func (p *Provider) ListDomains(ctx context.Context, apiKey string) ([]models.Domain, error) {
	domains, err := p.client.ListDomains(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	result := make([]models.Domain, 0, len(domains))
	for _, d := range domains {
		// Parse expires_at to renewal date format (YYYY-MM-DD)
		renewalDate := ""
		if d.ExpiresAt != "" {
			parts := strings.Split(d.ExpiresAt, " ")
			if len(parts) > 0 {
				renewalDate = parts[0]
			}
		}

		result = append(result, models.Domain{
			ID:          d.Domain, // Use domain name as ID
			Name:        d.Domain,
			UnicodeName: d.Domain,
			State:       "Active",
			RenewalDate: renewalDate,
			RenewalURL:  "https://vps8.zz.cd",
		})
	}

	return result, nil
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
	// Get domain info first
	domain, err := p.GetDomain(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}

	records, err := p.client.ListRecords(ctx, apiKey, domain.Name)
	if err != nil {
		return nil, err
	}

	result := make([]models.Record, 0, len(records))
	for _, r := range records {
		ttl := r.TTL
		if ttl == 0 {
			ttl = 100
		}

		// Handle node name
		nodeName := r.Name
		// If it's the domain itself or @, use empty string (root)
		if nodeName == "@" || nodeName == domain.Name {
			nodeName = ""
		} else if strings.HasSuffix(nodeName, "."+domain.Name) {
			// Remove domain suffix to get subdomain
			nodeName = strings.TrimSuffix(nodeName, "."+domain.Name)
		}

		result = append(result, models.Record{
			ID:         strconv.Itoa(r.ID),
			DomainID:   domainID,
			DomainName: domain.Name,
			NodeName:   nodeName,
			RecordType: r.Type,
			TTL:        ttl,
			State:      true,
			Content:    r.Content,
			Priority:   r.Priority,
			UpdatedOn:  time.Now().Format(time.RFC3339),
		})
	}

	return result, nil
}

func (p *Provider) CreateRecord(ctx context.Context, apiKey string, domainID string, record *models.Record) (*models.Record, error) {
	domain, err := p.GetDomain(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}

	vps8Record := Record{
		Name:    record.NodeName,
		Type:    record.RecordType,
		Content: record.Content,
		TTL:     record.TTL,
	}

	if vps8Record.Name == "" {
		vps8Record.Name = "@"
	}

	err = p.client.CreateRecord(ctx, apiKey, domain.Name, vps8Record)
	if err != nil {
		return nil, err
	}

	return &models.Record{
		DomainID:   domainID,
		DomainName: domain.Name,
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
	domain, err := p.GetDomain(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}

	recordID, err := strconv.Atoi(record.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid record ID: %s", record.ID)
	}

	err = p.client.UpdateRecord(ctx, apiKey, domain.Name, recordID, record.Content, record.TTL)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (p *Provider) DeleteRecord(ctx context.Context, apiKey string, domainID string, recordID string) error {
	domain, err := p.GetDomain(ctx, apiKey, domainID)
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(recordID)
	if err != nil {
		return fmt.Errorf("invalid record ID: %s", recordID)
	}

	return p.client.DeleteRecord(ctx, apiKey, domain.Name, id)
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
