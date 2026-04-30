package service

import (
	"database/sql"
	"dns-mng/database"
	"dns-mng/models"
	"time"
)

type NotificationService struct{}

func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

// GetNotificationSetting gets notification setting for a domain
func (s *NotificationService) GetNotificationSetting(userID, accountID int64, domainID string) (*models.NotificationSetting, error) {
	var setting models.NotificationSetting
	var lastNotifiedAt sql.NullTime
	var enabled int

	err := database.DB.QueryRow(
		`SELECT id, user_id, domain_id, account_id, days_before, enabled, last_notified_at, created_at, updated_at
		 FROM notification_settings WHERE user_id = ? AND account_id = ? AND domain_id = ?`,
		userID, accountID, domainID,
	).Scan(&setting.ID, &setting.UserID, &setting.DomainID, &setting.AccountID,
		&setting.DaysBefore, &enabled, &lastNotifiedAt, &setting.CreatedAt, &setting.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	setting.Enabled = enabled == 1
	if lastNotifiedAt.Valid {
		setting.LastNotifiedAt = &lastNotifiedAt.Time
	}

	return &setting, nil
}

// UpsertNotificationSetting creates or updates notification setting
func (s *NotificationService) UpsertNotificationSetting(userID, accountID int64, domainID string, req *models.UpdateNotificationSettingRequest) (*models.NotificationSetting, error) {
	now := time.Now()
	enabled := 0
	if req.Enabled {
		enabled = 1
	}

	// Try to update first
	result, err := database.DB.Exec(
		`UPDATE notification_settings SET days_before = ?, enabled = ?, updated_at = ?
		 WHERE user_id = ? AND account_id = ? AND domain_id = ?`,
		req.DaysBefore, enabled, now, userID, accountID, domainID,
	)
	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Insert new entry
		_, err = database.DB.Exec(
			`INSERT INTO notification_settings (user_id, domain_id, account_id, days_before, enabled, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			userID, domainID, accountID, req.DaysBefore, enabled, now, now,
		)
		if err != nil {
			return nil, err
		}
	}

	return s.GetNotificationSetting(userID, accountID, domainID)
}

// GetAllNotificationSettings gets all notification settings for a user
func (s *NotificationService) GetAllNotificationSettings(userID int64) ([]models.NotificationSetting, error) {
	rows, err := database.DB.Query(
		`SELECT id, user_id, domain_id, account_id, days_before, enabled, last_notified_at, created_at, updated_at
		 FROM notification_settings WHERE user_id = ? ORDER BY domain_id`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []models.NotificationSetting
	for rows.Next() {
		var setting models.NotificationSetting
		var lastNotifiedAt sql.NullTime
		var enabled int

		if err := rows.Scan(&setting.ID, &setting.UserID, &setting.DomainID, &setting.AccountID,
			&setting.DaysBefore, &enabled, &lastNotifiedAt, &setting.CreatedAt, &setting.UpdatedAt); err != nil {
			return nil, err
		}

		setting.Enabled = enabled == 1
		if lastNotifiedAt.Valid {
			setting.LastNotifiedAt = &lastNotifiedAt.Time
		}
		settings = append(settings, setting)
	}

	return settings, nil
}

// UpdateLastNotifiedAt updates the last notified timestamp
func (s *NotificationService) UpdateLastNotifiedAt(userID, accountID int64, domainID string) error {
	now := time.Now()
	_, err := database.DB.Exec(
		`UPDATE notification_settings SET last_notified_at = ? WHERE user_id = ? AND account_id = ? AND domain_id = ?`,
		now, userID, accountID, domainID,
	)
	return err
}

// GetExpiringDomains gets domains that need notification (excluding soft deleted)
func (s *NotificationService) GetExpiringDomains() ([]models.ExpiringDomain, error) {
	rows, err := database.DB.Query(`
		SELECT 
			dc.user_id,
			dc.account_id,
			dc.domain_id,
			dc.domain_name,
			dc.renewal_date,
			dc.renewal_url,
			ns.days_before,
			ns.last_notified_at,
			ec.to_email,
			ec.language
		FROM domain_cache dc
		INNER JOIN notification_settings ns ON 
			dc.user_id = ns.user_id AND 
			dc.account_id = ns.account_id AND 
			dc.domain_id = ns.domain_id
		INNER JOIN email_config ec ON dc.user_id = ec.user_id
		WHERE ns.enabled = 1 
			AND ec.enabled = 1
			AND dc.renewal_date != '' 
			AND dc.renewal_date != 'permanent'
			AND dc.deleted_at IS NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expiringDomains []models.ExpiringDomain
	today := time.Now()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	for rows.Next() {
		var domain models.ExpiringDomain
		var renewalDate string
		var daysBefore int
		var lastNotifiedAt sql.NullTime

		if err := rows.Scan(&domain.UserID, &domain.AccountID, &domain.DomainID,
			&domain.DomainName, &renewalDate, &domain.RenewalURL,
			&daysBefore, &lastNotifiedAt, &domain.ToEmail, &domain.Language); err != nil {
			continue
		}

		// Parse renewal date
		expiry, err := time.Parse("2006-01-02", renewalDate)
		if err != nil {
			continue
		}
		expiry = time.Date(expiry.Year(), expiry.Month(), expiry.Day(), 0, 0, 0, 0, expiry.Location())

		// Calculate days remaining
		daysRemaining := int(expiry.Sub(today).Hours() / 24)
		domain.DaysRemaining = daysRemaining
		domain.RenewalDate = renewalDate

		// Check if notification is needed
		if daysRemaining <= daysBefore && daysRemaining >= 0 {
			// Check if already notified today
			if lastNotifiedAt.Valid {
				lastNotified := time.Date(lastNotifiedAt.Time.Year(), lastNotifiedAt.Time.Month(),
					lastNotifiedAt.Time.Day(), 0, 0, 0, 0, lastNotifiedAt.Time.Location())
				if !today.After(lastNotified) {
					continue // Already notified today
				}
			}

			expiringDomains = append(expiringDomains, domain)
		}
	}

	return expiringDomains, nil
}
