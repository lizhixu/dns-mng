package database

import (
	"log"
)

// MigrateOperationLogsToAPILogs migrates old operation_logs to new api_call_logs format
// This is a one-time migration helper
func MigrateOperationLogsToAPILogs() error {
	// Check if operation_logs table exists
	var tableName string
	err := DB.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='operation_logs'`).Scan(&tableName)
	if err != nil {
		// Table doesn't exist, no migration needed
		log.Println("No operation_logs table found, skipping migration")
		return nil
	}

	// Check if we've already migrated (check if api_call_logs has data)
	var count int
	err = DB.QueryRow(`SELECT COUNT(*) FROM api_call_logs`).Scan(&count)
	if err == nil && count > 0 {
		log.Println("API call logs already exist, skipping migration")
		return nil
	}

	log.Println("Starting migration from operation_logs to api_call_logs...")

	// Note: This is a simplified migration since old logs don't have complete API call info
	// We'll create basic records from the old format
	_, err = DB.Exec(`
		INSERT INTO api_call_logs (user_id, method, path, query, request_body, status_code, 
		                           response_body, ip_address, duration_ms, created_at)
		SELECT 
			user_id,
			'LEGACY' as method,
			'/' || resource || '/' || COALESCE(resource_id, '') as path,
			'' as query,
			COALESCE(details, '{}') as request_body,
			200 as status_code,
			'{"action":"' || action || '"}' as response_body,
			COALESCE(ip_address, '') as ip_address,
			0 as duration_ms,
			created_at
		FROM operation_logs
	`)

	if err != nil {
		log.Printf("Migration failed: %v", err)
		return err
	}

	log.Println("Migration completed successfully")
	log.Println("Note: Old operation_logs table is preserved. You can drop it manually if needed.")
	
	return nil
}
