package models

// Domain represents a DNS domain from any provider
type Domain struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	UnicodeName string `json:"unicode_name,omitempty"`
	State       string `json:"state,omitempty"`
	Group       string `json:"group,omitempty"`
	IPv4Address string `json:"ipv4_address,omitempty"`
	IPv6Address string `json:"ipv6_address,omitempty"`
	TTL         int    `json:"ttl,omitempty"`
	CreatedOn   string `json:"created_on,omitempty"`
	UpdatedOn   string `json:"updated_on,omitempty"`
	AccountID   int64  `json:"account_id,omitempty"`
	AccountName string `json:"account_name,omitempty"`
}
