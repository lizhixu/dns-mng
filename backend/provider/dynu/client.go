package dynu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://api.dynu.com/v2"

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

func (c *Client) doRequest(ctx context.Context, method, path, apiKey string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetDomains lists all domains
func (c *Client) GetDomains(ctx context.Context, apiKey string) (*DynuDomainsResponse, error) {
	data, err := c.doRequest(ctx, http.MethodGet, "/dns", apiKey, nil)
	if err != nil {
		return nil, err
	}
	var result DynuDomainsResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse domains response: %w", err)
	}
	return &result, nil
}

// GetDomain gets a specific domain
func (c *Client) GetDomain(ctx context.Context, apiKey string, domainID string) (*DynuDomain, error) {
	data, err := c.doRequest(ctx, http.MethodGet, "/dns/"+domainID, apiKey, nil)
	if err != nil {
		return nil, err
	}
	var result DynuDomain
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse domain response: %w", err)
	}
	return &result, nil
}

// GetRecords lists all DNS records for a domain
func (c *Client) GetRecords(ctx context.Context, apiKey string, domainID string) (*DynuRecordsResponse, error) {
	data, err := c.doRequest(ctx, http.MethodGet, "/dns/"+domainID+"/record", apiKey, nil)
	if err != nil {
		return nil, err
	}
	var result DynuRecordsResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse records response: %w", err)
	}
	return &result, nil
}

// CreateRecord creates a new DNS record
func (c *Client) CreateRecord(ctx context.Context, apiKey string, domainID string, record map[string]interface{}) (*DynuRecord, error) {
	data, err := c.doRequest(ctx, http.MethodPost, "/dns/"+domainID+"/record", apiKey, record)
	if err != nil {
		return nil, err
	}
	var result DynuRecord
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse create record response: %w", err)
	}
	return &result, nil
}

// UpdateRecord updates an existing DNS record
func (c *Client) UpdateRecord(ctx context.Context, apiKey string, domainID string, recordID string, record map[string]interface{}) (*DynuRecord, error) {
	data, err := c.doRequest(ctx, http.MethodPost, "/dns/"+domainID+"/record/"+recordID, apiKey, record)
	if err != nil {
		return nil, err
	}
	var result DynuRecord
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse update record response: %w", err)
	}
	return &result, nil
}

// DeleteRecord deletes a DNS record
func (c *Client) DeleteRecord(ctx context.Context, apiKey string, domainID string, recordID string) error {
	_, err := c.doRequest(ctx, http.MethodDelete, "/dns/"+domainID+"/record/"+recordID, apiKey, nil)
	return err
}

// UpdateDomain updates the domain settings (for root A/AAAA records)
func (c *Client) UpdateDomain(ctx context.Context, apiKey string, domainID string, body map[string]interface{}) (*DynuDomain, error) {
	data, err := c.doRequest(ctx, http.MethodPost, "/dns/"+domainID, apiKey, body)
	if err != nil {
		return nil, err
	}
	var result DynuDomain
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse update domain response: %w", err)
	}
	return &result, nil
}
