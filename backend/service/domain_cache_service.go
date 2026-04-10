package service

import (
	"database/sql"
	"dns-mng/database"
	"dns-mng/models"
	"fmt"
	"time"
)

type DomainCacheService struct{}

func NewDomainCacheService() *DomainCacheService {
	return &DomainCacheService{}
}

// GetCacheByUser gets all domain cache entries for a user (excluding soft deleted)
func (s *DomainCacheService) GetCacheByUser(userID int64) ([]models.DomainCache, error) {
	rows, err := database.DB.Query(
		`SELECT id, user_id, account_id, domain_id, domain_name, renewal_date, renewal_url, deleted_at, created_at, updated_at
		 FROM domain_cache WHERE user_id = ? AND deleted_at IS NULL ORDER BY domain_name`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var caches []models.DomainCache
	for rows.Next() {
		var c models.DomainCache
		var deletedAt sql.NullTime
		if err := rows.Scan(&c.ID, &c.UserID, &c.AccountID, &c.DomainID, &c.DomainName,
			&c.RenewalDate, &c.RenewalURL, &deletedAt, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			c.DeletedAt = &deletedAt.Time
		}
		caches = append(caches, c)
	}
	return caches, nil
}

// GetCache gets a single domain cache entry (excluding soft deleted)
func (s *DomainCacheService) GetCache(userID, accountID int64, domainID string) (*models.DomainCache, error) {
	var c models.DomainCache
	var deletedAt sql.NullTime
	err := database.DB.QueryRow(
		`SELECT id, user_id, account_id, domain_id, domain_name, renewal_date, renewal_url, deleted_at, created_at, updated_at
		 FROM domain_cache WHERE user_id = ? AND account_id = ? AND domain_id = ? AND deleted_at IS NULL`,
		userID, accountID, domainID,
	).Scan(&c.ID, &c.UserID, &c.AccountID, &c.DomainID, &c.DomainName,
		&c.RenewalDate, &c.RenewalURL, &deletedAt, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if deletedAt.Valid {
		c.DeletedAt = &deletedAt.Time
	}
	return &c, nil
}

// UpsertCache creates or updates a domain cache entry, activates if soft deleted
func (s *DomainCacheService) UpsertCache(userID, accountID int64, domainID, domainName string, req *models.UpdateDomainCacheRequest) (*models.DomainCache, error) {
	now := time.Now()

	// First check if there's a soft deleted record
	var existingID int64
	var isDeleted bool
	err := database.DB.QueryRow(
		`SELECT id, CASE WHEN deleted_at IS NOT NULL THEN 1 ELSE 0 END FROM domain_cache 
		 WHERE user_id = ? AND account_id = ? AND domain_id = ?`,
		userID, accountID, domainID,
	).Scan(&existingID, &isDeleted)

	if err == nil && isDeleted {
		// Activate the soft deleted record
		_, err = database.DB.Exec(
			`UPDATE domain_cache SET deleted_at = NULL, renewal_date = ?, renewal_url = ?, domain_name = ?, updated_at = ?
			 WHERE id = ?`,
			req.RenewalDate, req.RenewalURL, domainName, now, existingID,
		)
		if err != nil {
			return nil, err
		}
	} else if err == nil {
		// Record exists and is not deleted, update it
		result, err := database.DB.Exec(
			`UPDATE domain_cache SET renewal_date = ?, renewal_url = ?, domain_name = ?, updated_at = ?
			 WHERE user_id = ? AND account_id = ? AND domain_id = ? AND deleted_at IS NULL`,
			req.RenewalDate, req.RenewalURL, domainName, now,
			userID, accountID, domainID,
		)
		if err != nil {
			return nil, err
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			// Insert new entry
			_, err = database.DB.Exec(
				`INSERT INTO domain_cache (user_id, account_id, domain_id, domain_name, renewal_date, renewal_url, created_at, updated_at)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				userID, accountID, domainID, domainName, req.RenewalDate, req.RenewalURL, now, now,
			)
			if err != nil {
				return nil, err
			}
		}
	} else {
		// Record doesn't exist, insert new
		_, err = database.DB.Exec(
			`INSERT INTO domain_cache (user_id, account_id, domain_id, domain_name, renewal_date, renewal_url, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			userID, accountID, domainID, domainName, req.RenewalDate, req.RenewalURL, now, now,
		)
		if err != nil {
			return nil, err
		}
	}

	// Update notification settings if provided
	if req.NotifyDaysBefore > 0 {
		notificationService := NewNotificationService()
		notificationService.UpsertNotificationSetting(userID, accountID, domainID, &models.UpdateNotificationSettingRequest{
			DaysBefore: req.NotifyDaysBefore,
			Enabled:    req.NotifyEnabled,
		})
	}

	return s.GetCache(userID, accountID, domainID)
}

// DeleteCache soft deletes a domain cache entry
func (s *DomainCacheService) DeleteCache(userID, accountID int64, domainID string) error {
	now := time.Now()
	_, err := database.DB.Exec(
		`UPDATE domain_cache SET deleted_at = ?, updated_at = ? 
		 WHERE user_id = ? AND account_id = ? AND domain_id = ? AND deleted_at IS NULL`,
		now, now, userID, accountID, domainID,
	)
	return err
}

// BatchGetCacheByUser gets cache entries as a map for efficient lookup
func (s *DomainCacheService) BatchGetCacheByUser(userID int64) (map[string]*models.DomainCache, error) {
	caches, err := s.GetCacheByUser(userID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*models.DomainCache, len(caches))
	for i := range caches {
		key := fmt.Sprintf("%d:%s", caches[i].AccountID, caches[i].DomainID)
		result[key] = &caches[i]
	}
	return result, nil
}

// BatchUpsertCache creates or updates multiple domain cache entries, activates soft deleted records
func (s *DomainCacheService) BatchUpsertCache(userID int64, items []models.BatchCacheItem) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	for _, item := range items {
		// Check if record exists and if it's soft deleted
		var existingID int64
		var isDeleted bool
		err := tx.QueryRow(
			`SELECT id, CASE WHEN deleted_at IS NOT NULL THEN 1 ELSE 0 END FROM domain_cache 
			 WHERE user_id = ? AND account_id = ? AND domain_id = ?`,
			userID, item.AccountID, item.DomainID,
		).Scan(&existingID, &isDeleted)

		if err == nil && isDeleted {
			// Activate the soft deleted record
			_, err = tx.Exec(
				`UPDATE domain_cache SET deleted_at = NULL, renewal_date = ?, renewal_url = ?, domain_name = ?, updated_at = ?
				 WHERE id = ?`,
				item.RenewalDate, item.RenewalURL, item.DomainName, now, existingID,
			)
			if err != nil {
				return err
			}
		} else if err == nil {
			// Record exists and is not deleted, update it
			result, err := tx.Exec(
				`UPDATE domain_cache SET renewal_date = ?, renewal_url = ?, domain_name = ?, updated_at = ?
				 WHERE user_id = ? AND account_id = ? AND domain_id = ? AND deleted_at IS NULL`,
				item.RenewalDate, item.RenewalURL, item.DomainName, now,
				userID, item.AccountID, item.DomainID,
			)
			if err != nil {
				return err
			}

			rowsAffected, _ := result.RowsAffected()
			if rowsAffected == 0 {
				// Insert new entry
				_, err = tx.Exec(
					`INSERT INTO domain_cache (user_id, account_id, domain_id, domain_name, renewal_date, renewal_url, created_at, updated_at)
					 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
					userID, item.AccountID, item.DomainID, item.DomainName, item.RenewalDate, item.RenewalURL, now, now,
				)
				if err != nil {
					return err
				}
			}
		} else {
			// Record doesn't exist, insert new
			_, err = tx.Exec(
				`INSERT INTO domain_cache (user_id, account_id, domain_id, domain_name, renewal_date, renewal_url, created_at, updated_at)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				userID, item.AccountID, item.DomainID, item.DomainName, item.RenewalDate, item.RenewalURL, now, now,
			)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// GetCacheStats returns statistics about cached domains (excluding soft deleted)
func (s *DomainCacheService) GetCacheStats(userID int64) (*models.CacheStats, error) {
	stats := &models.CacheStats{}

	// Total cached domains (excluding soft deleted)
	err := database.DB.QueryRow(
		`SELECT COUNT(*) FROM domain_cache WHERE user_id = ? AND deleted_at IS NULL`,
		userID,
	).Scan(&stats.TotalCached)
	if err != nil {
		return nil, err
	}

	// Domains with renewal dates (excluding soft deleted)
	err = database.DB.QueryRow(
		`SELECT COUNT(*) FROM domain_cache WHERE user_id = ? AND deleted_at IS NULL AND renewal_date != '' AND renewal_date != 'permanent'`,
		userID,
	).Scan(&stats.WithRenewalDate)
	if err != nil {
		return nil, err
	}

	// Permanent free domains (excluding soft deleted)
	err = database.DB.QueryRow(
		`SELECT COUNT(*) FROM domain_cache WHERE user_id = ? AND deleted_at IS NULL AND renewal_date = 'permanent'`,
		userID,
	).Scan(&stats.PermanentFree)
	if err != nil {
		return nil, err
	}

	// Domains with renewal URLs (excluding soft deleted)
	err = database.DB.QueryRow(
		`SELECT COUNT(*) FROM domain_cache WHERE user_id = ? AND deleted_at IS NULL AND renewal_url != ''`,
		userID,
	).Scan(&stats.WithRenewalURL)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// BatchDeleteCache soft deletes multiple domain cache entries
func (s *DomainCacheService) BatchDeleteCache(userID int64, items []models.BatchCacheDeleteItem) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	for _, item := range items {
		_, err = tx.Exec(
			`UPDATE domain_cache SET deleted_at = ?, updated_at = ? 
			 WHERE user_id = ? AND account_id = ? AND domain_id = ? AND deleted_at IS NULL`,
			now, now, userID, item.AccountID, item.DomainID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// BatchRestoreCache restores multiple soft deleted domain cache entries
func (s *DomainCacheService) BatchRestoreCache(userID int64, items []models.BatchCacheDeleteItem) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	for _, item := range items {
		_, err = tx.Exec(
			`UPDATE domain_cache SET deleted_at = NULL, updated_at = ? 
			 WHERE user_id = ? AND account_id = ? AND domain_id = ? AND deleted_at IS NOT NULL`,
			now, userID, item.AccountID, item.DomainID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetSoftDeletedDomains gets all soft deleted domains for a user
func (s *DomainCacheService) GetSoftDeletedDomains(userID int64) ([]models.DomainCache, error) {
	rows, err := database.DB.Query(
		`SELECT id, user_id, account_id, domain_id, domain_name, renewal_date, renewal_url, deleted_at, created_at, updated_at
		 FROM domain_cache WHERE user_id = ? AND deleted_at IS NOT NULL ORDER BY domain_name`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var caches []models.DomainCache
	for rows.Next() {
		var c models.DomainCache
		var deletedAt sql.NullTime
		if err := rows.Scan(&c.ID, &c.UserID, &c.AccountID, &c.DomainID, &c.DomainName,
			&c.RenewalDate, &c.RenewalURL, &deletedAt, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			c.DeletedAt = &deletedAt.Time
		}
		caches = append(caches, c)
	}
	return caches, nil
}

// UpdateLastSyncTime updates the last sync time for a domain
func (s *DomainCacheService) UpdateLastSyncTime(userID, accountID int64, domainID string, updatedOn *time.Time) error {
	now := time.Now()
	_, err := database.DB.Exec(
		`UPDATE domain_cache SET last_sync_at = ?, provider_updated_on = ?, updated_at = ? 
		 WHERE user_id = ? AND account_id = ? AND domain_id = ?`,
		now, updatedOn, now, userID, accountID, domainID,
	)
	return err
}

// GetAllCacheByUser gets all domain cache entries for a user (including soft deleted)
func (s *DomainCacheService) GetAllCacheByUser(userID int64) ([]models.DomainCache, error) {
	rows, err := database.DB.Query(
		`SELECT id, user_id, account_id, domain_id, domain_name, renewal_date, renewal_url, deleted_at, created_at, updated_at
		 FROM domain_cache WHERE user_id = ? ORDER BY domain_name`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var caches []models.DomainCache
	for rows.Next() {
		var c models.DomainCache
		var deletedAt sql.NullTime
		if err := rows.Scan(&c.ID, &c.UserID, &c.AccountID, &c.DomainID, &c.DomainName,
			&c.RenewalDate, &c.RenewalURL, &deletedAt, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			c.DeletedAt = &deletedAt.Time
		}
		caches = append(caches, c)
	}
	return caches, nil
}
