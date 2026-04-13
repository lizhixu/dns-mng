package models

// DDNSToken is for user-level tokens (one token per user, system-wide)
type DDNSToken struct {
	ID         int64  `json:"id"`
	UserID     int64  `json:"user_id"`
	Token      string `json:"token"`
	Enabled    bool   `json:"enabled"`
	LastUsedAt string `json:"last_used_at,omitempty"`
	LastIP     string `json:"last_ip,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type UpdateDDNSTokenRequest struct {
	Token   string `json:"token"`
	Enabled *bool  `json:"enabled"`
}
