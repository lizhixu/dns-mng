package dnshe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://api005.dnshe.com/index.php?m=domain_hub"

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(ctx context.Context, apiKey, apiSecret, method, endpoint, action string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s&endpoint=%s&action=%s", baseURL, endpoint, action)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("X-API-Secret", apiSecret)

	return c.httpClient.Do(req)
}

// ListSubdomains lists all subdomains
func (c *Client) ListSubdomains(ctx context.Context, apiKey, apiSecret string) (*SubdomainsResponse, error) {
	resp, err := c.doRequest(ctx, apiKey, apiSecret, "GET", "subdomains", "list", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result SubdomainsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &result, nil
}

// GetSubdomain gets subdomain details with DNS records
func (c *Client) GetSubdomain(ctx context.Context, apiKey, apiSecret string, subdomainID int) (*SubdomainDetailResponse, error) {
	url := fmt.Sprintf("%s&endpoint=subdomains&action=get&subdomain_id=%d", baseURL, subdomainID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("X-API-Secret", apiSecret)
	req.Header.Set("Accept", "application/json")

	fmt.Printf("DNSHE GetSubdomain Request: %s\n", url)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	fmt.Printf("DNSHE GetSubdomain Response (status %d): %s\n", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result SubdomainDetailResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &result, nil
}

// ListDNSRecords lists all DNS records for a subdomain
func (c *Client) ListDNSRecords(ctx context.Context, apiKey, apiSecret string, subdomainID int) (*DNSRecordsResponse, error) {
	url := fmt.Sprintf("%s&endpoint=dns_records&action=list&subdomain_id=%d", baseURL, subdomainID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("X-API-Secret", apiSecret)
	req.Header.Set("Accept", "application/json")

	fmt.Printf("DNSHE ListDNSRecords Request: %s\n", url)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	fmt.Printf("DNSHE ListDNSRecords Response (status %d): %s\n", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result DNSRecordsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &result, nil
}

// CreateDNSRecord creates a new DNS record
func (c *Client) CreateDNSRecord(ctx context.Context, apiKey, apiSecret string, subdomainID int, recordName, recordType, content string, ttl int, priority *int) error {
	data := map[string]interface{}{
		"subdomain_id": subdomainID,
		"type":         recordType,
		"content":      content,
		"ttl":          ttl,
	}
	
	// Add name if provided (optional field)
	if recordName != "" {
		data["name"] = recordName
	}
	
	if priority != nil {
		data["priority"] = *priority
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	fmt.Printf("DNSHE CreateDNSRecord Request: %s\n", string(jsonData))

	resp, err := c.doRequest(ctx, apiKey, apiSecret, "POST", "dns_records", "create", bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	fmt.Printf("DNSHE CreateDNSRecord Response (status %d): %s\n", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// UpdateDNSRecord updates an existing DNS record
func (c *Client) UpdateDNSRecord(ctx context.Context, apiKey, apiSecret string, recordID int, content string, ttl int, priority *int) error {
	data := map[string]interface{}{
		"record_id": recordID,
		"content":   content,
		"ttl":       ttl,
	}
	if priority != nil {
		data["priority"] = *priority
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.doRequest(ctx, apiKey, apiSecret, "POST", "dns_records", "update", bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteDNSRecord deletes a DNS record
func (c *Client) DeleteDNSRecord(ctx context.Context, apiKey, apiSecret string, recordID int) error {
	data := map[string]interface{}{
		"record_id": recordID,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.doRequest(ctx, apiKey, apiSecret, "POST", "dns_records", "delete", bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
