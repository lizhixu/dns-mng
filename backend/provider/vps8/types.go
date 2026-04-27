package vps8

// VPS8 API response types

// DomainListResponse represents the response from domain_list endpoint
type DomainListResponse struct {
	Result []Domain `json:"result"`
	Error  *string  `json:"error"`
}

// Domain represents a domain in VPS8
type Domain struct {
	Domain        string `json:"domain"`
	PlatformType  string `json:"platform_type"`
	SourceService string `json:"source_service"`
	CreatedAt     string `json:"created_at"`
	ExpiresAt     string `json:"expires_at"`
}

// RecordListResponse represents the response from record_list endpoint
type RecordListResponse struct {
	Result []Record `json:"result"`
	Error  *string  `json:"error"`
}

// Record represents a DNS record in VPS8
type Record struct {
	ID               int    `json:"id"`
	Name             string `json:"host"`
	Type             string `json:"type"`
	Content          string `json:"value"`
	TTL              int    `json:"ttl"`
	Priority         int    `json:"priority"`
	ProviderRecordID string `json:"provider_record_id"`
}

// CreateRecordRequest represents the request for record_create endpoint
type CreateRecordRequest struct {
	Domain  string `json:"domain"`
	Name    string `json:"host"`
	Type    string `json:"type"`
	Content string `json:"value"`
	TTL     int    `json:"ttl"`
}

// UpdateRecordRequest represents the request for record_update endpoint
type UpdateRecordRequest struct {
	Domain  string `json:"domain"`
	ID      int    `json:"id"`
	Content string `json:"value"`
	TTL     int    `json:"ttl"`
}

// DeleteRecordRequest represents the request for record_delete endpoint
type DeleteRecordRequest struct {
	Domain string `json:"domain"`
	ID     int    `json:"id"`
}
