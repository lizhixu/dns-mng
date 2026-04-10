package service

import (
	"dns-mng/database"
	"dns-mng/models"
	"fmt"
	"time"
)

type DomainCacheService struct{}

func NewDomainCacheService() *DomainCacheService {
	return &DomainCacheService{}
}

// GetCacheByUser gets all domain cache entries for a user
func (s *DomainCacheService) GetCacheByUser(userID int64) ([]models.DomainCache, error) {
	rows, err := database.DB.Query(
		`SELECT id, user_id, account_id, domain_id, domain_name, renewal_date, renewal_url, created_at, updated_at
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
		if err := rows.Scan(&c.ID, &c.UserID, &c.AccountID, &c.DomainID, &c.DomainName,
			&c.RenewalDate, &c.RenewalURL, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		caches = append(caches, c)
	}
	return caches, nil
}

// GetCache gets a single domain cache entry
func (s *DomainCacheService) GetCache(userID, accountID int64, domainID string) (*models.DomainCache, error) {
	var c models.DomainCache
	err := database.DB.QueryRow(
		`SELECT id, user_id, account_id, domain_id, domain_name, renewal_date, renewal_url, created_at, updated_at
		 FROM domain_cache WHERE user_id = ? AND account_id = ? AND domain_id = ?`,
		userID, accountID, domainID,
	).Scan(&c.ID, &c.UserID, &c.AccountID, &c.DomainID, &c.DomainName,
		&c.RenewalDate, &c.RenewalURL, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// UpsertCache creates or updates a domain cache entry
func (s *DomainCacheService) UpsertCache(userID, accountID int64, domainID, domainName string, req *models.UpdateDomainCacheRequest) (*models.DomainCache, error) {
	now := time.Now()

	// Try to update first
	result, err := database.DB.Exec(
		`UPDATE domain_cache SET renewal_date = ?, renewal_url = ?, domain_name = ?, updated_at = ?
		 WHERE user_id = ? AND account_id = ? AND domain_id = ?`,
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

// DeleteCache deletes a domain cache entry
func (s *DomainCacheService) DeleteCache(userID, accountID int64, domainID string) error {
	_, err := database.DB.Exec(
		`DELETE FROM domain_cache WHERE user_id = ? AND account_id = ? AND domain_id = ?`,
		userID, accountID, domainID,
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

// BatchUpsertCache creates or updates multiple domain cache entries
func (s *DomainCacheService) BatchUpsertCache(userID int64, items []models.BatchCacheItem) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	for _, item := range items {
		// Try to update first
		result, err := tx.Exec(
			`UPDATE domain_cache SET renewal_date = ?, renewal_url = ?, domain_name = ?, updated_at = ?
			 WHERE user_id = ? AND account_id = ? AND domain_id = ?`,
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
	}

	return tx.Commit()
}

// GetCacheStats returns statistics about cached domains
func (s *DomainCacheService) GetCacheStats(userID int64) (*models.CacheStats, error) {
	stats := &models.CacheStats{}

	// Total cached domains
	err := database.DB.QueryRow(
		`SELECT COUNT(*) FROM domain_cache WHERE user_id = ?`,
		userID,
	).Scan(&stats.TotalCached)
	if err != nil {
		return nil, err
	}

	// Domains with renewal dates
	err = database.DB.QueryRow(
		`SELECT COUNT(*) FROM domain_cache WHERE user_id = ? AND renewal_date != '' AND renewal_date != 'permanent'`,
		userID,
	).Scan(&stats.WithRenewalDate)
	if err != nil {
		return nil, err
	}

	// Permanent free domains
	err = database.DB.QueryRow(
		`SELECT COUNT(*) FROM domain_cache WHERE user_id = ? AND renewal_date = 'permanent'`,
		userID,
	).Scan(&stats.PermanentFree)
	if err != nil {
		return nil, err
	}

	// Domains with renewal URLs
	err = database.DB.QueryRow(
		`SELECT COUNT(*) FROM domain_cache WHERE user_id = ? AND renewal_url != ''`,
		userID,
	).Scan(&stats.WithRenewalURL)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// BatchDeleteCache deletes multiple domain cache entries
func (s *DomainCacheService) BatchDeleteCache(userID int64, items []models.BatchCacheDeleteItem) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, item := range items {
		_, err = tx.Exec(
			`DELETE FROM domain_cache WHERE user_id = ? AND account_id = ? AND domain_id = ?`,
			userID, item.AccountID, item.DomainID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}



