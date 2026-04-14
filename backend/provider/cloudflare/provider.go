package cloudflare

import (
	"context"
	"strings"
	"time"

	"dns-mng/models"
)

// Provider implements the DNSProvider interface for Cloudflare
type Provider struct {
	client *Client
}

func New() *Provider {
	return &Provider{
		client: NewClient(),
	}
}

func (p *Provider) Name() string {
	return "cloudflare"
}

func (p *Provider) DisplayName() string {
	return "Cloudflare"
}

func (p *Provider) WebsiteURL() string {
	return "https://dash.cloudflare.com"
}

func (p *Provider) DefaultTTL() int {
	return 1
}

func (p *Provider) ListDomains(ctx context.Context, apiKey string) ([]models.Domain, error) {
	apiToken, err := p.client.parseAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	zones, err := p.client.ListZones(ctx, apiToken)
	if err != nil {
		return nil, err
	}

	domains := make([]models.Domain, 0, len(zones))
	for _, z := range zones {
		status := "Active"
		if z.Status != "active" {
			status = "Inactive"
		}

		domains = append(domains, models.Domain{
			ID:          z.ID,
			Name:        z.Name,
			UnicodeName: z.Name,
			State:       status,
			CreatedOn:   z.CreatedOn,
			UpdatedOn:   z.ModifiedOn,
		})
	}
	return domains, nil
}

func (p *Provider) GetDomain(ctx context.Context, apiKey string, domainID string) (*models.Domain, error) {
	apiToken, err := p.client.parseAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	zone, err := p.client.GetZone(ctx, apiToken, domainID)
	if err != nil {
		return nil, err
	}

	status := "Active"
	if zone.Status != "active" {
		status = "Inactive"
	}

	return &models.Domain{
		ID:          zone.ID,
		Name:        zone.Name,
		UnicodeName: zone.Name,
		State:       status,
		CreatedOn:   zone.CreatedOn,
		UpdatedOn:   zone.ModifiedOn,
	}, nil
}

func (p *Provider) ListRecords(ctx context.Context, apiKey string, domainID string) ([]models.Record, error) {
	apiToken, err := p.client.parseAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	// Get zone info for domain name
	zone, err := p.client.GetZone(ctx, apiToken, domainID)
	if err != nil {
		return nil, err
	}

	records, err := p.client.ListRecords(ctx, apiToken, domainID)
	if err != nil {
		return nil, err
	}

	result := make([]models.Record, 0, len(records))
	for _, r := range records {
		// Skip certain record types that are not typically editable
		if r.Type == "SOA" || r.Type == "NS" {
			continue
		}

		// Get TTL - Cloudflare returns 1 for auto
		ttl := 300 // default
		if r.TTL > 1 {
			ttl = r.TTL
		}

		// Get priority for MX/SRV records
		priority := 0
		if r.Type == "MX" || r.Type == "SRV" {
			// Fetch individual record to get priority
			detailRecord, err := p.client.GetRecordByID(ctx, apiToken, domainID, r.ID)
			if err == nil && detailRecord.Priority != nil {
				priority = *detailRecord.Priority
			}
		}

		// Determine if record is active (not paused)
		state := true
		// Cloudflare doesn't have a direct "disabled" state for DNS records

		// Get node name (remove zone name suffix)
		nodeName := strings.TrimSuffix(r.Name, "."+zone.Name)
		if nodeName == zone.Name {
			nodeName = ""
		}
		if nodeName == "" {
			nodeName = "@"
		}

		// Store Cloudflare-specific data in Raw field
		raw := map[string]interface{}{
			"proxied": r.Proxied,
			"zone_id": r.ZoneID,
		}

		result = append(result, models.Record{
			ID:         r.ID,
			DomainID:   domainID,
			DomainName: zone.Name,
			NodeName:   nodeName,
			RecordType: r.Type,
			TTL:        ttl,
			State:      state,
			Content:    r.Content,
			Priority:   priority,
			UpdatedOn:  r.ModifiedOn,
			Raw:        raw,
		})
	}
	return result, nil
}

func (p *Provider) CreateRecord(ctx context.Context, apiKey string, domainID string, record *models.Record) (*models.Record, error) {
	apiToken, err := p.client.parseAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	// Get zone info
	zone, err := p.client.GetZone(ctx, apiToken, domainID)
	if err != nil {
		return nil, err
	}

	// Build full record name
	name := record.NodeName
	if name == "@" || name == "" {
		name = zone.Name
	} else {
		name = name + "." + zone.Name
	}

	ttl := ConvertTTL(record.TTL)
	priority := record.Priority
	if priority == 0 {
		priority = 10 // default priority
	}

	resp, err := p.client.CreateRecord(ctx, apiToken, domainID, record.RecordType, name, record.Content, ttl, priority)
	if err != nil {
		return nil, err
	}

	return &models.Record{
		ID:         resp.ID,
		DomainID:   domainID,
		DomainName: zone.Name,
		NodeName:   record.NodeName,
		RecordType: record.RecordType,
		TTL:        record.TTL,
		State:      true, // Cloudflare doesn't support disabling records
		Content:    record.Content,
		Priority:   record.Priority,
		UpdatedOn:  time.Now().Format(time.RFC3339),
	}, nil
}

func (p *Provider) UpdateRecord(ctx context.Context, apiKey string, domainID string, record *models.Record) (*models.Record, error) {
	apiToken, err := p.client.parseAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	// Get zone info
	zone, err := p.client.GetZone(ctx, apiToken, domainID)
	if err != nil {
		return nil, err
	}

	// Build full record name
	name := record.NodeName
	if name == "@" || name == "" {
		name = zone.Name
	} else {
		name = name + "." + zone.Name
	}

	ttl := ConvertTTL(record.TTL)
	priority := record.Priority
	if priority == 0 {
		priority = 10
	}

	_, err = p.client.UpdateRecord(ctx, apiToken, domainID, record.ID, record.RecordType, name, record.Content, ttl, priority)
	if err != nil {
		return nil, err
	}

	// Return record with state always true for Cloudflare
	record.State = true
	return record, nil
}

func (p *Provider) DeleteRecord(ctx context.Context, apiKey string, domainID string, recordID string) error {
	apiToken, err := p.client.parseAPIKey(apiKey)
	if err != nil {
		return err
	}

	return p.client.DeleteRecord(ctx, apiToken, domainID, recordID)
}
