package models

import "time"

// CFOptimize represents a Cloudflare CDN optimization configuration
type CFOptimize struct {
	ID                int64     `json:"id"`
	UserID            int64     `json:"user_id"`
	AccountID         int64     `json:"account_id"`
	ZoneID            string    `json:"zone_id"`
	ZoneName          string    `json:"zone_name"`
	OriginIP          string    `json:"origin_ip"`
	OriginRecordName  string    `json:"origin_record_name"`
	OriginRecordID    string    `json:"origin_record_id"`
	CnameTarget       string    `json:"cname_target"`
	CnameRecordName   string    `json:"cname_record_name"`
	CnameRecordID     string    `json:"cname_record_id"`
	CustomHostname    string    `json:"custom_hostname"`
	CustomHostnameID  string    `json:"custom_hostname_id"`
	Status            string    `json:"status"`
	SSLStatus         string    `json:"ssl_status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// CreateCFOptimizeRequest is the request for one-click CDN optimization
type CreateCFOptimizeRequest struct {
	AccountID      int64  `json:"account_id" binding:"required"`
	ZoneName       string `json:"zone_name" binding:"required"`
	Hostname       string `json:"hostname" binding:"required"`       // e.g. "www"
	OriginIP       string `json:"origin_ip" binding:"required"`      // e.g. "1.2.3.4"
	CnameTarget    string `json:"cname_target"`                      // optional, default "cloudflare.468123.xyz"
}

// CFOptimizeResponse is the response for a single optimization config
type CFOptimizeResponse struct {
	CFOptimize
	AccountName string `json:"account_name,omitempty"`
}
