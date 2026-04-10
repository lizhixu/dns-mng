package models

// ACME DNS-01 request payload.
// Compatible with common webhook patterns where the ACME client provides the FQDN and TXT value.
type AcmeDNS01Request struct {
	FQDN  string `json:"fqdn" binding:"required"`
	Value string `json:"value" binding:"required"`
	TTL   int    `json:"ttl,omitempty"`
}

type AcmeDNS01Response struct {
	Status   string `json:"status"`
	Domain   string `json:"domain,omitempty"`
	NodeName string `json:"node_name,omitempty"`
}

