package desec

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

// New is an alias for NewProvider
func New() *Provider {
	return NewProvider()
}

func (p *Provider) Name() string {
	return "desec"
}

func (p *Provider) DisplayName() string {
	return "deSEC"
}

func (p *Provider) WebsiteURL() string {
	return "https://desec.io"
}

func (p *Provider) ListDomains(ctx context.Context, apiKey string) ([]models.Domain, error) {
	domains, err := p.client.ListDomains(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	result := make([]models.Domain, 0, len(domains))
	for _, d := range domains {
		// Use touched as updated time, fallback to published or created
		updatedOn := d.Touched
		if updatedOn == "" {
			updatedOn = d.Published
		}
		if updatedOn == "" {
			updatedOn = d.Created
		}

		result = append(result, models.Domain{
			ID:          d.Name,
			Name:        d.Name,
			UnicodeName: d.Name,
			State:       "Active",
			CreatedOn:   d.Created,
			UpdatedOn:   updatedOn,
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
	rrsets, err := p.client.ListRRSets(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}

	domain, err := p.GetDomain(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}

	result := make([]models.Record, 0)
	for _, rrset := range rrsets {
		// Skip system records
		if rrset.Type == "SOA" || rrset.Type == "NS" {
			continue
		}

		// Use touched time as updated time, fallback to created time
		updatedOn := rrset.Touched
		if updatedOn == "" {
			updatedOn = rrset.Created
		}

		for i, record := range rrset.Records {
			recordID := fmt.Sprintf("%s-%s-%s-%d", rrset.Type, rrset.Subname, record, i)

			result = append(result, models.Record{
				ID:         recordID,
				DomainID:   domainID,
				DomainName: domain.Name,
				NodeName:   rrset.Subname,
				RecordType: rrset.Type,
				TTL:        rrset.TTL,
				State:      true,
				Content:    record,
				UpdatedOn:  updatedOn,
			})
		}
	}

	return result, nil
}

func (p *Provider) CreateRecord(ctx context.Context, apiKey string, domainID string, record *models.Record) (*models.Record, error) {
	ttl := record.TTL
	if ttl < 3600 {
		ttl = 3600 // deSEC minimum TTL is 3600
	}

	rrset := RRSetRequest{
		Subname: record.NodeName,
		Type:    record.RecordType,
		Records: []string{record.Content},
		TTL:     ttl,
	}

	err := p.client.CreateRRSet(ctx, apiKey, domainID, rrset)
	if err != nil {
		return nil, err
	}

	// Get domain info for domain name
	domain, err := p.GetDomain(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}

	recordID := fmt.Sprintf("%s-%s-%s-new", record.RecordType, record.NodeName, record.Content)
	return &models.Record{
		ID:         recordID,
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
	ttl := record.TTL
	if ttl < 3600 {
		ttl = 3600 // deSEC minimum TTL is 3600
	}

	rrset := RRSetRequest{
		Subname: record.NodeName,
		Type:    record.RecordType,
		Records: []string{record.Content},
		TTL:     ttl,
	}

	err := p.client.UpdateRRSet(ctx, apiKey, domainID, rrset)
	if err != nil {
		return nil, err
	}

	record.TTL = ttl
	return record, nil
}

func (p *Provider) DeleteRecord(ctx context.Context, apiKey string, domainID string, recordID string) error {
	parts := strings.Split(recordID, "-")
	if len(parts) < 3 {
		return fmt.Errorf("invalid record ID format: %s", recordID)
	}

	recordType := parts[0]
	subname := parts[1]

	return p.client.DeleteRRSet(ctx, apiKey, domainID, subname, recordType)
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
