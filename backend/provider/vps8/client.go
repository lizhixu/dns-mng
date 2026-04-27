package vps8

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://vps8.zz.cd/api/client/dnsopenapi"

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

func (c *Client) doRequest(ctx context.Context, apiKey, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth("client", apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 429 {
		resp.Body.Close()
		return nil, fmt.Errorf("rate limit exceeded")
	}

	return resp, nil
}

// ListDomains lists all domains
func (c *Client) ListDomains(ctx context.Context, apiKey string) ([]Domain, error) {
	resp, err := c.doRequest(ctx, apiKey, "/domain_list", map[string]string{})
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

	var result DomainListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w (body: %s)", err, string(body))
	}

	return result.Result, nil
}

// ListRecords lists all DNS records for a domain
func (c *Client) ListRecords(ctx context.Context, apiKey, domain string) ([]Record, error) {
	reqBody := map[string]string{
		"domain": domain,
	}

	resp, err := c.doRequest(ctx, apiKey, "/record_list", reqBody)
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

	var result RecordListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w (body: %s)", err, string(body))
	}

	return result.Result, nil
}

// CreateRecord creates a DNS record
func (c *Client) CreateRecord(ctx context.Context, apiKey, domain string, record Record) error {
	reqBody := CreateRecordRequest{
		Domain:  domain,
		Name:    record.Name,
		Type:    record.Type,
		Content: record.Content,
		TTL:     record.TTL,
	}

	resp, err := c.doRequest(ctx, apiKey, "/record_create", reqBody)
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

// UpdateRecord updates a DNS record
func (c *Client) UpdateRecord(ctx context.Context, apiKey, domain string, recordID int, content string, ttl int) error {
	reqBody := UpdateRecordRequest{
		Domain:  domain,
		ID:      recordID,
		Content: content,
		TTL:     ttl,
	}

	resp, err := c.doRequest(ctx, apiKey, "/record_update", reqBody)
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

// DeleteRecord deletes a DNS record
func (c *Client) DeleteRecord(ctx context.Context, apiKey, domain string, recordID int) error {
	reqBody := DeleteRecordRequest{
		Domain: domain,
		ID:     recordID,
	}

	resp, err := c.doRequest(ctx, apiKey, "/record_delete", reqBody)
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
