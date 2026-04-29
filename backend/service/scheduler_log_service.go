package service

import (
	"dns-mng/database"
	"dns-mng/models"
	"encoding/json"
	"time"
)

type SchedulerLogService struct{}

func NewSchedulerLogService() *SchedulerLogService {
	return &SchedulerLogService{}
}

// StartTask starts a new task and returns the log ID for later update
func (s *SchedulerLogService) StartTask(taskName string, details interface{}) (int64, error) {
	startedAt := time.Now()
	detailsJSON, _ := json.Marshal(details)

	result, err := database.DB.Exec(
		`INSERT INTO scheduler_logs (task_name, status, message, details, started_at) 
		 VALUES (?, ?, ?, ?, ?)`,
		taskName, "running", "Task started", string(detailsJSON), startedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	return id, err
}

// UpdateTask updates a running task with completion status
func (s *SchedulerLogService) UpdateTask(logID int64, status string, message string) error {
	completedAt := time.Now()

	// Get the started_at time first
	var startedAt time.Time
	err := database.DB.QueryRow(
		`SELECT started_at FROM scheduler_logs WHERE id = ?`,
		logID,
	).Scan(&startedAt)
	if err != nil {
		return err
	}

	durationMs := completedAt.Sub(startedAt).Milliseconds()

	_, err = database.DB.Exec(
		`UPDATE scheduler_logs 
		 SET status = ?, message = ?, completed_at = ?, duration_ms = ?
		 WHERE id = ?`,
		status, message, completedAt, durationMs, logID,
	)
	return err
}

// GetSchedulerLogs retrieves scheduler logs with pagination
func (s *SchedulerLogService) GetSchedulerLogs(page, pageSize int) (*models.SchedulerLogListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Get total count
	var total int
	err := database.DB.QueryRow(
		`SELECT COUNT(*) FROM scheduler_logs`,
	).Scan(&total)
	if err != nil {
		return nil, err
	}

	rows, err := database.DB.Query(
		`SELECT id, task_name, status, message, details, started_at, completed_at, duration_ms, created_at
		 FROM scheduler_logs
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		pageSize, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.SchedulerLog
	for rows.Next() {
		var log models.SchedulerLog
		err := rows.Scan(
			&log.ID, &log.TaskName, &log.Status, &log.Message, &log.Details,
			&log.StartedAt, &log.CompletedAt, &log.DurationMs, &log.CreatedAt,
		)
		if err != nil {
			continue
		}
		logs = append(logs, log)
	}

	return &models.SchedulerLogListResponse{
		Logs:       logs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + pageSize - 1) / pageSize,
	}, nil
}

// GetSchedulerLogsByTask retrieves scheduler logs for a specific task with pagination
func (s *SchedulerLogService) GetSchedulerLogsByTask(taskName string, page, pageSize int) (*models.SchedulerLogListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Get total count
	var total int
	err := database.DB.QueryRow(
		`SELECT COUNT(*) FROM scheduler_logs WHERE task_name = ?`,
		taskName,
	).Scan(&total)
	if err != nil {
		return nil, err
	}

	rows, err := database.DB.Query(
		`SELECT id, task_name, status, message, details, started_at, completed_at, duration_ms, created_at
		 FROM scheduler_logs
		 WHERE task_name = ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		taskName, pageSize, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.SchedulerLog
	for rows.Next() {
		var log models.SchedulerLog
		err := rows.Scan(
			&log.ID, &log.TaskName, &log.Status, &log.Message, &log.Details,
			&log.StartedAt, &log.CompletedAt, &log.DurationMs, &log.CreatedAt,
		)
		if err != nil {
			continue
		}
		logs = append(logs, log)
	}

	return &models.SchedulerLogListResponse{
		Logs:       logs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + pageSize - 1) / pageSize,
	}, nil
}
