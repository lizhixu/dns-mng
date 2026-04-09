package dnshe

// DNSHE API types

// Subdomain represents a subdomain in DNSHE
type Subdomain struct {
	ID          int    `json:"id"`
	Subdomain   string `json:"subdomain"`
	RootDomain  string `json:"rootdomain"`
	FullDomain  string `json:"full_domain"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	ExpiresAt   string `json:"expires_at,omitempty"`
}

// SubdomainsResponse represents the response from list subdomains
type SubdomainsResponse struct {
	Success    bool        `json:"success"`
	Count      int         `json:"count"`
	Subdomains []Subdomain `json:"subdomains"`
}

// SubdomainDetailResponse represents the response from get subdomain
type SubdomainDetailResponse struct {
	Success    bool         `json:"success"`
	Subdomain  Subdomain    `json:"subdomain"`
	DNSRecords []DNSRecord  `json:"dns_records"`
	DNSCount   int          `json:"dns_count"`
}

// DNSRecord represents a DNS record in DNSHE
type DNSRecord struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Content   string  `json:"content"`
	TTL       int     `json:"ttl"`
	Priority  *int    `json:"priority"`
	Proxied   bool    `json:"proxied"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

// DNSRecordsResponse represents the response from list DNS records
type DNSRecordsResponse struct {
	Success bool        `json:"success"`
	Count   int         `json:"count"`
	Records []DNSRecord `json:"records"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
