package database

import (
	"log"
	"time"
)

// CleanupOldAPILogs deletes API call logs older than the specified number of days
func CleanupOldAPILogs(daysToKeep int) error {
	if daysToKeep <= 0 {
		daysToKeep = 30 // Default to 30 days
	}

	cutoffDate := time.Now().AddDate(0, 0, -daysToKeep)
	
	result, err := DB.Exec(
		`DELETE FROM api_call_logs WHERE created_at < ?`,
		cutoffDate,
	)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Cleaned up %d old API call logs (older than %d days)", rowsAffected, daysToKeep)
	
	return nil
}

// GetAPILogsStats returns statistics about API call logs
func GetAPILogsStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total count
	var total int64
	err := DB.QueryRow(`SELECT COUNT(*) FROM api_call_logs`).Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// Count by date
	var today, last7days, last30days int64
	
	DB.QueryRow(`SELECT COUNT(*) FROM api_call_logs WHERE created_at >= datetime('now', '-1 day')`).Scan(&today)
	DB.QueryRow(`SELECT COUNT(*) FROM api_call_logs WHERE created_at >= datetime('now', '-7 days')`).Scan(&last7days)
	DB.QueryRow(`SELECT COUNT(*) FROM api_call_logs WHERE created_at >= datetime('now', '-30 days')`).Scan(&last30days)
	
	stats["today"] = today
	stats["last_7_days"] = last7days
	stats["last_30_days"] = last30days

	// Average size
	var avgSize float64
	DB.QueryRow(`SELECT AVG(LENGTH(request_body) + LENGTH(response_body)) FROM api_call_logs`).Scan(&avgSize)
	stats["avg_size_bytes"] = avgSize

	return stats, nil
}

// VacuumDatabase optimizes the database after cleanup
func VacuumDatabase() error {
	_, err := DB.Exec(`VACUUM`)
	if err != nil {
		return err
	}
	log.Println("Database vacuumed successfully")
	return nil
}
