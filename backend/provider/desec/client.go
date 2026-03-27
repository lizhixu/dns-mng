package desec

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

const baseURL = "https://desec.io/api/v1"

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

func (c *Client) doRequest(ctx context.Context, token, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Token "+token)

	return c.httpClient.Do(req)
}

// ListDomains lists all domains
func (c *Client) ListDomains(ctx context.Context, token string) ([]Domain, error) {
	resp, err := c.doRequest(ctx, token, "GET", "/domains/", nil)
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

	var domains []Domain
	if err := json.Unmarshal(body, &domains); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return domains, nil
}

// ListRRSets lists all RRSets for a domain
func (c *Client) ListRRSets(ctx context.Context, token, domain string) ([]RRSet, error) {
	path := fmt.Sprintf("/domains/%s/rrsets/", url.PathEscape(domain))
	resp, err := c.doRequest(ctx, token, "GET", path, nil)
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

	var rrsets []RRSet
	if err := json.Unmarshal(body, &rrsets); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return rrsets, nil
}

// CreateRRSet creates a new RRSet
func (c *Client) CreateRRSet(ctx context.Context, token, domain string, rrset RRSetRequest) error {
	path := fmt.Sprintf("/domains/%s/rrsets/", url.PathEscape(domain))

	data, err := json.Marshal(rrset)
	if err != nil {
		return fmt.Errorf("marshal rrset: %w", err)
	}

	resp, err := c.doRequest(ctx, token, "POST", path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// UpdateRRSet updates an existing RRSet
func (c *Client) UpdateRRSet(ctx context.Context, token, domain string, rrset RRSetRequest) error {
	path := fmt.Sprintf("/domains/%s/rrsets/%s/%s/",
		url.PathEscape(domain),
		url.PathEscape(rrset.Subname),
		url.PathEscape(rrset.Type))

	data, err := json.Marshal(rrset)
	if err != nil {
		return fmt.Errorf("marshal rrset: %w", err)
	}

	resp, err := c.doRequest(ctx, token, "PUT", path, bytes.NewReader(data))
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

// DeleteRRSet deletes an RRSet
func (c *Client) DeleteRRSet(ctx context.Context, token, domain, subname, recordType string) error {
	path := fmt.Sprintf("/domains/%s/rrsets/%s/%s/",
		url.PathEscape(domain),
		url.PathEscape(subname),
		url.PathEscape(recordType))

	resp, err := c.doRequest(ctx, token, "DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
