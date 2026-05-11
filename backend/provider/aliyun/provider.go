package aliyun

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dns-mng/models"
)

// Provider implements DNSProvider for Alibaba Cloud DNS (Alidns).
type Provider struct {
	client *Client
}

func New() *Provider {
	return &Provider{
		client: NewClient(),
	}
}

func (p *Provider) Name() string {
	return "aliyun"
}

func (p *Provider) DisplayName() string {
	return "阿里云云解析 DNS"
}

func (p *Provider) WebsiteURL() string {
	return "https://dns.console.aliyun.com"
}

func (p *Provider) DefaultTTL() int {
	return 600
}

func (p *Provider) ListDomains(ctx context.Context, apiKey string) ([]models.Domain, error) {
	list, err := p.client.ListDomains(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	domains := make([]models.Domain, 0, len(list))
	for _, d := range list {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if d.DomainName == "" {
			continue
		}
		allAli, err := p.client.AllAliDnsDelegation(ctx, apiKey, d.DomainName)
		if err != nil || !allAli {
			continue
		}
		status := "Active"
		if d.InstanceExpired {
			status = "Inactive"
		}
		renewal := ""
		if t := strings.TrimSpace(d.InstanceEndTime); t != "" {
			parts := strings.SplitN(t, " ", 2)
			if len(parts) > 0 {
				renewal = parts[0]
			}
		}
		domains = append(domains, models.Domain{
			ID:          d.DomainName,
			Name:        d.DomainName,
			UnicodeName: d.DomainName,
			State:       status,
			RenewalDate: renewal,
			CreatedOn:   strings.TrimSpace(d.CreateTime),
		})
	}
	return domains, nil
}

func (p *Provider) GetDomain(ctx context.Context, apiKey string, domainID string) (*models.Domain, error) {
	info, err := p.client.DescribeDomain(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(info.DomainName)
	if name == "" {
		name = domainID
	}
	return &models.Domain{
		ID:          name,
		Name:        name,
		UnicodeName: name,
		State:       "Active",
		CreatedOn:   strings.TrimSpace(info.CreateTime),
	}, nil
}

func (p *Provider) ListRecords(ctx context.Context, apiKey string, domainID string) ([]models.Record, error) {
	recs, err := p.client.ListAllRecords(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}

	out := make([]models.Record, 0, len(recs))
	for _, r := range recs {
		if r.RecordId == "" || r.Type == "" {
			continue
		}
		if strings.EqualFold(r.Type, "NS") {
			continue
		}
		line := strings.TrimSpace(r.Line)
		if line == "" {
			line = defaultRecordLine
		}
		ttl := int(r.TTL)
		if ttl <= 0 {
			ttl = 600
		}
		state := strings.EqualFold(r.Status, "ENABLE")
		priority := int(r.Priority)
		out = append(out, models.Record{
			ID:         r.RecordId,
			DomainID:   domainID,
			DomainName: domainID,
			NodeName:   rrToNodeName(r.RR),
			RecordType: strings.ToUpper(strings.TrimSpace(r.Type)),
			TTL:        ttl,
			State:      state,
			Content:    r.Value,
			Priority:   priority,
			Raw:        map[string]interface{}{"line": line},
		})
	}
	return out, nil
}

func (p *Provider) CreateRecord(ctx context.Context, apiKey string, domainID string, record *models.Record) (*models.Record, error) {
	ttl := int64(record.TTL)
	if ttl <= 0 {
		ttl = 600
	}
	line := recordLineFromRaw(record.Raw)
	rr := nodeNameToRR(record.NodeName)
	recType := strings.ToUpper(strings.TrimSpace(record.RecordType))

	id, err := p.client.AddDomainRecord(ctx, apiKey, domainID, rr, line, recType, record.Content, ttl, int64(record.Priority))
	if err != nil {
		return nil, err
	}
	status := "ENABLE"
	if !record.State {
		status = "DISABLE"
	}
	if err := p.client.SetRecordStatus(ctx, apiKey, id, status); err != nil {
		return nil, err
	}
	return &models.Record{
		ID:         id,
		DomainID:   domainID,
		DomainName: domainID,
		NodeName:   record.NodeName,
		RecordType: recType,
		TTL:        int(ttl),
		State:      record.State,
		Content:    record.Content,
		Priority:   record.Priority,
		Raw:        map[string]interface{}{"line": line},
		UpdatedOn:  time.Now().Format(time.RFC3339),
	}, nil
}

func (p *Provider) UpdateRecord(ctx context.Context, apiKey string, domainID string, record *models.Record) (*models.Record, error) {
	if strings.TrimSpace(record.ID) == "" {
		return nil, fmt.Errorf("missing record id")
	}
	ttl := int64(record.TTL)
	if ttl <= 0 {
		ttl = 600
	}
	line := recordLineFromRaw(record.Raw)
	rr := nodeNameToRR(record.NodeName)
	recType := strings.ToUpper(strings.TrimSpace(record.RecordType))

	if err := p.client.UpdateDomainRecord(ctx, apiKey, record.ID, rr, line, recType, record.Content, ttl, int64(record.Priority)); err != nil {
		return nil, err
	}
	status := "ENABLE"
	if !record.State {
		status = "DISABLE"
	}
	if err := p.client.SetRecordStatus(ctx, apiKey, record.ID, status); err != nil {
		return nil, err
	}
	record.Raw = map[string]interface{}{"line": line}
	return record, nil
}

func (p *Provider) DeleteRecord(ctx context.Context, apiKey string, domainID string, recordID string) error {
	return p.client.DeleteDomainRecord(ctx, apiKey, recordID)
}
