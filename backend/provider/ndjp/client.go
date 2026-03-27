package ndjp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://manage.ndjp.net/api"

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

func (c *Client) doRequest(ctx context.Context, apiToken, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Handle rate limiting (429 Too Many Requests)
	if resp.StatusCode == 429 {
		resp.Body.Close()
		return nil, fmt.Errorf("rate limit exceeded: NDJP allows max 10 requests per minute")
	}

	return resp, nil
}

// ListDomains lists all subdomains
func (c *Client) ListDomains(ctx context.Context, apiToken string) (*DomainsResponse, error) {
	resp, err := c.doRequest(ctx, apiToken, "GET", "/domains", nil)
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

	// Debug: print raw response
	fmt.Printf("NDJP API Response: %s\n", string(body))

	var result DomainsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w (body: %s)", err, string(body))
	}

	return &result, nil
}

// ListRecords lists all DNS records for a subdomain
func (c *Client) ListRecords(ctx context.Context, apiToken, subdomain string) ([]Record, error) {
	path := fmt.Sprintf("/domains/%s/records", url.PathEscape(subdomain))
	resp, err := c.doRequest(ctx, apiToken, "GET", path, nil)
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

	// Parse the response - data can be string or array of RRSets
	var rawResp map[string]interface{}
	if err := json.Unmarshal(body, &rawResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	// Debug: print raw response
	fmt.Printf("NDJP ListRecords Response: %s\n", string(body))

	// Handle empty or string data
	data, ok := rawResp["data"]
	if !ok || data == nil {
		return []Record{}, nil
	}

	// If data is a string (like empty zone), return empty array
	if _, isString := data.(string); isString {
		return []Record{}, nil
	}

	// Parse as array of RRSets
	dataBytes, _ := json.Marshal(data)
	var rrsets []RRSet
	if err := json.Unmarshal(dataBytes, &rrsets); err != nil {
		return nil, fmt.Errorf("parse rrsets: %w (data: %s)", err, string(dataBytes))
	}

	// Convert RRSets to flat Record list
	var records []Record
	for _, rrset := range rrsets {
		for _, rr := range rrset.Records {
			records = append(records, Record{
				Type:    rrset.Type,
				Name:    rrset.Name,
				Content: rr.Content,
				TTL:     rrset.TTL,
			})
		}
	}

	return records, nil
}

// AddRecord adds a DNS record
func (c *Client) AddRecord(ctx context.Context, apiToken, subdomain string, record Record) error {
	path := fmt.Sprintf("/domains/%s/records", url.PathEscape(subdomain))
	
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal record: %w", err)
	}

	resp, err := c.doRequest(ctx, apiToken, "POST", path, bytes.NewReader(data))
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
func (c *Client) UpdateRecord(ctx context.Context, apiToken, subdomain string, record Record) error {
	path := fmt.Sprintf("/domains/%s/records", url.PathEscape(subdomain))
	
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal record: %w", err)
	}

	resp, err := c.doRequest(ctx, apiToken, "PUT", path, bytes.NewReader(data))
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
func (c *Client) DeleteRecord(ctx context.Context, apiToken, subdomain, recordType, name, content string) error {
	path := fmt.Sprintf("/domains/%s/records?type=%s&name=%s",
		url.PathEscape(subdomain),
		url.QueryEscape(recordType),
		url.QueryEscape(name))
	
	if content != "" {
		path += "&content=" + url.QueryEscape(content)
	}

	resp, err := c.doRequest(ctx, apiToken, "DELETE", path, nil)
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
