package service

import (
	"dns-mng/database"
	"dns-mng/models"
)

type LogService struct{}

func NewLogService() *LogService {
	return &LogService{}
}

// CreateAPICallLog creates a new API call log with complete information
func (s *LogService) CreateAPICallLog(log *models.APICallLog) error {
	_, err := database.DB.Exec(
		`INSERT INTO api_call_logs (user_id, method, path, query, request_headers, request_body, 
		 status_code, response_body, ip_address, user_agent, duration_ms, error_message) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		log.UserID, log.Method, log.Path, log.Query, log.RequestHeaders, log.RequestBody,
		log.StatusCode, log.ResponseBody, log.IPAddress, log.UserAgent, log.DurationMs, log.ErrorMessage,
	)
	return err
}

// GetAPICallLogs retrieves API call logs for a user with pagination
func (s *LogService) GetAPICallLogs(userID int64, page, pageSize int) (*models.APICallLogListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Get total count with timeout
	var total int
	err := database.DB.QueryRow(
		`SELECT COUNT(*) FROM api_call_logs WHERE user_id = ?`,
		userID,
	).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Query with optimized fields - use SUBSTR to limit large text fields
	rows, err := database.DB.Query(
		`SELECT l.id, l.user_id, COALESCE(u.username, '') as username, l.method, l.path, 
		        COALESCE(CASE WHEN LENGTH(l.query) > 500 THEN SUBSTR(l.query, 1, 500) || '...' ELSE l.query END, '') as query,
		        COALESCE(CASE WHEN LENGTH(l.request_headers) > 2000 THEN SUBSTR(l.request_headers, 1, 2000) || '...' ELSE l.request_headers END, '') as request_headers,
		        COALESCE(CASE WHEN LENGTH(l.request_body) > 5000 THEN SUBSTR(l.request_body, 1, 5000) || '...' ELSE l.request_body END, '') as request_body,
		        l.status_code, 
		        COALESCE(CASE WHEN LENGTH(l.response_body) > 5000 THEN SUBSTR(l.response_body, 1, 5000) || '...' ELSE l.response_body END, '') as response_body,
		        COALESCE(l.ip_address, '') as ip_address, 
		        COALESCE(CASE WHEN LENGTH(l.user_agent) > 500 THEN SUBSTR(l.user_agent, 1, 500) || '...' ELSE l.user_agent END, '') as user_agent,
		        COALESCE(l.duration_ms, 0) as duration_ms, COALESCE(l.error_message, '') as error_message, l.created_at
		 FROM api_call_logs l
		 LEFT JOIN users u ON l.user_id = u.id
		 WHERE l.user_id = ?
		 ORDER BY l.created_at DESC
		 LIMIT ? OFFSET ?`,
		userID, pageSize, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.APICallLog
	for rows.Next() {
		var log models.APICallLog
		err := rows.Scan(
			&log.ID, &log.UserID, &log.Username, &log.Method, &log.Path, &log.Query,
			&log.RequestHeaders, &log.RequestBody, &log.StatusCode, &log.ResponseBody,
			&log.IPAddress, &log.UserAgent, &log.DurationMs, &log.ErrorMessage, &log.CreatedAt,
		)
		if err != nil {
			continue
		}
		logs = append(logs, log)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &models.APICallLogListResponse{
		Logs:       logs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + pageSize - 1) / pageSize,
	}, nil
}
