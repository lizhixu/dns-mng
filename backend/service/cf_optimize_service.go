package service

import (
	"context"
	"database/sql"
	"dns-mng/database"
	"dns-mng/models"
	"dns-mng/provider/cloudflare"
	"fmt"
	"log"
	"strings"
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

// helper to find an existing record by name and type in the slice
func findRecord(records []cloudflare.Record, name, recordType string) *cloudflare.Record {
	target := strings.ToLower(strings.TrimSuffix(name, "."))
	for _, r := range records {
		if strings.ToLower(strings.TrimSuffix(r.Name, ".")) == target && r.Type == recordType {
			return &r
		}
	}
	return nil
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
	intermediatePrefix := req.IntermediatePrefix
	if intermediatePrefix == "" {
		intermediatePrefix = "saas"
	}

	// 2. Find zone by name
	zone, err := s.client.GetZoneByName(ctx, apiToken, zoneName)
	if err != nil {
		return nil, fmt.Errorf("failed to find zone %s: %w", zoneName, err)
	}
	zoneID := zone.ID

	// 3. Build record names
	originRecordName := fmt.Sprintf("origin.%s", zoneName)
	intermediateRecordName := fmt.Sprintf("%s.%s", intermediatePrefix, zoneName)

	var cnameRecordName string
	if hostname == "@" || hostname == zoneName || hostname == "" {
		cnameRecordName = zoneName
	} else {
		cnameRecordName = fmt.Sprintf("%s.%s", hostname, zoneName)
	}
	customHostname := cnameRecordName

	// Fetch all existing records in the zone for smart reuse
	records, err := s.client.ListRecords(ctx, apiToken, zoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to list DNS records: %w", err)
	}

	// 4. Create/Update origin A record (proxied)
	var originRecordID string
	var createdOriginID string
	existingOrigin := findRecord(records, originRecordName, "A")
	if existingOrigin != nil {
		log.Printf("[CF Optimize] Updating existing origin A record: %s -> %s", originRecordName, originIP)
		updated, err := s.client.UpdateRecordWithProxied(ctx, apiToken, zoneID, existingOrigin.ID, "A", originRecordName, originIP, 1, true)
		if err != nil {
			return nil, fmt.Errorf("failed to update origin A record: %w", err)
		}
		originRecordID = updated.ID
	} else {
		log.Printf("[CF Optimize] Creating origin A record: %s -> %s", originRecordName, originIP)
		newRec, err := s.client.CreateRecordWithProxied(ctx, apiToken, zoneID, "A", originRecordName, originIP, 1, true)
		if err != nil {
			return nil, fmt.Errorf("failed to create origin A record: %w", err)
		}
		originRecordID = newRec.ID
		createdOriginID = newRec.ID
	}

	// 5. Create/Update intermediate CNAME record (gray cloud)
	var intermediateRecordID string
	var createdIntermediateID string
	existingIntermediate := findRecord(records, intermediateRecordName, "CNAME")
	if existingIntermediate != nil {
		log.Printf("[CF Optimize] Updating existing intermediate CNAME record: %s -> %s", intermediateRecordName, cnameTarget)
		updated, err := s.client.UpdateRecordWithProxied(ctx, apiToken, zoneID, existingIntermediate.ID, "CNAME", intermediateRecordName, cnameTarget, 1, false)
		if err != nil {
			if createdOriginID != "" {
				_ = s.client.DeleteRecord(ctx, apiToken, zoneID, createdOriginID)
			}
			return nil, fmt.Errorf("failed to update intermediate CNAME record: %w", err)
		}
		intermediateRecordID = updated.ID
	} else {
		log.Printf("[CF Optimize] Creating intermediate CNAME record: %s -> %s", intermediateRecordName, cnameTarget)
		newRec, err := s.client.CreateRecordWithProxied(ctx, apiToken, zoneID, "CNAME", intermediateRecordName, cnameTarget, 1, false)
		if err != nil {
			if createdOriginID != "" {
				_ = s.client.DeleteRecord(ctx, apiToken, zoneID, createdOriginID)
			}
			return nil, fmt.Errorf("failed to create intermediate CNAME record: %w", err)
		}
		intermediateRecordID = newRec.ID
		createdIntermediateID = newRec.ID
	}

	// 6. Create/Update business CNAME record pointing to intermediate domain
	var cnameRecordID string
	var createdCnameID string
	existingCname := findRecord(records, cnameRecordName, "CNAME")
	if existingCname != nil {
		log.Printf("[CF Optimize] Updating existing business CNAME record: %s -> %s", cnameRecordName, intermediateRecordName)
		updated, err := s.client.UpdateRecordWithProxied(ctx, apiToken, zoneID, existingCname.ID, "CNAME", cnameRecordName, intermediateRecordName, 1, false)
		if err != nil {
			if createdOriginID != "" {
				_ = s.client.DeleteRecord(ctx, apiToken, zoneID, createdOriginID)
			}
			if createdIntermediateID != "" {
				_ = s.client.DeleteRecord(ctx, apiToken, zoneID, createdIntermediateID)
			}
			return nil, fmt.Errorf("failed to update business CNAME record: %w", err)
		}
		cnameRecordID = updated.ID
	} else {
		log.Printf("[CF Optimize] Creating business CNAME record: %s -> %s", cnameRecordName, intermediateRecordName)
		newRec, err := s.client.CreateRecordWithProxied(ctx, apiToken, zoneID, "CNAME", cnameRecordName, intermediateRecordName, 1, false)
		if err != nil {
			if createdOriginID != "" {
				_ = s.client.DeleteRecord(ctx, apiToken, zoneID, createdOriginID)
			}
			if createdIntermediateID != "" {
				_ = s.client.DeleteRecord(ctx, apiToken, zoneID, createdIntermediateID)
			}
			return nil, fmt.Errorf("failed to create business CNAME record: %w", err)
		}
		cnameRecordID = newRec.ID
		createdCnameID = newRec.ID
	}

	// 7. Create custom hostname
	log.Printf("[CF Optimize] Creating custom hostname: %s (origin: %s)", customHostname, originRecordName)
	ch, err := s.client.CreateCustomHostname(ctx, apiToken, zoneID, customHostname, originRecordName)
	if err != nil {
		// Rollback newly created records
		if createdOriginID != "" {
			_ = s.client.DeleteRecord(ctx, apiToken, zoneID, createdOriginID)
		}
		if createdIntermediateID != "" {
			_ = s.client.DeleteRecord(ctx, apiToken, zoneID, createdIntermediateID)
		}
		if createdCnameID != "" {
			_ = s.client.DeleteRecord(ctx, apiToken, zoneID, createdCnameID)
		}
		errMsg := err.Error()
		if strings.Contains(errMsg, "403") || strings.Contains(errMsg, "Authentication error") {
			return nil, fmt.Errorf("custom hostname API 认证失败，请确保 API Token 具有「SSL 和证书」编辑权限，且账户已开通 Cloudflare for SaaS (Custom Hostnames): %w", err)
		}
		return nil, fmt.Errorf("failed to create custom hostname: %w", err)
	}
	customHostnameID := ch.ID
	status := ch.Status
	sslStatus := "pending"
	if ch.SSL != nil {
		sslStatus = ch.SSL.Status
	}

	// 8. Auto-create validation records on the same Cloudflare zone
	var validationRecordIDs []string

	// Ownership validation TXT record
	if ch.OwnershipVerification != nil && ch.OwnershipVerification.Status != "active" && ch.OwnershipVerification.Status != "verified" {
		txtName := ch.OwnershipVerification.Name
		txtValue := ch.OwnershipVerification.Value
		if txtName != "" && txtValue != "" {
			existingTxt := findRecord(records, txtName, "TXT")
			var recID string
			if existingTxt != nil {
				log.Printf("[CF Optimize] Updating existing ownership TXT: %s", txtName)
				updated, err := s.client.UpdateRecordWithProxied(ctx, apiToken, zoneID, existingTxt.ID, "TXT", txtName, txtValue, 1, false)
				if err == nil {
					recID = updated.ID
				}
			} else {
				log.Printf("[CF Optimize] Creating ownership TXT: %s", txtName)
				newRec, err := s.client.CreateRecordWithProxied(ctx, apiToken, zoneID, "TXT", txtName, txtValue, 1, false)
				if err == nil {
					recID = newRec.ID
				}
			}
			if recID != "" {
				validationRecordIDs = append(validationRecordIDs, recID)
			}
		}
	}

	// SSL Validation records (e.g. DCV delegation CNAME or TXT)
	if ch.SSL != nil && len(ch.SSL.ValidationRecords) > 0 {
		for _, v := range ch.SSL.ValidationRecords {
			if v.Status != "active" && v.Status != "verified" && v.TxtName != "" && v.TxtValue != "" {
				recType := "TXT"
				if strings.Contains(v.TxtValue, "dcv.cloudflare.com") {
					recType = "CNAME"
				}
				existingRec := findRecord(records, v.TxtName, recType)
				var recID string
				if existingRec != nil {
					log.Printf("[CF Optimize] Updating existing SSL verification %s: %s", recType, v.TxtName)
					updated, err := s.client.UpdateRecordWithProxied(ctx, apiToken, zoneID, existingRec.ID, recType, v.TxtName, v.TxtValue, 1, false)
					if err == nil {
						recID = updated.ID
					}
				} else {
					log.Printf("[CF Optimize] Creating SSL verification %s: %s", recType, v.TxtName)
					newRec, err := s.client.CreateRecordWithProxied(ctx, apiToken, zoneID, recType, v.TxtName, v.TxtValue, 1, false)
					if err == nil {
						recID = newRec.ID
					}
				}
				if recID != "" {
					validationRecordIDs = append(validationRecordIDs, recID)
				}
			}
		}
	}

	validationRecordIDsJoined := strings.Join(validationRecordIDs, ",")

	// 9. Save to database
	now := time.Now()
	result, err := database.DB.Exec(
		`INSERT INTO cf_optimize
			(user_id, account_id, zone_id, zone_name, origin_ip, origin_record_name, origin_record_id,
			 cname_target, cname_record_name, cname_record_id, custom_hostname, custom_hostname_id,
			 status, ssl_status, intermediate_record_name, intermediate_record_id, validation_record_ids,
			 created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, accountID, zoneID, zoneName, originIP, originRecordName, originRecordID,
		cnameTarget, cnameRecordName, cnameRecordID, customHostname, customHostnameID,
		status, sslStatus, intermediateRecordName, intermediateRecordID, validationRecordIDsJoined,
		now, now,
	)
	if err != nil {
		// Cleanup created validation records
		for _, rID := range validationRecordIDs {
			_ = s.client.DeleteRecord(ctx, apiToken, zoneID, rID)
		}
		// Rollback other newly created records
		if createdOriginID != "" {
			_ = s.client.DeleteRecord(ctx, apiToken, zoneID, createdOriginID)
		}
		if createdIntermediateID != "" {
			_ = s.client.DeleteRecord(ctx, apiToken, zoneID, createdIntermediateID)
		}
		if createdCnameID != "" {
			_ = s.client.DeleteRecord(ctx, apiToken, zoneID, createdCnameID)
		}
		// Delete custom hostname
		_ = s.client.DeleteCustomHostname(ctx, apiToken, zoneID, customHostnameID)

		return nil, fmt.Errorf("failed to save config to database: %w", err)
	}

	id, _ := result.LastInsertId()
	return &models.CFOptimize{
		ID:                     id,
		UserID:                 userID,
		AccountID:              accountID,
		ZoneID:                 zoneID,
		ZoneName:               zoneName,
		OriginIP:               originIP,
		OriginRecordName:       originRecordName,
		OriginRecordID:         originRecordID,
		CnameTarget:            cnameTarget,
		CnameRecordName:        cnameRecordName,
		CnameRecordID:          cnameRecordID,
		CustomHostname:         customHostname,
		CustomHostnameID:       customHostnameID,
		Status:                 status,
		SSLStatus:              sslStatus,
		IntermediateRecordName: intermediateRecordName,
		IntermediateRecordID:   intermediateRecordID,
		ValidationRecordIDs:    validationRecordIDsJoined,
		CreatedAt:              now,
		UpdatedAt:              now,
	}, nil
}

// List returns all CF optimize configs for a user
func (s *CFOptimizeService) List(userID int64) ([]models.CFOptimize, error) {
	rows, err := database.DB.Query(
		`SELECT id, user_id, account_id, zone_id, zone_name, origin_ip, origin_record_name, origin_record_id,
		        cname_target, cname_record_name, cname_record_id, custom_hostname, custom_hostname_id,
		        status, ssl_status, intermediate_record_name, intermediate_record_id, validation_record_ids,
		        created_at, updated_at
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
			&c.Status, &c.SSLStatus, &c.IntermediateRecordName, &c.IntermediateRecordID, &c.ValidationRecordIDs,
			&c.CreatedAt, &c.UpdatedAt,
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

			// 1. Delete custom hostname
			if config.CustomHostnameID != "" {
				_ = s.client.DeleteCustomHostname(ctx, apiToken, config.ZoneID, config.CustomHostnameID)
			}

			// 2. Delete business CNAME record
			if config.CnameRecordID != "" {
				_ = s.client.DeleteRecord(ctx, apiToken, config.ZoneID, config.CnameRecordID)
			}

			// 3. Delete auto-created validation records
			if config.ValidationRecordIDs != "" {
				vIDs := strings.Split(config.ValidationRecordIDs, ",")
				for _, rID := range vIDs {
					rID = strings.TrimSpace(rID)
					if rID != "" {
						_ = s.client.DeleteRecord(ctx, apiToken, config.ZoneID, rID)
					}
				}
			}

			// 4. Reference count check for intermediate gateway CNAME record
			if config.IntermediateRecordID != "" {
				var count int
				err := database.DB.QueryRow(
					"SELECT COUNT(*) FROM cf_optimize WHERE zone_id = ? AND id != ? AND intermediate_record_id = ?",
					config.ZoneID, configID, config.IntermediateRecordID,
				).Scan(&count)
				if err == nil && count == 0 {
					log.Printf("[CF Optimize] Cleaning up unused intermediate CNAME record: %s", config.IntermediateRecordName)
					_ = s.client.DeleteRecord(ctx, apiToken, config.ZoneID, config.IntermediateRecordID)
				}
			}

			// 5. Reference count check for origin A record
			if config.OriginRecordID != "" {
				var count int
				err := database.DB.QueryRow(
					"SELECT COUNT(*) FROM cf_optimize WHERE zone_id = ? AND id != ? AND origin_record_id = ?",
					config.ZoneID, configID, config.OriginRecordID,
				).Scan(&count)
				if err == nil && count == 0 {
					log.Printf("[CF Optimize] Cleaning up unused origin A record: %s", config.OriginRecordName)
					_ = s.client.DeleteRecord(ctx, apiToken, config.ZoneID, config.OriginRecordID)
				}
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
		        status, ssl_status, intermediate_record_name, intermediate_record_id, validation_record_ids,
		        created_at, updated_at
		 FROM cf_optimize WHERE id = ? AND user_id = ?`, configID, userID,
	).Scan(
		&c.ID, &c.UserID, &c.AccountID, &c.ZoneID, &c.ZoneName, &c.OriginIP,
		&c.OriginRecordName, &c.OriginRecordID, &c.CnameTarget, &c.CnameRecordName,
		&c.CnameRecordID, &c.CustomHostname, &c.CustomHostnameID,
		&c.Status, &c.SSLStatus, &c.IntermediateRecordName, &c.IntermediateRecordID, &c.ValidationRecordIDs,
		&c.CreatedAt, &c.UpdatedAt,
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
