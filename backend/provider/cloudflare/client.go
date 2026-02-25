package cloudflare

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

const baseURL = "https://api.cloudflare.com/client/v4"

// Client implements Cloudflare API client
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

// parseAPIKey parses the API key - now supports API Token only
func (c *Client) parseAPIKey(apiKey string) (string, error) {
	// Cloudflare API Token format (recommended)
	// Just the token itself, no email needed for API Token
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return "", fmt.Errorf("API token is required")
	}
	return apiKey, nil
}

func (c *Client) doRequest(ctx context.Context, apiToken, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiToken)

	return c.httpClient.Do(req)
}

func (c *Client) parseResponse(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}
	return nil
}

// Zone represents a Cloudflare zone
type Zone struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// Record represents a Cloudflare DNS record
type Record struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Content  string `json:"content"`
	TTL      int    `json:"ttl"`
	Proxied  bool   `json:"proxied"`
	ZoneID   string `json:"zone_id"`
	ZoneName string `json:"zone_name"`
}

// APIResponse is the standard Cloudflare API response
type APIResponse struct {
	Success bool        `json:"success"`
	Errors  []APIError  `json:"errors"`
	Messages []string   `json:"messages"`
	Result  interface{} `json:"result"`
}

// APIError represents a Cloudflare API error
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *Client) ListZones(ctx context.Context, apiToken string) ([]Zone, error) {
	resp, err := c.doRequest(ctx, apiToken, "GET", "/zones", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.parseResponse(resp); err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if !apiResp.Success && len(apiResp.Errors) > 0 {
		return nil, fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
	}

	var zones []Zone
	data, _ := json.Marshal(apiResp.Result)
	json.Unmarshal(data, &zones)

	return zones, nil
}

func (c *Client) GetZone(ctx context.Context, apiToken, zoneID string) (*Zone, error) {
	resp, err := c.doRequest(ctx, apiToken, "GET", "/zones/"+zoneID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.parseResponse(resp); err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if !apiResp.Success && len(apiResp.Errors) > 0 {
		return nil, fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
	}

	var zone Zone
	data, _ := json.Marshal(apiResp.Result)
	json.Unmarshal(data, &zone)

	return &zone, nil
}

func (c *Client) ListRecords(ctx context.Context, apiToken, zoneID string) ([]Record, error) {
	path := "/zones/" + zoneID + "/dns_records"
	resp, err := c.doRequest(ctx, apiToken, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.parseResponse(resp); err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if !apiResp.Success && len(apiResp.Errors) > 0 {
		return nil, fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
	}

	var records []Record
	data, _ := json.Marshal(apiResp.Result)
	json.Unmarshal(data, &records)

	return records, nil
}

func (c *Client) CreateRecord(ctx context.Context, apiToken, zoneID string, recordType, name, content string, ttl int, priority int) (*Record, error) {
	path := "/zones/" + zoneID + "/dns_records"

	data := fmt.Sprintf(`{
		"type": "%s",
		"name": "%s",
		"content": "%s",
		"ttl": %d,
		"priority": %d
	}`, recordType, name, content, ttl, priority)

	resp, err := c.doRequest(ctx, apiToken, "POST", path, strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.parseResponse(resp); err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if !apiResp.Success && len(apiResp.Errors) > 0 {
		return nil, fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
	}

	var record Record
	resultData, _ := json.Marshal(apiResp.Result)
	json.Unmarshal(resultData, &record)

	return &record, nil
}

func (c *Client) UpdateRecord(ctx context.Context, apiToken, zoneID, recordID string, recordType, name, content string, ttl int, priority int) (*Record, error) {
	path := "/zones/" + zoneID + "/dns_records/" + recordID

	data := fmt.Sprintf(`{
		"type": "%s",
		"name": "%s",
		"content": "%s",
		"ttl": %d,
		"priority": %d
	}`, recordType, name, content, ttl, priority)

	resp, err := c.doRequest(ctx, apiToken, "PUT", path, strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.parseResponse(resp); err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if !apiResp.Success && len(apiResp.Errors) > 0 {
		return nil, fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
	}

	var record Record
	resultData, _ := json.Marshal(apiResp.Result)
	json.Unmarshal(resultData, &record)

	return &record, nil
}

func (c *Client) DeleteRecord(ctx context.Context, apiToken, zoneID, recordID string) error {
	path := "/zones/" + zoneID + "/dns_records/" + recordID
	resp, err := c.doRequest(ctx, apiToken, "DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := c.parseResponse(resp); err != nil {
		return err
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	if !apiResp.Success && len(apiResp.Errors) > 0 {
		return fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
	}

	return nil
}

// GetZoneByName finds a zone by name
func (c *Client) GetZoneByName(ctx context.Context, apiToken, domainName string) (*Zone, error) {
	// URL encode the domain name
	encodedName := url.QueryEscape(domainName)
	resp, err := c.doRequest(ctx, apiToken, "GET", "/zones?name="+encodedName, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.parseResponse(resp); err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if !apiResp.Success && len(apiResp.Errors) > 0 {
		return nil, fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
	}

	var zones []Zone
	data, _ := json.Marshal(apiResp.Result)
	json.Unmarshal(data, &zones)

	if len(zones) == 0 {
		return nil, fmt.Errorf("zone not found: %s", domainName)
	}

	return &zones[0], nil
}

// GetRecordByID gets a single record by ID
func (c *Client) GetRecordByID(ctx context.Context, apiToken, zoneID, recordID string) (*Record, error) {
	path := "/zones/" + zoneID + "/dns_records/" + recordID
	resp, err := c.doRequest(ctx, apiToken, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.parseResponse(resp); err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if !apiResp.Success && len(apiResp.Errors) > 0 {
		return nil, fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
	}

	var record Record
	resultData, _ := json.Marshal(apiResp.Result)
	json.Unmarshal(resultData, &record)

	return &record, nil
}

// TTL constants for Cloudflare
const (
	TTLAuto     = 1
	TTL1Minute  = 60
	TTL2Minutes = 120
	TTL5Minutes = 300
	TTL10Minutes = 600
	TTL30Minutes = 1800
	TTL1Hour    = 3600
	TTL2Hours   = 7200
	TTL5Hours   = 18000
	TTL12Hours  = 43200
	TTL1Day     = 86400
	TTL5Days    = 432000
)

// ConvertTTL converts our TTL to Cloudflare's expected format
func ConvertTTL(ttl int) int {
	if ttl <= 0 {
		return TTLAuto
	}
	return ttl
}

// safeString converts *string to string
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// safeInt converts *int to int
func safeInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// safeBool converts *bool to bool
func safeBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// strconv helpers
func parseInt(s string) int {
	if s == "" {
		return 0
	}
	i, _ := strconv.Atoi(s)
	return i
}