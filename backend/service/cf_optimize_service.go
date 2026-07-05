package service

import (
	"context"
	"database/sql"
	"dns-mng/database"
	"dns-mng/models"
	"dns-mng/provider/cloudflare"
	"fmt"
	"log"
	"time"
)

// CFOptimizeService handles Cloudflare CDN optimization operations
type CFOptimizeService struct {
	client *cloudflare.Client
}

func NewCFOptimizeService() *CFOptimizeService {
	return &CFOptimizeService{
		client: cloudflare.NewClient(),
	}
}

// Create performs one-click CDN optimization
func (s *CFOptimizeService) Create(ctx context.Context, userID int64, accountID int64, req *models.CreateCFOptimizeRequest) (*models.CFOptimize, error) {
	// 1. Get account and verify ownership
	account, err := s.getAccount(userID, accountID)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}
	if account.ProviderType != "cloudflare" {
		return nil, fmt.Errorf("account is not a Cloudflare provider (type: %s)", account.ProviderType)
	}

	apiToken := account.APIKey
	zoneName := req.ZoneName
	hostname := req.Hostname
	originIP := req.OriginIP
	cnameTarget := req.CnameTarget
	if cnameTarget == "" {
		cnameTarget = "cloudflare.468123.xyz"
	}

	// 2. Find zone by name
	zone, err := s.client.GetZoneByName(ctx, apiToken, zoneName)
	if err != nil {
		return nil, fmt.Errorf("failed to find zone %s: %w", zoneName, err)
	}
	zoneID := zone.ID

	// 3. Build record names
	originRecordName := fmt.Sprintf("origin.%s", zoneName)
	cnameRecordName := hostname
	if hostname == "@" || hostname == zoneName {
		cnameRecordName = zoneName
	} else if hostname != "" {
		cnameRecordName = fmt.Sprintf("%s.%s", hostname, zoneName)
	}
	customHostname := cnameRecordName

	// 4. Create origin A record (proxied)
	log.Printf("[CF Optimize] Creating origin A record: %s -> %s (proxied)", originRecordName, originIP)
	originRecord, err := s.client.CreateRecordWithProxied(ctx, apiToken, zoneID, "A", originRecordName, originIP, 1, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create origin A record: %w", err)
	}
	originRecordID := originRecord.ID

	// 5. Create CNAME record pointing to optimized domain
	log.Printf("[CF Optimize] Creating CNAME record: %s -> %s", cnameRecordName, cnameTarget)
	cnameRecord, err := s.client.CreateRecordWithProxied(ctx, apiToken, zoneID, "CNAME", cnameRecordName, cnameTarget, 1, false)
	if err != nil {
		// Rollback: delete the origin record
		_ = s.client.DeleteRecord(ctx, apiToken, zoneID, originRecordID)
		return nil, fmt.Errorf("failed to create CNAME record: %w", err)
	}
	cnameRecordID := cnameRecord.ID

	// 6. Create custom hostname
	log.Printf("[CF Optimize] Creating custom hostname: %s (origin: %s)", customHostname, originRecordName)
	ch, err := s.client.CreateCustomHostname(ctx, apiToken, zoneID, customHostname, originRecordName)
	if err != nil {
		// Rollback: delete both records
		_ = s.client.DeleteRecord(ctx, apiToken, zoneID, originRecordID)
		_ = s.client.DeleteRecord(ctx, apiToken, zoneID, cnameRecordID)
		return nil, fmt.Errorf("failed to create custom hostname: %w", err)
	}
	customHostnameID := ch.ID
	status := ch.Status
	sslStatus := "pending"
	if ch.SSL != nil {
		sslStatus = ch.SSL.Status
	}

	// 7. Save to database
	now := time.Now()
	result, err := database.DB.Exec(
		`INSERT INTO cf_optimize
			(user_id, account_id, zone_id, zone_name, origin_ip, origin_record_name, origin_record_id,
			 cname_target, cname_record_name, cname_record_id, custom_hostname, custom_hostname_id,
			 status, ssl_status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, accountID, zoneID, zoneName, originIP, originRecordName, originRecordID,
		cnameTarget, cnameRecordName, cnameRecordID, customHostname, customHostnameID,
		status, sslStatus, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	id, _ := result.LastInsertId()
	return &models.CFOptimize{
		ID:               id,
		UserID:           userID,
		AccountID:        accountID,
		ZoneID:           zoneID,
		ZoneName:         zoneName,
		OriginIP:         originIP,
		OriginRecordName: originRecordName,
		OriginRecordID:   originRecordID,
		CnameTarget:      cnameTarget,
		CnameRecordName:  cnameRecordName,
		CnameRecordID:    cnameRecordID,
		CustomHostname:   customHostname,
		CustomHostnameID: customHostnameID,
		Status:           status,
		SSLStatus:        sslStatus,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

// List returns all CF optimize configs for a user
func (s *CFOptimizeService) List(userID int64) ([]models.CFOptimize, error) {
	rows, err := database.DB.Query(
		`SELECT id, user_id, account_id, zone_id, zone_name, origin_ip, origin_record_name, origin_record_id,
		        cname_target, cname_record_name, cname_record_id, custom_hostname, custom_hostname_id,
		        status, ssl_status, created_at, updated_at
		 FROM cf_optimize WHERE user_id = ? ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []models.CFOptimize
	for rows.Next() {
		var c models.CFOptimize
		if err := rows.Scan(
			&c.ID, &c.UserID, &c.AccountID, &c.ZoneID, &c.ZoneName, &c.OriginIP,
			&c.OriginRecordName, &c.OriginRecordID, &c.CnameTarget, &c.CnameRecordName,
			&c.CnameRecordID, &c.CustomHostname, &c.CustomHostnameID,
			&c.Status, &c.SSLStatus, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return configs, nil
}

// Refresh updates the status of a custom hostname from Cloudflare
func (s *CFOptimizeService) Refresh(ctx context.Context, userID, configID int64) (*models.CFOptimize, error) {
	// Get config
	config, err := s.get(userID, configID)
	if err != nil {
		return nil, fmt.Errorf("config not found: %w", err)
	}

	// Get account API key
	account, err := s.getAccount(userID, config.AccountID)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	// Query Cloudflare for current status
	ch, err := s.client.GetCustomHostname(ctx, account.APIKey, config.ZoneID, config.CustomHostnameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom hostname status: %w", err)
	}

	// Update local status
	status := ch.Status
	sslStatus := "pending"
	if ch.SSL != nil {
		sslStatus = ch.SSL.Status
	}

	now := time.Now()
	_, err = database.DB.Exec(
		"UPDATE cf_optimize SET status = ?, ssl_status = ?, updated_at = ? WHERE id = ?",
		status, sslStatus, now, configID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	config.Status = status
	config.SSLStatus = sslStatus
	config.UpdatedAt = now
	return config, nil
}

// Delete removes a CF optimize config and optionally cleans up DNS records
func (s *CFOptimizeService) Delete(ctx context.Context, userID, configID int64, cleanup bool) error {
	config, err := s.get(userID, configID)
	if err != nil {
		return fmt.Errorf("config not found: %w", err)
	}

	if cleanup {
		account, err := s.getAccount(userID, config.AccountID)
		if err == nil {
			apiToken := account.APIKey
			// Delete custom hostname
			if config.CustomHostnameID != "" {
				_ = s.client.DeleteCustomHostname(ctx, apiToken, config.ZoneID, config.CustomHostnameID)
			}
			// Delete CNAME record
			if config.CnameRecordID != "" {
				_ = s.client.DeleteRecord(ctx, apiToken, config.ZoneID, config.CnameRecordID)
			}
			// Delete origin A record
			if config.OriginRecordID != "" {
				_ = s.client.DeleteRecord(ctx, apiToken, config.ZoneID, config.OriginRecordID)
			}
		}
	}

	_, err = database.DB.Exec("DELETE FROM cf_optimize WHERE id = ? AND user_id = ?", configID, userID)
	return err
}

// get retrieves a single config by ID with ownership check
func (s *CFOptimizeService) get(userID, configID int64) (*models.CFOptimize, error) {
	var c models.CFOptimize
	err := database.DB.QueryRow(
		`SELECT id, user_id, account_id, zone_id, zone_name, origin_ip, origin_record_name, origin_record_id,
		        cname_target, cname_record_name, cname_record_id, custom_hostname, custom_hostname_id,
		        status, ssl_status, created_at, updated_at
		 FROM cf_optimize WHERE id = ? AND user_id = ?`, configID, userID,
	).Scan(
		&c.ID, &c.UserID, &c.AccountID, &c.ZoneID, &c.ZoneName, &c.OriginIP,
		&c.OriginRecordName, &c.OriginRecordID, &c.CnameTarget, &c.CnameRecordName,
		&c.CnameRecordID, &c.CustomHostname, &c.CustomHostnameID,
		&c.Status, &c.SSLStatus, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("config not found")
		}
		return nil, err
	}
	return &c, nil
}

// getAccount retrieves an account by ID with ownership check
func (s *CFOptimizeService) getAccount(userID, accountID int64) (*models.Account, error) {
	var a models.Account
	err := database.DB.QueryRow(
		"SELECT id, user_id, name, provider_type, api_key, created_at, updated_at FROM accounts WHERE id = ? AND user_id = ?",
		accountID, userID,
	).Scan(&a.ID, &a.UserID, &a.Name, &a.ProviderType, &a.APIKey, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}
