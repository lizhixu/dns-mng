package tencentcloud

import (
	"context"
	"fmt"
	"strconv"

	"dns-mng/models"
)

// Provider implements the DNSProvider interface for Tencent Cloud DNSPod
type Provider struct {
	client *Client
}

func New() *Provider {
	return &Provider{
		client: NewClient(),
	}
}

func (p *Provider) Name() string {
	return "tencentcloud"
}

func (p *Provider) DisplayName() string {
	return "腾讯云 DNSPod"
}

func (p *Provider) ListDomains(ctx context.Context, apiKey string) ([]models.Domain, error) {
	domainList, err := p.client.ListDomains(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	domains := make([]models.Domain, 0, len(domainList))
	for _, d := range domainList {
		if d.Name == nil {
			continue
		}
		
		status := "Active"
		if d.Status != nil && *d.Status != "enable" {
			status = "Inactive"
		}

		domains = append(domains, models.Domain{
			ID:          *d.Name, // Use domain name as ID
			Name:        *d.Name,
			UnicodeName: *d.Name,
			State:       status,
			CreatedOn:   safeString(d.CreatedOn),
			UpdatedOn:   safeString(d.UpdatedOn),
		})
	}
	return domains, nil
}

func (p *Provider) GetDomain(ctx context.Context, apiKey string, domainID string) (*models.Domain, error) {
	domainInfo, err := p.client.GetDomain(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}

	status := "Active"
	if domainInfo.Status != nil && *domainInfo.Status != "enable" {
		status = "Inactive"
	}

	return &models.Domain{
		ID:          domainID,
		Name:        domainID,
		UnicodeName: domainID,
		State:       status,
		CreatedOn:   safeString(domainInfo.CreatedOn),
		UpdatedOn:   safeString(domainInfo.UpdatedOn),
	}, nil
}

func (p *Provider) ListRecords(ctx context.Context, apiKey string, domainID string) ([]models.Record, error) {
	recordList, err := p.client.ListRecords(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}

	records := make([]models.Record, 0, len(recordList))
	for _, r := range recordList {
		if r.RecordId == nil || r.Type == nil {
			continue
		}

		// Skip NS records as they are not editable
		if *r.Type == "NS" {
			continue
		}

		nodeName := ""
		if r.Name != nil && *r.Name != "@" {
			nodeName = *r.Name
		}

		ttl := 600
		if r.TTL != nil {
			ttl = int(*r.TTL)
		}

		state := true
		if r.Status != nil && *r.Status != "ENABLE" {
			state = false
		}

		priority := 0
		if r.MX != nil {
			priority = int(*r.MX)
		}

		records = append(records, models.Record{
			ID:         strconv.FormatUint(*r.RecordId, 10),
			DomainID:   domainID,
			DomainName: domainID,
			NodeName:   nodeName,
			RecordType: *r.Type,
			TTL:        ttl,
			State:      state,
			Content:    safeString(r.Value),
			Priority:   priority,
			UpdatedOn:  safeString(r.UpdatedOn),
		})
	}
	return records, nil
}

func (p *Provider) CreateRecord(ctx context.Context, apiKey string, domainID string, record *models.Record) (*models.Record, error) {
	ttl := uint64(record.TTL)
	if ttl == 0 {
		ttl = 600
	}

	mx := uint64(record.Priority)
	
	resp, err := p.client.CreateRecord(ctx, apiKey, domainID, record.RecordType, record.NodeName, record.Content, ttl, mx)
	if err != nil {
		return nil, err
	}

	if resp.Response.RecordId == nil {
		return nil, fmt.Errorf("no record ID returned")
	}

	return &models.Record{
		ID:         strconv.FormatUint(*resp.Response.RecordId, 10),
		DomainID:   domainID,
		DomainName: domainID,
		NodeName:   record.NodeName,
		RecordType: record.RecordType,
		TTL:        record.TTL,
		State:      record.State,
		Content:    record.Content,
		Priority:   record.Priority,
	}, nil
}

func (p *Provider) UpdateRecord(ctx context.Context, apiKey string, domainID string, record *models.Record) (*models.Record, error) {
	recordID, err := strconv.ParseUint(record.ID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid record ID: %w", err)
	}

	ttl := uint64(record.TTL)
	if ttl == 0 {
		ttl = 600
	}

	mx := uint64(record.Priority)

	err = p.client.UpdateRecord(ctx, apiKey, domainID, recordID, record.RecordType, record.NodeName, record.Content, ttl, mx)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (p *Provider) DeleteRecord(ctx context.Context, apiKey string, domainID string, recordID string) error {
	id, err := strconv.ParseUint(recordID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid record ID: %w", err)
	}

	return p.client.DeleteRecord(ctx, apiKey, domainID, id)
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
