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

// APICallLog represents a complete API call record
type APICallLog struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	Username       string    `json:"username"`
	Method         string    `json:"method"`          // HTTP method: GET, POST, PUT, DELETE, etc.
	Path           string    `json:"path"`            // Request path
	Query          string    `json:"query"`           // Query parameters (complete)
	RequestHeaders string    `json:"request_headers"` // Request headers (complete JSON)
	RequestBody    string    `json:"request_body"`    // Request body (complete)
	StatusCode     int       `json:"status_code"`     // Response status code
	ResponseBody   string    `json:"response_body"`   // Response body (complete)
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	DurationMs     int       `json:"duration_ms"` // Request duration in milliseconds
	ErrorMessage   string    `json:"error_message,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// APICallLogListResponse represents a paginated list of API call logs
type APICallLogListResponse struct {
	Logs       []APICallLog `json:"logs"`
	Total      int          `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalPages int          `json:"total_pages"`
}

// SchedulerLogListResponse represents a paginated list of scheduler logs
type SchedulerLogListResponse struct {
	Logs       []SchedulerLog `json:"logs"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// LogsResponse combines API call logs and scheduler logs
type LogsResponse struct {
	APICallLogs    []APICallLog   `json:"api_call_logs"`
	SchedulerLogs  []SchedulerLog `json:"scheduler_logs"`
	TotalAPICalls  int            `json:"total_api_calls"`
	TotalScheduler int            `json:"total_scheduler"`
}
