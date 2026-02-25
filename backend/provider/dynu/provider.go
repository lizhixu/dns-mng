package dynu

import (
	"context"
	"fmt"

	"dns-mng/models"
)

// Provider implements the DNSProvider interface for Dynu.com
type Provider struct {
	client *Client
}

func New() *Provider {
	return &Provider{
		client: NewClient(),
	}
}

func (p *Provider) Name() string {
	return "dynu"
}

func (p *Provider) DisplayName() string {
	return "Dynu.com"
}

func (p *Provider) ListDomains(ctx context.Context, apiKey string) ([]models.Domain, error) {
	resp, err := p.client.GetDomains(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	domains := make([]models.Domain, 0, len(resp.Domains))
	for _, d := range resp.Domains {
		domains = append(domains, models.Domain{
			ID:          fmt.Sprintf("%d", d.ID),
			Name:        d.Name,
			UnicodeName: d.UnicodeName,
			State:       d.State,
			Group:       d.Group,
			IPv4Address: d.IPv4Address,
			IPv6Address: d.IPv6Address,
			TTL:         d.TTL,
			CreatedOn:   d.CreatedOn,
			UpdatedOn:   d.UpdatedOn,
		})
	}
	return domains, nil
}

func (p *Provider) GetDomain(ctx context.Context, apiKey string, domainID string) (*models.Domain, error) {
	d, err := p.client.GetDomain(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}

	return &models.Domain{
		ID:          fmt.Sprintf("%d", d.ID),
		Name:        d.Name,
		UnicodeName: d.UnicodeName,
		State:       d.State,
		Group:       d.Group,
		IPv4Address: d.IPv4Address,
		IPv6Address: d.IPv6Address,
		TTL:         d.TTL,
		CreatedOn:   d.CreatedOn,
		UpdatedOn:   d.UpdatedOn,
	}, nil
}

func (p *Provider) ListRecords(ctx context.Context, apiKey string, domainID string) ([]models.Record, error) {
	// Get domain info to include root A/AAAA records
	domain, err := p.client.GetDomain(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.GetRecords(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}

	records := make([]models.Record, 0, len(resp.DnsRecords)+2)

	// Add root A record if IPv4 address exists
	if domain.IPv4Address != "" {
		records = append(records, models.Record{
			ID:         "root-a",
			DomainID:   fmt.Sprintf("%d", domain.ID),
			DomainName: domain.Name,
			NodeName:   "",
			RecordType: "A",
			TTL:        domain.TTL,
			State:      domain.State == "Active",
			Content:    domain.IPv4Address,
			UpdatedOn:  domain.UpdatedOn,
		})
	}

	// Add root AAAA record if IPv6 address exists
	if domain.IPv6Address != "" {
		records = append(records, models.Record{
			ID:         "root-aaaa",
			DomainID:   fmt.Sprintf("%d", domain.ID),
			DomainName: domain.Name,
			NodeName:   "",
			RecordType: "AAAA",
			TTL:        domain.TTL,
			State:      domain.State == "Active",
			Content:    domain.IPv6Address,
			UpdatedOn:  domain.UpdatedOn,
		})
	}

	// Add subdomain records
	for _, r := range resp.DnsRecords {
		// Skip NS records as they are not editable
		if r.RecordType == "NS" {
			continue
		}

		// Resolve content: prefer type-specific fields over generic content
		var content string
		switch r.RecordType {
		case "A":
			content = r.IPv4Address
		case "AAAA":
			content = r.IPv6Address
		case "CNAME":
			content = r.Host
		case "MX":
			content = r.Host
		case "TXT", "SPF":
			content = r.TextData
		case "SRV":
			content = r.Host
		}
		if content == "" {
			content = r.Content
		}
		records = append(records, models.Record{
			ID:         fmt.Sprintf("%d", r.ID),
			DomainID:   fmt.Sprintf("%d", r.DomainID),
			DomainName: r.DomainName,
			NodeName:   r.NodeName,
			RecordType: r.RecordType,
			TTL:        r.TTL,
			State:      r.State,
			Content:    content,
			Priority:   r.Priority,
			UpdatedOn:  r.UpdatedOn,
		})
	}
	return records, nil
}

func (p *Provider) CreateRecord(ctx context.Context, apiKey string, domainID string, record *models.Record) (*models.Record, error) {
	// Handle root A/AAAA records - update domain instead
	if record.NodeName == "" && (record.RecordType == "A" || record.RecordType == "AAAA") {
		return p.updateRootRecord(ctx, apiKey, domainID, record)
	}

	body := p.buildRecordBody(record)
	r, err := p.client.CreateRecord(ctx, apiKey, domainID, body)
	if err != nil {
		return nil, err
	}
	return p.dynuRecordToModel(r), nil
}

func (p *Provider) UpdateRecord(ctx context.Context, apiKey string, domainID string, record *models.Record) (*models.Record, error) {
	// Handle root A/AAAA records - update domain instead
	if record.NodeName == "" && (record.RecordType == "A" || record.RecordType == "AAAA") {
		return p.updateRootRecord(ctx, apiKey, domainID, record)
	}

	body := p.buildRecordBody(record)
	r, err := p.client.UpdateRecord(ctx, apiKey, domainID, record.ID, body)
	if err != nil {
		return nil, err
	}
	return p.dynuRecordToModel(r), nil
}

// updateRootRecord updates the domain's root A/AAAA record via /dns/{id} endpoint
func (p *Provider) updateRootRecord(ctx context.Context, apiKey string, domainID string, record *models.Record) (*models.Record, error) {
	// First, get the current domain info to preserve existing settings
	currentDomain, err := p.client.GetDomain(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}

	// Build update body with all required fields
	body := map[string]interface{}{
		"name":                currentDomain.Name,
		"group":               currentDomain.Group,
		"ipv4Address":         currentDomain.IPv4Address,
		"ipv6Address":         currentDomain.IPv6Address,
		"ttl":                 currentDomain.TTL,
		"ipv4":                currentDomain.IPv4,
		"ipv6":                currentDomain.IPv6,
		"ipv4WildcardAlias":   currentDomain.IPv4WildcardAlias,
		"ipv6WildcardAlias":   currentDomain.IPv6WildcardAlias,
	}

	// Determine record type from ID if not provided
	recordType := record.RecordType
	recordID := record.ID
	if recordID == "" {
		// Creating new root record
		if recordType == "A" {
			recordID = "root-a"
		} else if recordType == "AAAA" {
			recordID = "root-aaaa"
		}
	} else {
		// Updating existing root record - derive type from ID
		if recordID == "root-a" {
			recordType = "A"
		} else if recordID == "root-aaaa" {
			recordType = "AAAA"
		}
	}

	// Update the specific IP address
	if recordType == "A" {
		body["ipv4Address"] = record.Content
		body["ipv4"] = true
	} else if recordType == "AAAA" {
		body["ipv6Address"] = record.Content
		body["ipv6"] = true
	}

	// Update the domain
	_, err = p.client.UpdateDomain(ctx, apiKey, domainID, body)
	if err != nil {
		return nil, err
	}

	// Fetch the updated domain info to get accurate values
	domain, err := p.client.GetDomain(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}

	// Get the updated content
	content := domain.IPv4Address
	if recordType == "AAAA" {
		content = domain.IPv6Address
	}

	return &models.Record{
		ID:         recordID,
		DomainID:   fmt.Sprintf("%d", domain.ID),
		DomainName: domain.Name,
		NodeName:   "",
		RecordType: recordType,
		TTL:        domain.TTL,
		State:      domain.State == "Active",
		Content:    content,
		UpdatedOn:  domain.UpdatedOn,
	}, nil
}

func (p *Provider) DeleteRecord(ctx context.Context, apiKey string, domainID string, recordID string) error {
	// Handle root A/AAAA records - clear domain's IP address
	if recordID == "root-a" || recordID == "root-aaaa" {
		// Get current domain info to preserve existing settings
		currentDomain, err := p.client.GetDomain(ctx, apiKey, domainID)
		if err != nil {
			return err
		}

		// Build update body with all required fields
		body := map[string]interface{}{
			"name":                currentDomain.Name,
			"group":               currentDomain.Group,
			"ipv4Address":         currentDomain.IPv4Address,
			"ipv6Address":         currentDomain.IPv6Address,
			"ttl":                 currentDomain.TTL,
			"ipv4":                currentDomain.IPv4,
			"ipv6":                currentDomain.IPv6,
			"ipv4WildcardAlias":   currentDomain.IPv4WildcardAlias,
			"ipv6WildcardAlias":   currentDomain.IPv6WildcardAlias,
		}

		// Clear the specific IP address
		if recordID == "root-a" {
			body["ipv4Address"] = ""
			body["ipv4"] = false
		} else {
			body["ipv6Address"] = ""
			body["ipv6"] = false
		}

		_, err = p.client.UpdateDomain(ctx, apiKey, domainID, body)
		return err
	}

	return p.client.DeleteRecord(ctx, apiKey, domainID, recordID)
}

func (p *Provider) buildRecordBody(record *models.Record) map[string]interface{} {
	body := map[string]interface{}{
		"nodeName":   record.NodeName,
		"recordType": record.RecordType,
		"ttl":        record.TTL,
		"state":      record.State,
	}

	switch record.RecordType {
	case "A":
		body["ipv4Address"] = record.Content
		body["group"] = ""
	case "AAAA":
		body["ipv6Address"] = record.Content
		body["group"] = ""
	case "CNAME":
		body["host"] = record.Content
	case "MX":
		body["host"] = record.Content
		body["priority"] = record.Priority
	case "TXT", "SPF":
		body["textData"] = record.Content
	case "SRV":
		body["host"] = record.Content
		body["priority"] = record.Priority
	default:
		body["textData"] = record.Content
	}

	return body
}

func (p *Provider) dynuRecordToModel(r *DynuRecord) *models.Record {
	// Resolve content: prefer type-specific fields over generic content
	var content string
	switch r.RecordType {
	case "A":
		content = r.IPv4Address
	case "AAAA":
		content = r.IPv6Address
	case "CNAME", "MX", "SRV":
		content = r.Host
	case "TXT", "SPF":
		content = r.TextData
	}
	if content == "" {
		content = r.Content
	}
	return &models.Record{
		ID:         fmt.Sprintf("%d", r.ID),
		DomainID:   fmt.Sprintf("%d", r.DomainID),
		DomainName: r.DomainName,
		NodeName:   r.NodeName,
		RecordType: r.RecordType,
		TTL:        r.TTL,
		State:      r.State,
		Content:    content,
		Priority:   r.Priority,
		UpdatedOn:  r.UpdatedOn,
	}
}
