package ipv64

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

func New() *Provider {
	return &Provider{
		client: NewClient(),
	}
}

func (p *Provider) Name() string {
	return "ipv64"
}

func (p *Provider) DisplayName() string {
	return "IPv64.net"
}

func (p *Provider) WebsiteURL() string {
	return "https://ipv64.net"
}

// parseTime converts IPv64 time format to RFC3339
// IPv64 format: "2026-04-13 04:16:29"
// RFC3339 format: "2006-01-02T15:04:05Z07:00"
func parseTime(ipv64Time string) string {
	if ipv64Time == "" {
		return time.Now().Format(time.RFC3339)
	}

	// Try to parse IPv64 time format
	t, err := time.Parse("2006-01-02 15:04:05", ipv64Time)
	if err != nil {
		// If parsing fails, return current time
		return time.Now().Format(time.RFC3339)
	}

	return t.Format(time.RFC3339)
}

func (p *Provider) ListDomains(ctx context.Context, apiKey string) ([]models.Domain, error) {
	resp, err := p.client.GetDomains(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	result := make([]models.Domain, 0, len(resp.Subdomains))
	for domainName, domainInfo := range resp.Subdomains {
		// 使用记录中最新的更新时间作为域名的更新时间
		var latestUpdate string
		for _, record := range domainInfo.Records {
			if record.LastUpdate != "" {
				if latestUpdate == "" || record.LastUpdate > latestUpdate {
					latestUpdate = record.LastUpdate
				}
			}
		}

		// 如果没有记录或记录没有更新时间，使用当前时间
		if latestUpdate == "" {
			latestUpdate = time.Now().Format("2006-01-02 15:04:05")
		}

		state := "Active"
		if domainInfo.Deactivated != 0 {
			state = "Inactive"
		}

		result = append(result, models.Domain{
			ID:        domainName,
			Name:      domainName,
			State:     state,
			UpdatedOn: parseTime(latestUpdate),
		})
	}

	return result, nil
}

func (p *Provider) GetDomain(ctx context.Context, apiKey string, domainID string) (*models.Domain, error) {
	resp, err := p.client.GetDomains(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	if info, ok := resp.Subdomains[domainID]; ok {
		state := "Active"
		if info.Deactivated != 0 {
			state = "Inactive"
		}
		return &models.Domain{
			ID:    domainID,
			Name:  domainID,
			State: state,
		}, nil
	}

	return nil, fmt.Errorf("domain not found: %s", domainID)
}

func (p *Provider) ListRecords(ctx context.Context, apiKey, domainID string) ([]models.Record, error) {
	resp, err := p.client.GetDomains(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	domainInfo, ok := resp.Subdomains[domainID]
	if !ok {
		return nil, fmt.Errorf("domain not found: %s", domainID)
	}

	result := make([]models.Record, 0, len(domainInfo.Records))
	for _, r := range domainInfo.Records {
		// Parse node name from praefix
		nodeName := r.Praefix
		if nodeName == "" {
			nodeName = "@"
		}

		result = append(result, models.Record{
			ID:         strconv.Itoa(r.RecordID),
			DomainID:   domainID,
			DomainName: domainID,
			NodeName:   nodeName,
			RecordType: r.Type,
			TTL:        r.TTL,
			State:      r.Deactivated == 0,
			Content:    r.Content,
			UpdatedOn:  parseTime(r.LastUpdate),
		})
	}

	return result, nil
}

func (p *Provider) CreateRecord(ctx context.Context, apiKey, domainID string, record *models.Record) (*models.Record, error) {
	// Convert node name to praefix
	praefix := record.NodeName
	if praefix == "@" || praefix == "" {
		praefix = ""
	}

	err := p.client.AddRecord(ctx, apiKey, domainID, praefix, record.RecordType, record.Content)
	if err != nil {
		return nil, err
	}

	// IPv64 API doesn't return the created record, so we need to fetch all records
	// and find the one we just created
	resp, err := p.client.GetDomains(ctx, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created record: %w", err)
	}

	domainInfo, ok := resp.Subdomains[domainID]
	if !ok {
		return nil, fmt.Errorf("domain not found after creating record: %s", domainID)
	}

	// Find the newly created record by matching praefix, type, and content
	for _, r := range domainInfo.Records {
		if r.Praefix == praefix && r.Type == record.RecordType && r.Content == record.Content {
			nodeName := r.Praefix
			if nodeName == "" {
				nodeName = "@"
			}
			return &models.Record{
				ID:         strconv.Itoa(r.RecordID),
				DomainID:   domainID,
				DomainName: domainID,
				NodeName:   nodeName,
				RecordType: r.Type,
				TTL:        r.TTL,
				State:      r.Deactivated == 0,
				Content:    r.Content,
				UpdatedOn:  parseTime(r.LastUpdate),
			}, nil
		}
	}

	// If we can't find the record, return a basic one
	// This shouldn't happen, but just in case
	return &models.Record{
		ID:         "0",
		DomainID:   domainID,
		DomainName: domainID,
		NodeName:   record.NodeName,
		RecordType: record.RecordType,
		TTL:        60,
		State:      true,
		Content:    record.Content,
		UpdatedOn:  time.Now().Format(time.RFC3339),
	}, nil
}

func (p *Provider) UpdateRecord(ctx context.Context, apiKey, domainID string, record *models.Record) (*models.Record, error) {
	// IPv64 doesn't have an update API, so we need to delete and recreate
	// First, get the old record to know what to delete
	recordID, err := strconv.Atoi(record.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid record ID: %s", record.ID)
	}

	// Delete the old record
	err = p.client.DeleteRecordByID(ctx, apiKey, recordID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete old record: %w", err)
	}

	// Create the new record
	praefix := record.NodeName
	if praefix == "@" || praefix == "" {
		praefix = ""
	}

	err = p.client.AddRecord(ctx, apiKey, domainID, praefix, record.RecordType, record.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to create new record: %w", err)
	}

	return record, nil
}

func (p *Provider) DeleteRecord(ctx context.Context, apiKey, domainID, recordID string) error {
	// Try to parse as integer ID first
	id, err := strconv.Atoi(recordID)
	if err == nil {
		return p.client.DeleteRecordByID(ctx, apiKey, id)
	}

	// If not an integer, it might be a composite ID (domain:praefix:type:content)
	parts := strings.Split(recordID, ":")
	if len(parts) >= 4 {
		return p.client.DeleteRecord(ctx, apiKey, parts[0], parts[1], parts[2], parts[3])
	}

	return fmt.Errorf("invalid record ID format: %s", recordID)
}
