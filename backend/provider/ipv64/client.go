package ipv64

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const baseURL = "https://ipv64.net/api.php"

// API Limit: Maximum 5 API requests within 10 seconds
const rateLimitInterval = 2 * time.Second // Minimum interval between requests

type Client struct {
	httpClient *http.Client
	lastRequestTime time.Time
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// checkRateLimit ensures we don't exceed rate limits (5 requests per 10 seconds)
func (c *Client) checkRateLimit() {
	elapsed := time.Since(c.lastRequestTime)
	if elapsed < rateLimitInterval {
		time.Sleep(rateLimitInterval - elapsed)
	}
	c.lastRequestTime = time.Now()
}

// parseAPIResponse parses and validates the API response
func parseAPIResponse(body []byte, statusCode int) (*APIResponse, error) {
	// First try to parse as standard API response
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	// Check HTTP status code according to API documentation
	// Accept 2xx status codes as success
	if statusCode >= 200 && statusCode < 300 {
		// Check if info field indicates success
		if apiResp.Info == "success" || apiResp.Info == "" {
			return &apiResp, nil
		}
		// If info is not success, return error with details
		return nil, fmt.Errorf("API returned non-success: info=%s, status=%s", apiResp.Info, apiResp.Status)
	}

	// Handle error status codes
	switch statusCode {
	case http.StatusBadRequest:
		return nil, fmt.Errorf("bad request (400): %s", apiResp.Info)
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("unauthorized (401): %s", apiResp.Info)
	case http.StatusForbidden:
		return nil, fmt.Errorf("forbidden (403): %s", apiResp.Info)
	case http.StatusTooManyRequests:
		return nil, fmt.Errorf("rate limit exceeded (429): %s", apiResp.Info)
	default:
		return nil, fmt.Errorf("API error (status %d): %s", statusCode, apiResp.Info)
	}
}

// GetDomains gets all domains and their records
func (c *Client) GetDomains(ctx context.Context, apiKey string) (*GetDomainsResponse, error) {
	c.checkRateLimit()

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"?get_domains", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Handle non-OK status codes
	if resp.StatusCode != http.StatusOK {
		if _, err := parseAPIResponse(body, resp.StatusCode); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result GetDomainsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	// Validate response according to API documentation
	if result.Info != "success" && result.Status != "200 OK" {
		return nil, fmt.Errorf("API error: info=%s, status=%s", result.Info, result.Status)
	}

	return &result, nil
}

// AddDomain creates a new domain
func (c *Client) AddDomain(ctx context.Context, apiKey, domain string) error {
	c.checkRateLimit()

	data := url.Values{}
	data.Set("add_domain", domain)

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// Parse and validate response according to API documentation
	if _, err := parseAPIResponse(body, resp.StatusCode); err != nil {
		return err
	}

	return nil
}

// DeleteDomain deletes a domain
func (c *Client) DeleteDomain(ctx context.Context, apiKey, domain string) error {
	c.checkRateLimit()

	data := url.Values{}
	data.Set("del_domain", domain)

	req, err := http.NewRequestWithContext(ctx, "DELETE", baseURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// Parse and validate response according to API documentation
	if _, err := parseAPIResponse(body, resp.StatusCode); err != nil {
		return err
	}

	return nil
}

// AddRecord creates a new DNS record
func (c *Client) AddRecord(ctx context.Context, apiKey, domain, praefix, recordType, content string) error {
	c.checkRateLimit()

	data := url.Values{}
	data.Set("add_record", domain)
	data.Set("praefix", praefix)
	data.Set("type", recordType)
	data.Set("content", content)

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// Parse and validate response according to API documentation
	if _, err := parseAPIResponse(body, resp.StatusCode); err != nil {
		return err
	}

	return nil
}

// DeleteRecord deletes a DNS record by ID
// According to API docs: [DELETE] del_record => DNS Record ID [Integer Format]
func (c *Client) DeleteRecordByID(ctx context.Context, apiKey string, recordID int) error {
	c.checkRateLimit()

	data := url.Values{}
	data.Set("del_record", strconv.Itoa(recordID))

	req, err := http.NewRequestWithContext(ctx, "DELETE", baseURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// Parse and validate response according to API documentation
	if _, err := parseAPIResponse(body, resp.StatusCode); err != nil {
		return err
	}

	return nil
}

// DeleteRecord deletes a DNS record by domain, praefix, type, and content
// According to API docs: [DELETE] del_record => Domainname, praefix, type, content
func (c *Client) DeleteRecord(ctx context.Context, apiKey, domain, praefix, recordType, content string) error {
	c.checkRateLimit()

	data := url.Values{}
	data.Set("del_record", domain)
	data.Set("praefix", praefix)
	data.Set("type", recordType)
	data.Set("content", content)

	req, err := http.NewRequestWithContext(ctx, "DELETE", baseURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// Parse and validate response according to API documentation
	if _, err := parseAPIResponse(body, resp.StatusCode); err != nil {
		return err
	}

	return nil
}
