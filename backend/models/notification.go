package models

import "time"

// NotificationSetting represents notification settings for a domain
type NotificationSetting struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	DomainID       string    `json:"domain_id"`
	AccountID      int64     `json:"account_id"`
	DaysBefore     int       `json:"days_before"`
	Enabled        bool      `json:"enabled"`
	LastNotifiedAt *time.Time `json:"last_notified_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// EmailConfig represents email configuration for a user
type EmailConfig struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	SMTPHost     string    `json:"smtp_host"`
	SMTPPort     int       `json:"smtp_port"`
	SMTPUsername string    `json:"smtp_username"`
	SMTPPassword string    `json:"smtp_password,omitempty"` // Omit in responses
	FromEmail    string    `json:"from_email"`
	FromName     string    `json:"from_name,omitempty"`
	ToEmail      string    `json:"to_email"` // Recipient email
	Language     string    `json:"language"`  // Email language: zh, en, or empty (follow system)
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UpdateNotificationSettingRequest is the request body for updating notification settings
type UpdateNotificationSettingRequest struct {
	DaysBefore int  `json:"days_before" binding:"required,min=1,max=365"`
	Enabled    bool `json:"enabled"`
}

// UpdateEmailConfigRequest is the request body for updating email config
type UpdateEmailConfigRequest struct {
	SMTPHost     string `json:"smtp_host" binding:"required"`
	SMTPPort     int    `json:"smtp_port" binding:"required,min=1,max=65535"`
	SMTPUsername string `json:"smtp_username" binding:"required"`
	SMTPPassword string `json:"smtp_password"`
	FromEmail    string `json:"from_email" binding:"required,email"`
	FromName     string `json:"from_name"`
	ToEmail      string `json:"to_email" binding:"required,email"`
	Language     string `json:"language"`
	Enabled      bool   `json:"enabled"`
}

// TestEmailRequest is the request body for testing email configuration
type TestEmailRequest struct {
	// No fields needed - will use configured to_email
}

// ExpiringDomain represents a domain that is expiring soon
type ExpiringDomain struct {
	UserID        int64
	AccountID     int64
	DomainID      string
	DomainName    string
	AccountName   string
	RenewalDate   string
	RenewalURL    string
	DaysRemaining int
	ToEmail       string // Recipient email from config
	Language      string // Email language from config
}
