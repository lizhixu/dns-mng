package models

import "time"

// SchedulerLog represents a scheduler task log entry
type SchedulerLog struct {
	ID          int64      `json:"id"`
	TaskName    string     `json:"task_name"`
	Status      string     `json:"status"` // success, error, running
	Message     string     `json:"message,omitempty"`
	Details     string     `json:"details,omitempty"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	DurationMs  int        `json:"duration_ms,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// OperationLog represents an operation log entry
type OperationLog struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Username   string    `json:"username"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resource_id,omitempty"`
	Details    string    `json:"details,omitempty"`
	IPAddress  string    `json:"ip_address,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// LogListResponse represents a paginated list of operation logs
type LogListResponse struct {
	Logs       []OperationLog `json:"logs"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// SchedulerLogListResponse represents a paginated list of scheduler logs
type SchedulerLogListResponse struct {
	Logs       []SchedulerLog `json:"logs"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// LogsResponse combines operation logs and scheduler logs
type LogsResponse struct {
	OperationLogs  []OperationLog `json:"operation_logs"`
	SchedulerLogs  []SchedulerLog `json:"scheduler_logs"`
	TotalOps       int            `json:"total_ops"`
	TotalScheduler int            `json:"total_scheduler"`
}
