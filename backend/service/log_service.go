package service

import (
	"dns-mng/database"
	"dns-mng/models"
	"encoding/json"
)

type LogService struct{}

func NewLogService() *LogService {
	return &LogService{}
}

// CreateLog creates a new operation log
func (s *LogService) CreateLog(userID int64, action, resource, resourceID string, details interface{}, ipAddress string) error {
	detailsJSON, _ := json.Marshal(details)
	
	_, err := database.DB.Exec(
		`INSERT INTO operation_logs (user_id, action, resource, resource_id, details, ip_address) 
		 VALUES (?, ?, ?, ?, ?, ?)`,
		userID, action, resource, resourceID, string(detailsJSON), ipAddress,
	)
	return err
}

// GetLogs retrieves operation logs for a user
func (s *LogService) GetLogs(userID int64, limit int) ([]models.OperationLog, error) {
	if limit <= 0 {
		limit = 50
	}
	
	rows, err := database.DB.Query(
		`SELECT l.id, l.user_id, u.username, l.action, l.resource, l.resource_id, 
		        l.details, l.ip_address, l.created_at
		 FROM operation_logs l
		 JOIN users u ON l.user_id = u.id
		 WHERE l.user_id = ?
		 ORDER BY l.created_at DESC
		 LIMIT ?`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.OperationLog
	for rows.Next() {
		var log models.OperationLog
		err := rows.Scan(
			&log.ID, &log.UserID, &log.Username, &log.Action, &log.Resource,
			&log.ResourceID, &log.Details, &log.IPAddress, &log.CreatedAt,
		)
		if err != nil {
			continue
		}
		logs = append(logs, log)
	}
	return logs, nil
}
