package models

// Record represents a DNS record from any provider
type Record struct {
	ID         string `json:"id"`
	DomainID   string `json:"domain_id"`
	DomainName string `json:"domain_name,omitempty"`
	NodeName   string `json:"node_name"`
	RecordType string `json:"record_type"`
	TTL        int    `json:"ttl"`
	State      bool   `json:"state"`
	Content    string `json:"content"`
	Priority   int    `json:"priority,omitempty"`
	UpdatedOn  string `json:"updated_on,omitempty"`
	// Raw stores provider-specific extra fields
	Raw map[string]interface{} `json:"raw,omitempty"`
}

type CreateRecordRequest struct {
	NodeName   string `json:"node_name" binding:"required"`
	RecordType string `json:"record_type" binding:"required"`
	TTL        int    `json:"ttl"`
	State      *bool  `json:"state"`
	Content    string `json:"content" binding:"required"`
	Priority   int    `json:"priority,omitempty"`
}

type UpdateRecordRequest struct {
	NodeName   string `json:"node_name"`
	RecordType string `json:"record_type"`
	TTL        int    `json:"ttl"`
	State      *bool  `json:"state"`
	Content    string `json:"content"`
	Priority   int    `json:"priority,omitempty"`
}
