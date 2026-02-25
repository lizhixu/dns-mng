package models

import "time"

// OperationLog represents an operation log entry
type OperationLog struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Username   string    `json:"username,omitempty"`
	Action     string    `json:"action"`      // create, update, delete
	Resource   string    `json:"resource"`    // account, record
	ResourceID string    `json:"resource_id"` // ID of the resource
	Details    string    `json:"details"`     // JSON string with operation details
	IPAddress  string    `json:"ip_address"`
	CreatedAt  time.Time `json:"created_at"`
}
