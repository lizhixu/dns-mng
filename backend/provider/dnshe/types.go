package dnshe

// DNSHE API types

// Subdomain represents a subdomain in DNSHE
type Subdomain struct {
	ID                int         `json:"id"`
	Subdomain         string      `json:"subdomain"`
	RootDomain        string      `json:"rootdomain"`
	FullDomain        string      `json:"full_domain"`
	Status            string      `json:"status"`
	CreatedAt         string      `json:"created_at"`
	UpdatedAt         string      `json:"updated_at"`
	ExpiresAt         string      `json:"expires_at,omitempty"`
	NeverExpires      int         `json:"never_expires"`
	CloudflareZoneID  interface{} `json:"cloudflare_zone_id,omitempty"`
	ProviderAccountID interface{} `json:"provider_account_id,omitempty"`
}

// RegisterSubdomainResponse represents the response from register subdomain
type RegisterSubdomainResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	SubdomainID int    `json:"subdomain_id"`
	FullDomain  string `json:"full_domain"`
}

// DeleteSubdomainResponse represents the response from delete subdomain
type DeleteSubdomainResponse struct {
	Success          bool   `json:"success"`
	Message          string `json:"message"`
	SubdomainID      int    `json:"subdomain_id"`
	FullDomain       string `json:"full_domain"`
	DNSRecordsDeleted int   `json:"dns_records_deleted"`
}

// RenewSubdomainResponse represents the response from renew subdomain
type RenewSubdomainResponse struct {
	Success           bool    `json:"success"`
	Message           string  `json:"message"`
	SubdomainID       int     `json:"subdomain_id"`
	Subdomain         string  `json:"subdomain"`
	PreviousExpiresAt string  `json:"previous_expires_at"`
	NewExpiresAt      string  `json:"new_expires_at"`
	RenewedAt         string  `json:"renewed_at"`
	NeverExpires      int     `json:"never_expires"`
	Status            string  `json:"status"`
	RemainingDays     int     `json:"remaining_days"`
	ChargedAmount     float64 `json:"charged_amount"`
}

// Quota represents the DNSHE account quota
type Quota struct {
	Used        int `json:"used"`
	Base        int `json:"base"`
	InviteBonus int `json:"invite_bonus"`
	Total       int `json:"total"`
	Available   int `json:"available"`
}

// QuotaResponse represents the response from quota query
type QuotaResponse struct {
	Success bool  `json:"success"`
	Quota   Quota `json:"quota"`
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
