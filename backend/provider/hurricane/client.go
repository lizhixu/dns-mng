package hurricane

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	baseURL  = "https://dns.he.net"
	indexCGI = "https://dns.he.net/index.cgi"
)

// cleanHTML removes HTML entities from a string
func cleanHTML(s string) string {
	s = strings.ReplaceAll(s, "&quot;", "")
	s = strings.ReplaceAll(s, "&#34;", "")
	s = strings.ReplaceAll(s, "&amp;", "")
	s = strings.ReplaceAll(s, "&#38;", "")
	s = strings.ReplaceAll(s, "&lt;", "")
	s = strings.ReplaceAll(s, "&#60;", "")
	s = strings.ReplaceAll(s, "&gt;", "")
	s = strings.ReplaceAll(s, "&#62;", "")
	s = strings.ReplaceAll(s, "&#39;", "")
	s = strings.ReplaceAll(s, "&#x27;", "")
	return s
}

type Client struct {
	httpClient *http.Client
	username   string
	password   string
	mu         sync.Mutex
	loggedIn   bool
}

func NewClient(username, password string) *Client {
	// Create cookie jar with options to handle cookies properly
	jar, _ := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: nil, // Allow all cookies
	})
	return &Client{
		httpClient: &http.Client{
			// No timeout at all: HE (dns.he.net) is frequently slow or stalls during
			// TLS handshake, even from US nodes. We deliberately disable the overall
			// client Timeout and the Transport dial/TLS handshake timeouts so a slow
			// but eventually-successful request is never aborted proactively. The only
			// bound is the caller-supplied context (if any).
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   0, // no dial timeout
					KeepAlive: 30 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   0, // no TLS handshake timeout
				ResponseHeaderTimeout: 0, // no response-header timeout
				ExpectContinueTimeout: 0, // no expect-continue timeout
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
			},
			Jar: jar,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Follow redirects but preserve cookies
				return nil
			},
		},
		username: username,
		password: password,
		loggedIn: false,
	}
}

// isLoginForm checks if the response is a login page (session expired).
func isLoginForm(body string) bool {
	return strings.Contains(body, `name="login"`) || strings.Contains(body, `name="email"`)
}

// doLogin performs the actual login POST. Caller must hold c.mu.
func (c *Client) doLogin(ctx context.Context) error {
	baseURLParsed, _ := url.Parse(baseURL)

	// Step 1: Visit homepage to get initial cookie
	req1, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/", nil)
	if err != nil {
		return fmt.Errorf("create homepage request: %w", err)
	}
	req1.Header.Set("User-Agent", "Mozilla/5.0")

	resp1, err := c.httpClient.Do(req1)
	if err != nil {
		return fmt.Errorf("homepage request failed: %w", err)
	}
	io.Copy(io.Discard, resp1.Body)
	resp1.Body.Close()

	cookies := c.httpClient.Jar.Cookies(baseURLParsed)

	// Step 2: Login with cookies
	data := url.Values{}
	data.Set("email", c.username)
	data.Set("pass", c.password)
	data.Set("submit", "Login!")

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", baseURL+"/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read login response: %w", err)
	}
	bodyStr := string(body)

	if strings.Contains(bodyStr, "Incorrect") || strings.Contains(bodyStr, "Invalid") {
		return fmt.Errorf("login failed: invalid credentials")
	}
	if !strings.Contains(bodyStr, "domains_table") {
		return fmt.Errorf("login failed: unexpected response")
	}
	c.loggedIn = true

	// If no cookies received, try to fetch the page again to get cookies
	if len(cookies) == 0 {
		time.Sleep(100 * time.Millisecond)
		req2, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/", nil)
		if err == nil {
			req2.Header.Set("User-Agent", "Mozilla/5.0")
			req2.Header.Set("Referer", baseURL+"/")
			resp2, err := c.httpClient.Do(req2)
			if err == nil {
				io.Copy(io.Discard, resp2.Body)
				resp2.Body.Close()
			}
		}
	}
	return nil
}

// login ensures the session is active. Uses cookie jar to persist session.
func (c *Client) login(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.loggedIn {
		return nil
	}
	return c.doLogin(ctx)
}

// resetSession marks the session as expired (called when a request detects a login form).
func (c *Client) resetSession() {
	c.loggedIn = false
}

// ListZones lists all DNS zones by fetching the domains page after login.
// Note: Domain listing is now primarily served from database cache.
// This method is kept as a fallback for refresh scenarios.
func (c *Client) ListZones(ctx context.Context) ([]Domain, error) {
	if err := c.login(ctx); err != nil {
		return nil, err
	}

	// Fetch the domains page
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	html := string(body)

	// Session expired — re-login and retry
	if isLoginForm(html) {
		c.mu.Lock()
		c.resetSession()
		if err := c.doLogin(ctx); err != nil {
			c.mu.Unlock()
			return nil, err
		}
		c.mu.Unlock()

		req2, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/", nil)
		if err != nil {
			return nil, fmt.Errorf("create retry request: %w", err)
		}
		req2.Header.Set("User-Agent", "Mozilla/5.0")
		resp2, err := c.httpClient.Do(req2)
		if err != nil {
			return nil, fmt.Errorf("retry request failed: %w", err)
		}
		defer resp2.Body.Close()
		body2, err := io.ReadAll(resp2.Body)
		if err != nil {
			return nil, fmt.Errorf("read retry response: %w", err)
		}
		html = string(body2)
	}

	return c.parseZones(html)
}

// parseZones extracts zone information from HTML
// As documented: 需要解析html中的 name 和 onclick 后的 javascript
// onclick="javascript:document.location.href='?hosted_dns_zoneid=1303724&menu=edit_zone&hosted_dns_editzone'"
func (c *Client) parseZones(html string) ([]Domain, error) {
	var domains []Domain

	// Parse table rows in tbody
	rowRegex := regexp.MustCompile(`<tr>[\s\S]*?</tr>`)
	rows := rowRegex.FindAllString(html, -1)

	for _, row := range rows {
		// Extract zone ID from onclick attribute - as documented
		// Looking for: hosted_dns_zoneid=1303724 in onclick
		onclickRegex := regexp.MustCompile(`hosted_dns_zoneid=(\d+)`)
		onclickMatch := onclickRegex.FindStringSubmatch(row)
		if len(onclickMatch) < 2 {
			continue
		}
		zoneID := onclickMatch[1]

		// Extract domain name from name attribute - as documented
		// Looking for: name="qzz.frii.site"
		nameRegex := regexp.MustCompile(`name="([^"]+)"`)
		nameMatch := nameRegex.FindStringSubmatch(row)
		if len(nameMatch) < 2 {
			continue
		}
		domainName := nameMatch[1]

		if domainName != "" && zoneID != "" {
			domains = append(domains, Domain{
				ID:     zoneID,
				Name:   domainName,
				Status: "active",
			})
		}
	}

	return domains, nil
}

// ListRecords lists all DNS records for a zone.
// If the response is a login form (session expired), re-logs in and retries once.
func (c *Client) ListRecords(ctx context.Context, zoneID string) ([]Record, error) {
	if err := c.login(ctx); err != nil {
		return nil, err
	}

	body, err := c.fetchRecordsPage(ctx, zoneID)
	if err != nil {
		return nil, err
	}

	// Session expired — re-login and retry
	if isLoginForm(body) {
		c.mu.Lock()
		c.resetSession()
		if err := c.doLogin(ctx); err != nil {
			c.mu.Unlock()
			return nil, err
		}
		c.mu.Unlock()

		body, err = c.fetchRecordsPage(ctx, zoneID)
		if err != nil {
			return nil, err
		}
	}

	return c.parseRecords(body, zoneID)
}

func (c *Client) fetchRecordsPage(ctx context.Context, zoneID string) (string, error) {
	reqURL := fmt.Sprintf("%s/?hosted_dns_zoneid=%s&menu=edit_zone&hosted_dns_editzone", baseURL, zoneID)
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", indexCGI)
	req.Header.Set("Origin", baseURL)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	return string(body), nil
}

// parseRecords extracts DNS records from HTML table
// HTML structure per documentation:
// <tr class="dns_tr" id="9119194957">
//
//	<td class="hidden">1303724</td>           <- Zone ID
//	<td class="hidden">9119194957</td>        <- Record ID
//	<td class="dns_view">qzz.frii.site</td>   <- Name
//	<td align="center"><span class="rrlabel NS">NS</span></td> <- Type
//	<td align="left">172800</td>              <- TTL
//	<td align="center">-</td>                 <- Priority
//	<td align="left" data="ns1.he.net">ns1.he.net</td> <- Content
//
// </tr>
func (c *Client) parseRecords(html string, zoneID string) ([]Record, error) {
	var records []Record

	// Find all table rows with class="dns_tr" (editable records only)
	// Use FindAllStringSubmatch to get the captured group (record ID)
	rowStartRegex := regexp.MustCompile(`<tr class="dns_tr"[^>]*id="(\d+)"`)
	rowEndRegex := regexp.MustCompile(`</tr>`)

	// Find all row matches with captured group
	rowMatches := rowStartRegex.FindAllStringSubmatch(html, -1)

	for _, startMatch := range rowMatches {
		if len(startMatch) < 2 {
			continue
		}
		recordID := startMatch[1]

		// Find the position of this match in the HTML
		startPos := strings.Index(html, startMatch[0])
		if startPos < 0 {
			continue
		}

		// Find the end of this row
		endMatch := rowEndRegex.FindStringIndex(html[startPos:])
		if endMatch == nil {
			continue
		}
		endPos := startPos + endMatch[1]

		row := html[startPos:endPos]

		record := Record{
			ID:     recordID,
			ZoneID: zoneID,
		}

		// Extract name from dns_view td
		// <td width="95%" class="dns_view">qzz.frii.site</td>
		nameRegex := regexp.MustCompile(`<td[^>]*class="dns_view"[^>]*>([^<]+)</td>`)
		if nameMatch := nameRegex.FindStringSubmatch(row); len(nameMatch) > 1 {
			record.Name = strings.TrimSpace(nameMatch[1])
		}

		// Extract type from rrlabel span
		// <span class="rrlabel NS" data="NS">NS</span>
		typeRegex := regexp.MustCompile(`<span class="rrlabel ([A-Z]+)"`)
		if typeMatch := typeRegex.FindStringSubmatch(row); len(typeMatch) > 1 {
			record.Type = typeMatch[1]
		}

		// Extract TTL - first <td align="left"> without data attribute
		// <td align="left">172800</td>
		ttlRegex := regexp.MustCompile(`<td align="left">(\d+)</td>`)
		if ttlMatch := ttlRegex.FindStringSubmatch(row); len(ttlMatch) > 1 {
			if ttl, err := strconv.Atoi(ttlMatch[1]); err == nil {
				record.TTL = ttl
			}
		}

		// Extract priority from <td align="center">
		// <td align="center">-</td> or <td align="center">10</td>
		priorityRegex := regexp.MustCompile(`<td align="center">([^<]+)</td>`)
		priorityMatches := priorityRegex.FindAllStringSubmatch(row, -1)
		for _, pm := range priorityMatches {
			p := strings.TrimSpace(pm[1])
			if p != "-" && p != "" {
				if priority, err := strconv.Atoi(p); err == nil {
					record.Priority = priority
					break
				}
			}
		}

		// Extract content from <td align="left" data="...">
		// <td align="left" data="ns1.he.net">ns1.he.net</td>
		contentRegex := regexp.MustCompile(`<td align="left" data="([^"]+)"`)
		if contentMatch := contentRegex.FindStringSubmatch(row); len(contentMatch) > 1 {
			record.Content = cleanHTML(contentMatch[1])
		}

		if record.ID != "" && record.Type != "" && record.Content != "" {
			records = append(records, record)
		}
	}

	return records, nil
}

// CreateRecord creates a new DNS record and returns the new record ID.
// If the response is a login form (session expired), re-logs in and retries once.
func (c *Client) CreateRecord(ctx context.Context, zoneID string, record Record) (string, error) {
	// Check for duplicate records before creating
	existingRecords, err := c.ListRecords(ctx, zoneID)
	if err == nil {
		for _, r := range existingRecords {
			if r.Name == record.Name && r.Type == record.Type {
				return "", fmt.Errorf("record already exists: %s %s", r.Name, r.Type)
			}
		}
	}

	bodyStr, err := c.postCreateRecord(ctx, zoneID, record)
	if err != nil {
		return "", err
	}

	// Session expired — re-login and retry
	if isLoginForm(bodyStr) {
		c.mu.Lock()
		c.resetSession()
		if err := c.doLogin(ctx); err != nil {
			c.mu.Unlock()
			return "", err
		}
		c.mu.Unlock()

		bodyStr, err = c.postCreateRecord(ctx, zoneID, record)
		if err != nil {
			return "", err
		}
	}

	// Check for error message in dns_err div
	errRegex := regexp.MustCompile(`<div id="dns_err"[^>]*>([^<]+)</div>`)
	if matches := errRegex.FindStringSubmatch(bodyStr); len(matches) > 1 {
		errMsg := strings.TrimSpace(matches[1])
		errMsg = strings.ReplaceAll(errMsg, "&amp;", "&")
		errMsg = strings.ReplaceAll(errMsg, "&lt;", "<")
		errMsg = strings.ReplaceAll(errMsg, "&gt;", ">")
		return "", fmt.Errorf("create record failed: %s", errMsg)
	}

	return "new", nil
}

func (c *Client) postCreateRecord(ctx context.Context, zoneID string, record Record) (string, error) {
	data := url.Values{}
	data.Set("account", "")
	data.Set("menu", "edit_zone")
	data.Set("Type", record.Type)
	data.Set("hosted_dns_zoneid", zoneID)
	data.Set("hosted_dns_recordid", "")
	data.Set("hosted_dns_editzone", "1")
	data.Set("Name", record.Name)
	data.Set("Content", record.Content)
	data.Set("TTL", strconv.Itoa(record.TTL))
	data.Set("hosted_dns_editrecord", "Submit")
	if record.Priority > 0 {
		data.Set("Priority", strconv.Itoa(record.Priority))
	} else {
		data.Set("Priority", "")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", indexCGI, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", fmt.Sprintf("%s/?hosted_dns_zoneid=%s&menu=edit_zone&hosted_dns_editzone", baseURL, zoneID))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("create record failed: status %d", resp.StatusCode)
	}

	return string(body), nil
}

// UpdateRecord updates an existing DNS record.
// If the response is a login form (session expired), re-logs in and retries once.
func (c *Client) UpdateRecord(ctx context.Context, zoneID string, domainName string, record Record) error {
	if err := c.login(ctx); err != nil {
		return err
	}

	bodyStr, err := c.postUpdateRecord(ctx, zoneID, domainName, record)
	if err != nil {
		return err
	}

	// Session expired — re-login and retry
	if isLoginForm(bodyStr) {
		c.mu.Lock()
		c.resetSession()
		if err := c.doLogin(ctx); err != nil {
			c.mu.Unlock()
			return err
		}
		c.mu.Unlock()

		bodyStr, err = c.postUpdateRecord(ctx, zoneID, domainName, record)
		if err != nil {
			return err
		}
	}

	errRegex := regexp.MustCompile(`<div id="dns_err"[^>]*>([^<]+)</div>`)
	if matches := errRegex.FindStringSubmatch(bodyStr); len(matches) > 1 {
		errMsg := strings.TrimSpace(matches[1])
		errMsg = strings.ReplaceAll(errMsg, "&amp;", "&")
		errMsg = strings.ReplaceAll(errMsg, "&lt;", "<")
		errMsg = strings.ReplaceAll(errMsg, "&gt;", ">")
		return fmt.Errorf("update record failed: %s", errMsg)
	}

	return nil
}

func (c *Client) postUpdateRecord(ctx context.Context, zoneID string, domainName string, record Record) (string, error) {
	fullName := record.Name
	if fullName != "" && domainName != "" && !strings.HasSuffix(fullName, domainName) {
		fullName = fullName + "." + domainName
	}

	data := url.Values{}
	data.Set("account", "")
	data.Set("menu", "edit_zone")
	data.Set("Type", record.Type)
	data.Set("hosted_dns_zoneid", zoneID)
	data.Set("hosted_dns_recordid", record.ID)
	data.Set("hosted_dns_editzone", "1")
	data.Set("Name", fullName)
	data.Set("Content", record.Content)
	data.Set("TTL", strconv.Itoa(record.TTL))
	data.Set("hosted_dns_editrecord", "Update")
	if record.Priority > 0 {
		data.Set("Priority", strconv.Itoa(record.Priority))
	} else {
		data.Set("Priority", "-")
	}

	reqURL := fmt.Sprintf("%s/?hosted_dns_zoneid=%s&menu=edit_zone&hosted_dns_editzone", baseURL, zoneID)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	return string(body), nil
}

// DeleteRecord deletes a DNS record.
// If the response is a login form (session expired), re-logs in and retries once.
func (c *Client) DeleteRecord(ctx context.Context, zoneID, recordID string) error {
	if err := c.login(ctx); err != nil {
		return err
	}

	bodyStr, err := c.postDeleteRecord(ctx, zoneID, recordID)
	if err != nil {
		return err
	}

	// Session expired — re-login and retry
	if isLoginForm(bodyStr) {
		c.mu.Lock()
		c.resetSession()
		if err := c.doLogin(ctx); err != nil {
			c.mu.Unlock()
			return err
		}
		c.mu.Unlock()

		bodyStr, err = c.postDeleteRecord(ctx, zoneID, recordID)
		if err != nil {
			return err
		}
	}

	errRegex := regexp.MustCompile(`<div id="dns_err"[^>]*>([^<]+)</div>`)
	if matches := errRegex.FindStringSubmatch(bodyStr); len(matches) > 1 {
		errMsg := strings.TrimSpace(matches[1])
		errMsg = strings.ReplaceAll(errMsg, "&amp;", "&")
		errMsg = strings.ReplaceAll(errMsg, "&lt;", "<")
		errMsg = strings.ReplaceAll(errMsg, "&gt;", ">")
		return fmt.Errorf("delete record failed: %s", errMsg)
	}

	return nil
}

func (c *Client) postDeleteRecord(ctx context.Context, zoneID, recordID string) (string, error) {
	data := url.Values{}
	data.Set("hosted_dns_zoneid", zoneID)
	data.Set("hosted_dns_recordid", recordID)
	data.Set("menu", "edit_zone")
	data.Set("hosted_dns_delconfirm", "delete")
	data.Set("hosted_dns_editzone", "1")
	data.Set("hosted_dns_delrecord", "1")

	req, err := http.NewRequestWithContext(ctx, "POST", indexCGI, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	return string(body), nil
}
