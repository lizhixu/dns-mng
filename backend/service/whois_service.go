package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"dns-mng/database"
	"dns-mng/models"
)

// ErrWHOISAPIKeyNotConfigured is returned when WHOIS lookup is requested but
// the user has not configured an API key.
var ErrWHOISAPIKeyNotConfigured = errors.New("WHOIS API key is not configured")

// ErrWHOISAPIKeyRequired is returned when the user tries to create a config
// without supplying an API key. An empty API key is only valid for updates
// (meaning "keep current"). Mirrors the EmailConfig distinction where a fresh
// record requires all credentials.
var ErrWHOISAPIKeyRequired = errors.New("WHOIS API key is required")

// whoisJSONEndpoint is the upstream WhoisJSON.com lookup endpoint.
const whoisJSONEndpoint = "https://whoisjson.com/api/v1/whois"

// WHOISService manages per-user WHOIS lookup configuration and proxies lookups
// to the upstream WhoisJSON.com API. The service is stateless; configuration is
// read from the database on demand.
type WHOISService struct {
	client *http.Client
}

// NewWHOISService creates a WHOISService with a 10s HTTP client timeout.
func NewWHOISService() *WHOISService {
	return &WHOISService{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// GetConfig returns the user's WHOIS configuration, including the API key in
// plaintext. The key belongs to the user themselves (same pattern as the DDNS
// token: returned in clear text so the user can view/edit it). Returns
// (nil, nil) when no row exists — the handler maps this to
// `{"configured": false}`.
func (s *WHOISService) GetConfig(userID int64) (*models.WHOISConfig, error) {
	var cfg models.WHOISConfig

	err := database.DB.QueryRow(
		`SELECT id, user_id, api_key, created_at, updated_at
		 FROM whois_config WHERE user_id = ?`,
		userID,
	).Scan(&cfg.ID, &cfg.UserID, &cfg.APIKey, &cfg.CreatedAt, &cfg.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &cfg, nil
}

// getConfigWithKey reads the raw config (including the API key). Used internally
// before calling upstream, mirroring EmailService.getEmailConfigWithPassword.
func (s *WHOISService) getConfigWithKey(userID int64) (*models.WHOISConfig, error) {
	var cfg models.WHOISConfig

	err := database.DB.QueryRow(
		`SELECT id, user_id, api_key, created_at, updated_at
		 FROM whois_config WHERE user_id = ?`,
		userID,
	).Scan(&cfg.ID, &cfg.UserID, &cfg.APIKey, &cfg.CreatedAt, &cfg.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &cfg, nil
}

// UpsertConfig creates or updates the user's WHOIS configuration. When
// req.APIKey is empty, the existing key is preserved (leave-blank-keep-current
// contract, matching EmailService.UpsertEmailConfig). Returns the updated
// config (including the API key in plaintext, same as GetConfig).
func (s *WHOISService) UpsertConfig(userID int64, req *models.UpdateWHOISConfigRequest) (*models.WHOISConfig, error) {
	now := time.Now()

	// Check if config exists
	var existingID int64
	err := database.DB.QueryRow(`SELECT id FROM whois_config WHERE user_id = ?`, userID).Scan(&existingID)

	switch err {
	case sql.ErrNoRows:
		// Insert new config — an initial config must include an API key.
		// An empty key would otherwise create a permanently unusable config.
		if strings.TrimSpace(req.APIKey) == "" {
			return nil, ErrWHOISAPIKeyRequired
		}
		_, err = database.DB.Exec(
			`INSERT INTO whois_config (user_id, api_key, created_at, updated_at)
			 VALUES (?, ?, ?, ?)`,
			userID, req.APIKey, now, now,
		)
	case nil:
		// Update existing config
		if req.APIKey != "" {
			// Update with new key
			_, err = database.DB.Exec(
				`UPDATE whois_config SET api_key = ?, updated_at = ? WHERE user_id = ?`,
				req.APIKey, now, userID,
			)
		} else {
			// No changes requested — simply touch updated_at
			_, err = database.DB.Exec(
				`UPDATE whois_config SET updated_at = ? WHERE user_id = ?`,
				now, userID,
			)
		}
	default:
		// Any other SELECT error (connection, context, ...) — surface it
		// rather than falling through to a misleading INSERT failure.
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	// Return sanitized view
	return s.GetConfig(userID)
}

// Query performs a WHOIS lookup for the given domain using the user's
// configured API key. Returns ErrWHOISAPIKeyNotConfigured when the user has not
// configured a key.
func (s *WHOISService) Query(userID int64, domain string) (*models.WHOISLookupResult, error) {
	cfg, err := s.getConfigWithKey(userID)
	if err != nil {
		return nil, err
	}
	if cfg == nil || strings.TrimSpace(cfg.APIKey) == "" {
		return nil, ErrWHOISAPIKeyNotConfigured
	}

	// Build request to upstream WhoisJSON.com
	reqURL := fmt.Sprintf("%s?domain=%s", whoisJSONEndpoint, url.QueryEscape(domain))
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	// WhoisJSON.com uses a non-standard Authorization scheme: "TOKEN=KEY"
	req.Header.Set("Authorization", "TOKEN="+cfg.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB cap
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		// Upstream error (e.g. 403 invalid key, 429 rate limit). Surface message.
		var errBody struct {
			StatusCode int    `json:"statusCode"`
			Message    string `json:"message"`
		}
		if jsonErr := json.Unmarshal(body, &errBody); jsonErr == nil && errBody.Message != "" {
			return nil, fmt.Errorf("whoisjson.com returned status %d: %s", resp.StatusCode, errBody.Message)
		}
		return nil, fmt.Errorf("whoisjson.com returned status %d", resp.StatusCode)
	}

	// Parse upstream payload. WhoisJSON.com is inconsistent across TLDs:
	// status/nameserver may be string or []string, dnssec may be string or
	// bool, registered may be bool or string, etc. To stay robust, decode the
	// body into a generic map and pull fields with type-tolerant helpers.
	var rawMap map[string]any
	if err := json.Unmarshal(body, &rawMap); err != nil {
		return nil, fmt.Errorf("failed to parse whoisjson.com response: %w", err)
	}

	// Preserve raw for debugging (already parsed above; reuse it).
	result := &models.WHOISLookupResult{
		Domain:      domain,
		Registered:  toBool(rawMap["registered"]),
		Registrar:   toRegistrar(rawMap["registrar"]),
		Created:     toString(rawMap["created"]),
		Changed:     toString(rawMap["changed"]),
		Expires:     toString(rawMap["expires"]),
		Status:      toStringSlice(rawMap["status"]),
		Nameservers: toStringSlice(rawMap["nameserver"]),
		IPs:         toStringSlice(rawMap["ips"]),
		DNSSEC:      toString(rawMap["dnssec"]),
		WhoisServer: toString(rawMap["whoisserver"]),
		Contacts:    toContacts(rawMap["contacts"]),
		Raw:         rawMap,
	}

	// Message contract:
	//   - !registered      → "Domain may be unregistered or privacy-protected"
	//   - registered but key fields (created/expires/registrar) all empty →
	//     privacy-protected response that hides ownership/registration data
	//   - otherwise         → no message
	if !result.Registered {
		result.Message = "Domain may be unregistered or privacy-protected"
	} else if result.Created == "" && result.Expires == "" && result.Registrar == nil {
		result.Message = "WHOIS data for this domain is privacy-protected"
	}

	return result, nil
}

// toString coerces a JSON value to a string. Handles string, bool, float64,
// nil and falls back to fmt for anything else. Empty result is "".
func toString(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	case float64:
		// JSON numbers come back as float64 via map[string]any.
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", x)
	}
}

// toBool coerces a JSON value to bool. WhoisJSON.com sometimes returns
// `registered` as a string "true"/"false" instead of a bool.
func toBool(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		b, _ := strconv.ParseBool(x)
		return b
	case float64:
		return x != 0
	default:
		return false
	}
}

// toStringSlice normalizes a JSON value that may be a string, an array of
// strings (or of any scalars), null, or missing into a []string.
func toStringSlice(v any) []string {
	switch x := v.(type) {
	case nil:
		return nil
	case string:
		if x == "" {
			return nil
		}
		return []string{x}
	case []any:
		out := make([]string, 0, len(x))
		for _, item := range x {
			if s := toString(item); s != "" {
				out = append(out, s)
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	default:
		s := toString(x)
		if s == "" {
			return nil
		}
		return []string{s}
	}
}

// toRegistrar builds a WHOISRegistrar from a generic map. Returns nil when the
// map is absent or all fields are empty, so the frontend doesn't receive a
// misleading all-empty registrar object.
func toRegistrar(v any) *models.WHOISRegistrar {
	m, ok := v.(map[string]any)
	if !ok || m == nil {
		return nil
	}
	r := &models.WHOISRegistrar{
		ID:    toString(m["id"]),
		Name:  toString(m["name"]),
		Email: toString(m["email"]),
		URL:   toString(m["url"]),
		Phone: toString(m["phone"]),
	}
	if r.ID == "" && r.Name == "" && r.Email == "" && r.URL == "" && r.Phone == "" {
		return nil
	}
	return r
}

// toContacts builds a WHOISContacts from a generic map.
func toContacts(v any) *models.WHOISContacts {
	m, ok := v.(map[string]any)
	if !ok || m == nil {
		return nil
	}
	c := &models.WHOISContacts{
		Owner: toContactList(m["owner"]),
		Admin: toContactList(m["admin"]),
		Tech:  toContactList(m["tech"]),
	}
	if c.Owner == nil && c.Admin == nil && c.Tech == nil {
		return nil
	}
	return c
}

// toContactList normalizes one of owner/admin/tech which may be a single
// object or an array of objects.
func toContactList(v any) []models.WHOISContact {
	switch x := v.(type) {
	case nil:
		return nil
	case []any:
		out := make([]models.WHOISContact, 0, len(x))
		for _, item := range x {
			if m, ok := item.(map[string]any); ok {
				if c := toContact(m); c != nil {
					out = append(out, *c)
				}
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	case map[string]any:
		if c := toContact(x); c != nil {
			return []models.WHOISContact{*c}
		}
		return nil
	default:
		return nil
	}
}

func toContact(m map[string]any) *models.WHOISContact {
	if m == nil {
		return nil
	}
	c := &models.WHOISContact{
		Handle:       toString(m["handle"]),
		Name:         toString(m["name"]),
		Email:        toString(m["email"]),
		Organization: toString(m["organization"]),
		Country:      toString(m["country"]),
	}
	if c.Handle == "" && c.Name == "" && c.Email == "" && c.Organization == "" && c.Country == "" {
		return nil
	}
	return c
}
