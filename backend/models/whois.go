package models

import "time"

// WHOISConfig represents WHOIS lookup configuration for a user.
// The API key is returned in plaintext to the owning user (same pattern as the
// DDNS token: the key belongs to the user and they may view/edit it directly).
// All endpoints are JWT-protected, so the key is only returned to its owner.
type WHOISConfig struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	APIKey    string    `json:"api_key"` // Returned in plaintext to the owning user
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateWHOISConfigRequest is the request body for updating WHOIS configuration.
// APIKey has no binding:"required", so an empty value means "keep existing key".
type UpdateWHOISConfigRequest struct {
	APIKey  string `json:"api_key"`
	Enabled bool   `json:"enabled"`
}

// WHOISLookupResult is the structured result of a WHOIS lookup, returned to the
// client as JSON. The Raw field preserves the upstream provider payload for
// debugging/advanced use.
type WHOISLookupResult struct {
	Domain      string          `json:"domain"`
	Registered  bool           `json:"registered"`
	Message     string          `json:"message,omitempty"`
	Registrar   *WHOISRegistrar `json:"registrar,omitempty"`
	Created     string          `json:"created,omitempty"`
	Changed     string          `json:"changed,omitempty"`
	Expires     string          `json:"expires,omitempty"`
	Status      []string        `json:"status,omitempty"`
	Nameservers []string        `json:"nameservers,omitempty"`
	IPs         []string        `json:"ips,omitempty"`
	DNSSEC      string          `json:"dnssec,omitempty"`
	WhoisServer string          `json:"whois_server,omitempty"`
	Contacts    *WHOISContacts  `json:"contacts,omitempty"`
	Raw         map[string]any  `json:"raw,omitempty"`
}

// WHOISRegistrar holds registrar info from the upstream WHOIS response.
type WHOISRegistrar struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
	Phone string `json:"phone,omitempty"`
}

// WHOISContacts groups owner/admin/tech contact arrays.
type WHOISContacts struct {
	Owner []WHOISContact `json:"owner,omitempty"`
	Admin []WHOISContact `json:"admin,omitempty"`
	Tech  []WHOISContact `json:"tech,omitempty"`
}

// WHOISContact holds a single contact record (typically redacted for privacy).
type WHOISContact struct {
	Handle       string `json:"handle,omitempty"`
	Name         string `json:"name,omitempty"`
	Email        string `json:"email,omitempty"`
	Organization string `json:"organization,omitempty"`
	Country      string `json:"country,omitempty"`
}
