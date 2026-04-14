package ipv64

// GetDomainsResponse represents the response from get_domains API
type GetDomainsResponse struct {
	Subdomains map[string]DomainInfo `json:"subdomains"`
	Info       string                `json:"info"`
	Status     string                `json:"status"`
	AddDomain  string                `json:"add_domain"`
}

// DomainInfo represents domain information
type DomainInfo struct {
	Updates          int      `json:"updates"`
	Wildcard         int      `json:"wildcard"`
	DomainUpdateHash string   `json:"domain_update_hash"`
	IPv6Prefix       string   `json:"ipv6prefix"`
	DualStack        string   `json:"dualstack"`
	Deactivated      int      `json:"deactivated"`
	Records          []Record `json:"records"`
}

// Record represents a DNS record
type Record struct {
	RecordID       int    `json:"record_id"`
	Content        string `json:"content"`
	TTL            int    `json:"ttl"`
	Type           string `json:"type"`
	Praefix        string `json:"praefix"`
	LastUpdate     string `json:"last_update"`
	RecordKey      string `json:"record_key"`
	Deactivated    int    `json:"deactivated"`
	FailoverPolicy string `json:"failover_policy"`
}

// APIResponse represents a standard API response
// According to API docs: Response, Status, API-Call fields
type APIResponse struct {
	Response  string `json:"response"`   // Response to your API call
	Info      string `json:"info"`       // Info message (e.g., "success")
	Status    string `json:"status"`     // HTTP status code as string (e.g., "200 OK")
	APICall   string `json:"api_call"`   // Which call was called
	DelRecord string `json:"del_record"` // For delete record responses
	AddRecord string `json:"add_record"` // For add record responses
	AddDomain string `json:"add_domain"` // For add domain responses
	DelDomain string `json:"del_domain"` // For delete domain responses
}
