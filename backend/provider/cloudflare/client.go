package cloudflare

import (
	"bytes"
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
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Status      string   `json:"status"`
	NameServers []string `json:"name_servers"`
	CreatedOn   string   `json:"created_on"`
	ModifiedOn  string   `json:"modified_on"`
}

// Record represents a Cloudflare DNS record
type Record struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Name       string `json:"name"`
	Content    string `json:"content"`
	TTL        int    `json:"ttl"`
	Proxied    bool   `json:"proxied"`
	Priority   *int   `json:"priority,omitempty"`
	ZoneID     string `json:"zone_id"`
	ZoneName   string `json:"zone_name"`
	CreatedOn  string `json:"created_on"`
	ModifiedOn string `json:"modified_on"`
}

// APIResponse is the standard Cloudflare API response
type APIResponse struct {
	Success    bool        `json:"success"`
	Errors     []APIError  `json:"errors"`
	Messages   []string    `json:"messages"`
	Result     interface{} `json:"result"`
	ResultInfo *ResultInfo `json:"result_info,omitempty"`
}

// ResultInfo represents Cloudflare pagination metadata
type ResultInfo struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Count      int `json:"count"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

// APIError represents a Cloudflare API error
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *Client) ListZones(ctx context.Context, apiToken string) ([]Zone, error) {
	const perPage = 50

	var allZones []Zone
	for page := 1; ; page++ {
		path := fmt.Sprintf("/zones?page=%d&per_page=%d", page, perPage)
		resp, err := c.doRequest(ctx, apiToken, "GET", path, nil)
		if err != nil {
			return nil, err
		}

		if err := c.parseResponse(resp); err != nil {
			resp.Body.Close()
			return nil, err
		}

		var apiResp APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decode response: %w", err)
		}
		resp.Body.Close()

		if !apiResp.Success && len(apiResp.Errors) > 0 {
			return nil, fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
		}

		var zones []Zone
		data, _ := json.Marshal(apiResp.Result)
		if err := json.Unmarshal(data, &zones); err != nil {
			return nil, fmt.Errorf("decode zones: %w", err)
		}
		allZones = append(allZones, zones...)

		if apiResp.ResultInfo == nil {
			break
		}
		if apiResp.ResultInfo.TotalPages <= page || len(zones) == 0 {
			break
		}
	}

	return allZones, nil
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

	// Build request body
	reqBody := map[string]interface{}{
		"type":    recordType,
		"name":    name,
		"content": content,
		"ttl":     ttl,
	}

	// Add priority for MX and SRV records
	if recordType == "MX" || recordType == "SRV" {
		reqBody["priority"] = priority
	}

	// Set proxied to false by default (gray cloud)
	// Only A, AAAA, and CNAME records can be proxied
	if recordType == "A" || recordType == "AAAA" || recordType == "CNAME" {
		reqBody["proxied"] = false
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.doRequest(ctx, apiToken, "POST", path, strings.NewReader(string(data)))
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

	// Build request body
	reqBody := map[string]interface{}{
		"type":    recordType,
		"name":    name,
		"content": content,
		"ttl":     ttl,
	}

	// Add priority for MX and SRV records
	if recordType == "MX" || recordType == "SRV" {
		reqBody["priority"] = priority
	}

	// Set proxied to false by default (gray cloud)
	// Only A, AAAA, and CNAME records can be proxied
	if recordType == "A" || recordType == "AAAA" || recordType == "CNAME" {
		reqBody["proxied"] = false
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.doRequest(ctx, apiToken, "PUT", path, strings.NewReader(string(data)))
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

func (c *Client) UpdateRecordWithProxied(ctx context.Context, apiToken, zoneID, recordID string, recordType, name, content string, ttl int, proxied bool) (*Record, error) {
	path := "/zones/" + zoneID + "/dns_records/" + recordID

	reqBody := map[string]interface{}{
		"type":    recordType,
		"name":    name,
		"content": content,
		"ttl":     ttl,
	}

	if recordType == "A" || recordType == "AAAA" || recordType == "CNAME" {
		reqBody["proxied"] = proxied
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.doRequest(ctx, apiToken, "POST", path, strings.NewReader(string(data)))
	if err != nil {
		// Cloudflare API supports PUT for update, but let's make sure we send PUT.
		// Wait, let's check: the URL path is same, but method is PUT.
		// Yes, let's use PUT!
	}
	resp, err = c.doRequest(ctx, apiToken, "PUT", path, strings.NewReader(string(data)))
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

// cfErrorCode1039 is the Cloudflare error code returned when a DNS record is
// configured as the fallback origin for SSL for SaaS and cannot be deleted
// until the fallback origin configuration is removed.
const cfErrorCode1039 = 1039

func (c *Client) DeleteRecord(ctx context.Context, apiToken, zoneID, recordID string) error {
	if err := c.deleteRecordOnce(ctx, apiToken, zoneID, recordID); err == nil {
		return nil
	} else if !isCFErrorCode(err, cfErrorCode1039) {
		return err
	}

	// The record is a fallback origin for SSL for SaaS. Remove the fallback
	// origin configuration first, then retry the record deletion.
	if err := c.DeleteFallbackOrigin(ctx, apiToken, zoneID); err != nil {
		return fmt.Errorf("remove fallback origin before deleting record: %w", err)
	}
	return c.deleteRecordOnce(ctx, apiToken, zoneID, recordID)
}

// deleteRecordOnce performs a single DNS record deletion attempt and returns
// an error that preserves the Cloudflare error code (if any).
func (c *Client) deleteRecordOnce(ctx context.Context, apiToken, zoneID, recordID string) error {
	path := "/zones/" + zoneID + "/dns_records/" + recordID
	resp, err := c.doRequest(ctx, apiToken, "DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Cloudflare returns 400 with a JSON body for business-rule errors (e.g.
	// code 1039). parseResponse only checks the HTTP status and discards the
	// body, so read and parse the body ourselves to preserve the error code.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		var apiResp APIResponse
		if json.Unmarshal(body, &apiResp) == nil && len(apiResp.Errors) > 0 {
			return cfError{code: apiResp.Errors[0].Code, message: apiResp.Errors[0].Message}
		}
		return fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	return nil
}

// cfError wraps a Cloudflare API error preserving its numeric code so callers
// can branch on specific error codes (e.g. 1039 fallback origin).
type cfError struct {
	code    int
	message string
}

func (e cfError) Error() string {
	return fmt.Sprintf("API error: %s", e.message)
}

// isCFErrorCode reports whether err is a cfError with the given code.
func isCFErrorCode(err error, code int) bool {
	ce, ok := err.(cfError)
	return ok && ce.code == code
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

// GetAccountID returns the account ID associated with the API token.
func (c *Client) GetAccountID(ctx context.Context, apiToken string) (string, error) {
	resp, err := c.doRequest(ctx, apiToken, "GET", "/accounts", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err := c.parseResponse(resp); err != nil {
		return "", err
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if !apiResp.Success && len(apiResp.Errors) > 0 {
		return "", fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
	}

	// accounts 返回的是数组，每个元素有 id/name
	type accountItem struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	var accounts []accountItem
	data, _ := json.Marshal(apiResp.Result)
	if err := json.Unmarshal(data, &accounts); err != nil {
		return "", fmt.Errorf("decode accounts: %w", err)
	}
	if len(accounts) == 0 {
		return "", fmt.Errorf("no Cloudflare accounts found for this API token")
	}
	return accounts[0].ID, nil
}

// CreateZone creates a new zone under the given Cloudflare account.
func (c *Client) CreateZone(ctx context.Context, apiToken, accountID, domainName string) (*Zone, error) {
	body := map[string]interface{}{
		"name": domainName,
		"account": map[string]string{
			"id": accountID,
		},
		"type": "full",
	}
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.doRequest(ctx, apiToken, "POST", "/zones", bytes.NewReader(jsonData))
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
	if err := json.Unmarshal(data, &zone); err != nil {
		return nil, fmt.Errorf("decode zone: %w", err)
	}
	return &zone, nil
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

// CreateRecordWithProxied creates a DNS record with explicit proxied flag
func (c *Client) CreateRecordWithProxied(ctx context.Context, apiToken, zoneID string, recordType, name, content string, ttl int, proxied bool) (*Record, error) {
	path := "/zones/" + zoneID + "/dns_records"

	reqBody := map[string]interface{}{
		"type":    recordType,
		"name":    name,
		"content": content,
		"ttl":     ttl,
	}

	if recordType == "A" || recordType == "AAAA" || recordType == "CNAME" {
		reqBody["proxied"] = proxied
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.doRequest(ctx, apiToken, "POST", path, strings.NewReader(string(data)))
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

// CreateCustomHostname creates a custom hostname (Cloudflare for SaaS)
func (c *Client) CreateCustomHostname(ctx context.Context, apiToken, zoneID, hostname, originServer string) (*CustomHostname, error) {
	path := "/zones/" + zoneID + "/custom_hostnames"

	reqBody := map[string]interface{}{
		"hostname": hostname,
		"ssl": map[string]interface{}{
			"method": "http",
			"type":   "dv",
			"settings": map[string]interface{}{
				"http2":           "on",
				"min_tls_version": "1.2",
			},
		},
	}

	if originServer != "" {
		reqBody["custom_origin_server"] = originServer
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.doRequest(ctx, apiToken, "POST", path, strings.NewReader(string(data)))
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

	var ch CustomHostname
	resultData, _ := json.Marshal(apiResp.Result)
	json.Unmarshal(resultData, &ch)

	return &ch, nil
}

// SetFallbackOrigin sets or updates the fallback origin for custom hostnames in a zone
func (c *Client) SetFallbackOrigin(ctx context.Context, apiToken, zoneID, origin string) error {
	path := "/zones/" + zoneID + "/custom_hostnames/fallback_origin"

	reqBody := map[string]interface{}{
		"origin": origin,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.doRequest(ctx, apiToken, "PUT", path, strings.NewReader(string(data)))
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

// DeleteFallbackOrigin deletes the fallback origin configuration for custom hostnames in a zone
func (c *Client) DeleteFallbackOrigin(ctx context.Context, apiToken, zoneID string) error {
	path := "/zones/" + zoneID + "/custom_hostnames/fallback_origin"

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

// ListCustomHostnames lists all custom hostnames for a zone
func (c *Client) ListCustomHostnames(ctx context.Context, apiToken, zoneID string) ([]CustomHostname, error) {
	path := "/zones/" + zoneID + "/custom_hostnames"
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

	var chs []CustomHostname
	data, _ := json.Marshal(apiResp.Result)
	json.Unmarshal(data, &chs)

	return chs, nil
}

// GetCustomHostname gets a single custom hostname by ID
func (c *Client) GetCustomHostname(ctx context.Context, apiToken, zoneID, chID string) (*CustomHostname, error) {
	path := "/zones/" + zoneID + "/custom_hostnames/" + chID
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

	var ch CustomHostname
	resultData, _ := json.Marshal(apiResp.Result)
	json.Unmarshal(resultData, &ch)

	return &ch, nil
}

// DeleteCustomHostname deletes a custom hostname by ID
func (c *Client) DeleteCustomHostname(ctx context.Context, apiToken, zoneID, chID string) error {
	path := "/zones/" + zoneID + "/custom_hostnames/" + chID
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

// CustomHostname represents a Cloudflare custom hostname (for SaaS)
type CustomHostname struct {
	ID                    string                      `json:"id"`
	Hostname              string                      `json:"hostname"`
	Status                string                      `json:"status"`
	SSL                   *CustomHostnameSSL          `json:"ssl,omitempty"`
	OwnershipVerification *CustomHostnameVerification `json:"ownership_verification,omitempty"`
	VerificationErrors    []string                    `json:"verification_errors,omitempty"`
	CreatedOn             string                      `json:"created_on,omitempty"`
}

// CustomHostnameSSL represents SSL config of a custom hostname
type CustomHostnameSSL struct {
	Status            string                `json:"status"`
	Method            string                `json:"method,omitempty"`
	Type              string                `json:"type,omitempty"`
	CnameTarget       string                `json:"cname_target,omitempty"`
	CnameName         string                `json:"cname_name,omitempty"`
	ValidationRecords []SSLValidationRecord `json:"validation_records,omitempty"`
}

type SSLValidationRecord struct {
	Status   string `json:"status"`
	TxtName  string `json:"txt_name,omitempty"`
	TxtValue string `json:"txt_value,omitempty"`
}

// CustomHostnameVerification represents ownership verification info
type CustomHostnameVerification struct {
	Status string `json:"status"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Value  string `json:"value"`
}

// CustomHostnameVerificationInfo holds the TXT records needed for verification
type CustomHostnameVerificationInfo struct {
	HostnameTXTName  string `json:"hostname_txt_name"`
	HostnameTXTValue string `json:"hostname_txt_value"`
	SSLTXTName       string `json:"ssl_txt_name"`
	SSLTXTValue      string `json:"ssl_txt_value"`
}

// TTL constants for Cloudflare
const (
	TTLAuto      = 1
	TTL1Minute   = 60
	TTL2Minutes  = 120
	TTL5Minutes  = 300
	TTL10Minutes = 600
	TTL30Minutes = 1800
	TTL1Hour     = 3600
	TTL2Hours    = 7200
	TTL5Hours    = 18000
	TTL12Hours   = 43200
	TTL1Day      = 86400
	TTL5Days     = 432000
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
