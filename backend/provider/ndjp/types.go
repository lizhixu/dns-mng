package ndjp

// NDJP API response types

// Domain represents a subdomain in NDJP
type Domain struct {
	Subdomain string `json:"subdomain"`
	FQDN      string `json:"fqdn"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// DomainsResponse represents the response from /domains endpoint
type DomainsResponse struct {
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Data    []string `json:"data"` // Array of subdomain strings
}

// RecordsResponse represents the response from /domains/{subdomain}/records endpoint
type RecordsResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"` // Can be string or array
}

// Record represents a DNS record in NDJP
type Record struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl,omitempty"`
}

// RRSet represents a resource record set
type RRSet struct {
	Name     string          `json:"name"`
	Type     string          `json:"type"`
	TTL      int             `json:"ttl"`
	Records  []RRSetRecord   `json:"records"`
	Comments []interface{}   `json:"comments"`
}

// RRSetRecord represents a single record in an RRSet
type RRSetRecord struct {
	Content  string `json:"content"`
	Disabled bool   `json:"disabled"`
}

// APIResponse represents a standard NDJP API response
type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
