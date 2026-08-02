package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"dns-mng/database"
)

// ─── 导出结构体 ────────────────────────────────────────────────

type backupData struct {
	Accounts          []backupAccount          `json:"accounts"`
	DomainCaches      []backupDomainCache      `json:"domain_caches"`
	DDNSToken         *backupDDNSToken         `json:"ddns_token"`
	EmailConfig       *backupEmailConfig       `json:"email_config"`
	WHOISConfig       *backupWHOISConfig       `json:"whois_config,omitempty"`
	DNSHEAutoRenew    *backupDNSHEAutoRenew    `json:"dnshe_auto_renew,omitempty"`
	CFOptimizeConfigs []backupCFOptimizeConfig `json:"cf_optimize_configs,omitempty"`
}

type backupAccount struct {
	Name         string `json:"name"`
	ProviderType string `json:"provider_type"`
	APIKey       string `json:"api_key"`
}

type backupDomainCache struct {
	AccountKey        string `json:"account_key"` // "provider_type::name"
	DomainID          string `json:"domain_id"`
	DomainName        string `json:"domain_name"`
	RenewalDate       string `json:"renewal_date,omitempty"`
	RenewalURL        string `json:"renewal_url,omitempty"`
	UsesDNSHEDNS      *bool  `json:"uses_dnshe_dns,omitempty"`
	DeletedAt         string `json:"deleted_at,omitempty"`
	LastSyncAt        string `json:"last_sync_at,omitempty"`
	ProviderUpdatedOn string `json:"provider_updated_on,omitempty"`
	HasNotification   bool   `json:"has_notification,omitempty"`
	DaysBefore        int    `json:"days_before,omitempty"`
	NotifyEnabled     bool   `json:"notify_enabled,omitempty"`
	LastNotifiedAt    string `json:"last_notified_at,omitempty"`
}

type backupDDNSToken struct {
	Token   string `json:"token"`
	Enabled bool   `json:"enabled"`
}

type backupEmailConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	FromEmail    string `json:"from_email"`
	FromName     string `json:"from_name"`
	ToEmail      string `json:"to_email"`
	Language     string `json:"language"`
	Enabled      bool   `json:"enabled"`
}

type backupWHOISConfig struct {
	APIKey string `json:"api_key"`
}

type backupDNSHEAutoRenew struct {
	Enabled    bool   `json:"enabled"`
	DaysBefore int    `json:"days_before"`
	LastRunAt  string `json:"last_run_at,omitempty"`
}

type backupCFOptimizeConfig struct {
	AccountKey             string `json:"account_key"`
	ZoneID                 string `json:"zone_id"`
	ZoneName               string `json:"zone_name"`
	OriginIP               string `json:"origin_ip"`
	OriginRecordName       string `json:"origin_record_name"`
	OriginRecordID         string `json:"origin_record_id,omitempty"`
	CnameTarget            string `json:"cname_target"`
	CnameRecordName        string `json:"cname_record_name"`
	CnameRecordID          string `json:"cname_record_id,omitempty"`
	CustomHostname         string `json:"custom_hostname"`
	CustomHostnameID       string `json:"custom_hostname_id,omitempty"`
	Status                 string `json:"status"`
	SSLStatus              string `json:"ssl_status"`
	IntermediateRecordName string `json:"intermediate_record_name,omitempty"`
	IntermediateRecordID   string `json:"intermediate_record_id,omitempty"`
	ValidationRecordIDs    string `json:"validation_record_ids,omitempty"`
}

type backupFile struct {
	Version    int        `json:"version"`
	ExportedAt string     `json:"exported_at"`
	Encrypted  bool       `json:"encrypted"`
	Data       backupData `json:"data"`
}

// ─── 导入结果 ────────────────────────────────────────────────────

type ImportResult struct {
	AccountsImported       int  `json:"accounts_imported"`
	AccountsSkipped        int  `json:"accounts_skipped"`
	DomainCachesImported   int  `json:"domain_caches_imported"`
	DomainCachesSkipped    int  `json:"domain_caches_skipped"`
	DDNSTokenImported      bool `json:"ddns_token_imported"`
	DDNSTokenSkipped       bool `json:"ddns_token_skipped"`
	EmailConfigImported    bool `json:"email_config_imported"`
	EmailConfigSkipped     bool `json:"email_config_skipped"`
	WHOISConfigImported    bool `json:"whois_config_imported"`
	WHOISConfigSkipped     bool `json:"whois_config_skipped"`
	DNSHEAutoRenewImported bool `json:"dnshe_auto_renew_imported"`
	DNSHEAutoRenewSkipped  bool `json:"dnshe_auto_renew_skipped"`
	CFOptimizeImported     int  `json:"cf_optimize_imported"`
	CFOptimizeSkipped      int  `json:"cf_optimize_skipped"`
}

// ─── BackupService ──────────────────────────────────────────────

type BackupService struct {
	accountService      *AccountService
	domainCacheService  *DomainCacheService
	ddnsTokenService    *DDNSTokenService
	emailService        *EmailService
	notificationService *NotificationService
}

func NewBackupService(
	accountService *AccountService,
	domainCacheService *DomainCacheService,
	ddnsTokenService *DDNSTokenService,
	emailService *EmailService,
	notificationService *NotificationService,
) *BackupService {
	return &BackupService{
		accountService:      accountService,
		domainCacheService:  domainCacheService,
		ddnsTokenService:    ddnsTokenService,
		emailService:        emailService,
		notificationService: notificationService,
	}
}

// Export 导出用户的所有配置为 JSON 字节（可选 AES 加密）。
func (s *BackupService) Export(userID int64, password string) ([]byte, error) {
	data := backupData{}

	// 1. 账户
	accounts, err := s.accountService.List(userID)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	for _, a := range accounts {
		data.Accounts = append(data.Accounts, backupAccount{
			Name:         a.Name,
			ProviderType: a.ProviderType,
			APIKey:       a.APIKey,
		})
	}

	// 2. 域名缓存（续费信息 + 通知设置）
	accountKeyMap, err := s.buildAccountKeyMap(userID)
	if err != nil {
		return nil, fmt.Errorf("build account key map: %w", err)
	}
	caches, err := s.listAllDomainCaches(userID)
	if err != nil {
		return nil, fmt.Errorf("export domain caches: %w", err)
	}

	notifMap := make(map[string]backupDomainCache)
	notifications, err := s.notificationService.GetAllNotificationSettings(userID)
	if err != nil {
		return nil, fmt.Errorf("export notification settings: %w", err)
	}
	for _, ns := range notifications {
		lastNotified := ""
		if ns.LastNotifiedAt != nil {
			lastNotified = ns.LastNotifiedAt.UTC().Format(time.RFC3339)
		}
		notifMap[fmt.Sprintf("%d:%s", ns.AccountID, ns.DomainID)] = backupDomainCache{
			HasNotification: true,
			DaysBefore:      ns.DaysBefore,
			NotifyEnabled:   ns.Enabled,
			LastNotifiedAt:  lastNotified,
		}
	}

	for _, c := range caches {
		accountKey := accountKeyMap[c.AccountID]
		if accountKey == "" {
			accountKey = fmt.Sprintf("unknown::%d", c.AccountID)
		}
		usesDNSHE := c.UsesDNSHEDNS
		entry := backupDomainCache{
			AccountKey:        accountKey,
			DomainID:          c.DomainID,
			DomainName:        c.DomainName,
			RenewalDate:       c.RenewalDate,
			RenewalURL:        c.RenewalURL,
			UsesDNSHEDNS:      &usesDNSHE,
			DeletedAt:         formatNullTime(c.DeletedAt),
			LastSyncAt:        formatNullTime(c.LastSyncAt),
			ProviderUpdatedOn: formatNullTime(c.ProviderUpdatedOn),
		}
		if notif, ok := notifMap[fmt.Sprintf("%d:%s", c.AccountID, c.DomainID)]; ok {
			entry.HasNotification = true
			entry.DaysBefore = notif.DaysBefore
			entry.NotifyEnabled = notif.NotifyEnabled
			entry.LastNotifiedAt = notif.LastNotifiedAt
		}
		data.DomainCaches = append(data.DomainCaches, entry)
	}

	// 3. DDNS Token
	token, err := s.ddnsTokenService.GetToken(userID)
	if err != nil {
		return nil, fmt.Errorf("export ddns token: %w", err)
	}
	if token != nil {
		data.DDNSToken = &backupDDNSToken{Token: token.Token, Enabled: token.Enabled}
	}

	// 4. 邮件配置（含密码）
	emailCfg, err := s.emailService.getEmailConfigWithPassword(userID)
	if err != nil {
		return nil, fmt.Errorf("export email config: %w", err)
	}
	if emailCfg != nil {
		data.EmailConfig = &backupEmailConfig{
			SMTPHost:     emailCfg.SMTPHost,
			SMTPPort:     emailCfg.SMTPPort,
			SMTPUsername: emailCfg.SMTPUsername,
			SMTPPassword: emailCfg.SMTPPassword,
			FromEmail:    emailCfg.FromEmail,
			FromName:     emailCfg.FromName,
			ToEmail:      emailCfg.ToEmail,
			Language:     emailCfg.Language,
			Enabled:      emailCfg.Enabled,
		}
	}

	whoisConfig, err := s.getWHOISConfig(userID)
	if err != nil {
		return nil, fmt.Errorf("export whois config: %w", err)
	}
	if whoisConfig != nil {
		data.WHOISConfig = whoisConfig
	}

	autoRenew, err := s.getDNSHEAutoRenew(userID)
	if err != nil {
		return nil, fmt.Errorf("export dnshe auto-renew config: %w", err)
	}
	if autoRenew != nil {
		data.DNSHEAutoRenew = autoRenew
	}

	cfConfigs, err := s.listCFOptimizeConfigs(userID, accountKeyMap)
	if err != nil {
		return nil, fmt.Errorf("export cf optimize configs: %w", err)
	}
	data.CFOptimizeConfigs = cfConfigs

	file := backupFile{
		Version:    1,
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Encrypted:  password != "",
		Data:       data,
	}

	plainJSON, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal backup: %w", err)
	}

	if password != "" {
		encrypted, err := EncryptBackup(plainJSON, password)
		if err != nil {
			return nil, fmt.Errorf("encrypt backup: %w", err)
		}
		return encrypted, nil
	}

	return plainJSON, nil
}

// Import 从备份文件导入配置。overwrite=true 时覆盖已存在的同名项，否则跳过。
func (s *BackupService) Import(userID int64, fileBytes []byte, password string, overwrite bool) (*ImportResult, error) {
	plainJSON, err := DecryptBackup(fileBytes, password)
	if err != nil {
		return nil, err
	}

	var file backupFile
	if err := json.Unmarshal(plainJSON, &file); err != nil {
		return nil, fmt.Errorf("解析备份文件失败: %w", err)
	}
	if file.Version != 1 {
		return nil, fmt.Errorf("不支持的备份版本: %d", file.Version)
	}

	result := &ImportResult{}

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	accountKeyToID := make(map[string]int64)
	for _, acc := range file.Data.Accounts {
		key := fmt.Sprintf("%s::%s", acc.ProviderType, acc.Name)
		existingID, err := s.findAccountByKey(userID, key, tx)
		if err != nil {
			return nil, fmt.Errorf("find account %q: %w", key, err)
		}

		if existingID > 0 && !overwrite {
			result.AccountsSkipped++
			accountKeyToID[key] = existingID
			continue
		}

		if existingID > 0 && overwrite {
			_, err := tx.Exec(
				"UPDATE accounts SET api_key = ?, updated_at = ? WHERE id = ? AND user_id = ?",
				acc.APIKey, time.Now(), existingID, userID,
			)
			if err != nil {
				return nil, fmt.Errorf("update account %q: %w", acc.Name, err)
			}
			accountKeyToID[key] = existingID
			result.AccountsImported++
			continue
		}

		res, err := tx.Exec(
			"INSERT INTO accounts (user_id, name, provider_type, api_key, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			userID, acc.Name, acc.ProviderType, acc.APIKey, time.Now(), time.Now(),
		)
		if err != nil {
			return nil, fmt.Errorf("insert account %q: %w", acc.Name, err)
		}
		id, _ := res.LastInsertId()
		accountKeyToID[key] = id
		result.AccountsImported++
	}

	for _, dc := range file.Data.DomainCaches {
		accountID, ok := accountKeyToID[dc.AccountKey]
		if !ok {
			foundID, err := s.findAccountByKey(userID, dc.AccountKey, tx)
			if err != nil {
				return nil, fmt.Errorf("find account %q for domain cache: %w", dc.AccountKey, err)
			}
			if foundID > 0 {
				accountID = foundID
			} else {
				log.Printf("Warning: skip domain cache for unknown account key: %s", dc.AccountKey)
				result.DomainCachesSkipped++
				continue
			}
		}

		var existingID int64
		err := tx.QueryRow(
			"SELECT id FROM domain_cache WHERE user_id = ? AND account_id = ? AND domain_id = ?",
			userID, accountID, dc.DomainID,
		).Scan(&existingID)

		now := time.Now()
		usesDNSHE := 1
		if dc.UsesDNSHEDNS != nil && !*dc.UsesDNSHEDNS {
			usesDNSHE = 0
		}
		deletedAt := parseNullableTime(dc.DeletedAt)
		lastSyncAt := parseNullableTime(dc.LastSyncAt)
		providerUpdatedOn := parseNullableTime(dc.ProviderUpdatedOn)

		if err == sql.ErrNoRows {
			_, err = tx.Exec(
				`INSERT INTO domain_cache (user_id, account_id, domain_id, domain_name, renewal_date, renewal_url, uses_dnshe_dns, deleted_at, last_sync_at, provider_updated_on, created_at, updated_at)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				userID, accountID, dc.DomainID, dc.DomainName, dc.RenewalDate, dc.RenewalURL, usesDNSHE, deletedAt, lastSyncAt, providerUpdatedOn, now, now,
			)
		} else if err == nil {
			if overwrite {
				_, err = tx.Exec(
					`UPDATE domain_cache SET domain_name=?, renewal_date=?, renewal_url=?, uses_dnshe_dns=?, deleted_at=?, last_sync_at=?, provider_updated_on=?, updated_at=? WHERE id=?`,
					dc.DomainName, dc.RenewalDate, dc.RenewalURL, usesDNSHE, deletedAt, lastSyncAt, providerUpdatedOn, now, existingID,
				)
			} else {
				result.DomainCachesSkipped++
				continue
			}
		}

		if err != nil {
			return nil, fmt.Errorf("import domain cache %s: %w", dc.DomainName, err)
		}
		result.DomainCachesImported++

		if dc.HasNotification || dc.DaysBefore > 0 || dc.NotifyEnabled || dc.LastNotifiedAt != "" {
			if err := s.importNotificationSetting(tx, userID, accountID, dc, now, overwrite); err != nil {
				return nil, err
			}
		}
	}

	if file.Data.DDNSToken != nil {
		existing, err := s.findDDNSToken(userID, tx)
		if err != nil {
			return nil, fmt.Errorf("check ddns token: %w", err)
		}

		if existing != "" && !overwrite {
			result.DDNSTokenSkipped = true
		} else {
			ownerID, err := s.findDDNSTokenOwner(file.Data.DDNSToken.Token, tx)
			if err != nil {
				return nil, fmt.Errorf("check ddns token owner: %w", err)
			}
			if ownerID > 0 && ownerID != userID {
				result.DDNSTokenSkipped = true
			} else {
				enabled := backupBoolToInt(file.Data.DDNSToken.Enabled)
				if existing != "" {
					_, err = tx.Exec(
						"UPDATE ddns_tokens SET token = ?, enabled = ?, updated_at = datetime('now') WHERE user_id = ?",
						file.Data.DDNSToken.Token, enabled, userID,
					)
				} else {
					_, err = tx.Exec(
						"INSERT INTO ddns_tokens (user_id, token, enabled, created_at, updated_at) VALUES (?, ?, ?, datetime('now'), datetime('now'))",
						userID, file.Data.DDNSToken.Token, enabled,
					)
				}
				if err != nil {
					return nil, fmt.Errorf("import ddns token: %w", err)
				}
				result.DDNSTokenImported = true
			}
		}
	}

	if file.Data.EmailConfig != nil {
		ec := file.Data.EmailConfig
		existing, err := s.findEmailConfig(userID, tx)
		if err != nil {
			return nil, fmt.Errorf("check email config: %w", err)
		}

		if existing && !overwrite {
			result.EmailConfigSkipped = true
		} else {
			enabled := backupBoolToInt(ec.Enabled)
			now := time.Now()
			if existing {
				_, err = tx.Exec(
					`UPDATE email_config SET smtp_host=?, smtp_port=?, smtp_username=?, smtp_password=?,
					 from_email=?, from_name=?, to_email=?, language=?, enabled=?, updated_at=?
					 WHERE user_id=?`,
					ec.SMTPHost, ec.SMTPPort, ec.SMTPUsername, ec.SMTPPassword,
					ec.FromEmail, ec.FromName, ec.ToEmail, ec.Language, enabled, now, userID,
				)
			} else {
				_, err = tx.Exec(
					`INSERT INTO email_config (user_id, smtp_host, smtp_port, smtp_username, smtp_password,
					 from_email, from_name, to_email, language, enabled, created_at, updated_at)
					 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
					userID, ec.SMTPHost, ec.SMTPPort, ec.SMTPUsername, ec.SMTPPassword,
					ec.FromEmail, ec.FromName, ec.ToEmail, ec.Language, enabled, now, now,
				)
			}
			if err != nil {
				return nil, fmt.Errorf("import email config: %w", err)
			}
			result.EmailConfigImported = true
		}
	}

	if file.Data.WHOISConfig != nil {
		imported, skipped, err := s.importWHOISConfig(tx, userID, file.Data.WHOISConfig, overwrite)
		if err != nil {
			return nil, err
		}
		result.WHOISConfigImported = imported
		result.WHOISConfigSkipped = skipped
	}

	if file.Data.DNSHEAutoRenew != nil {
		imported, skipped, err := s.importDNSHEAutoRenew(tx, userID, file.Data.DNSHEAutoRenew, overwrite)
		if err != nil {
			return nil, err
		}
		result.DNSHEAutoRenewImported = imported
		result.DNSHEAutoRenewSkipped = skipped
	}

	for _, cfg := range file.Data.CFOptimizeConfigs {
		accountID, ok := accountKeyToID[cfg.AccountKey]
		if !ok {
			foundID, err := s.findAccountByKey(userID, cfg.AccountKey, tx)
			if err != nil {
				return nil, fmt.Errorf("find account %q for cf optimize config: %w", cfg.AccountKey, err)
			}
			if foundID > 0 {
				accountID = foundID
			} else {
				result.CFOptimizeSkipped++
				continue
			}
		}
		imported, skipped, err := s.importCFOptimizeConfig(tx, userID, accountID, cfg, overwrite)
		if err != nil {
			return nil, err
		}
		if imported {
			result.CFOptimizeImported++
		}
		if skipped {
			result.CFOptimizeSkipped++
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return result, nil
}

// ─── 内部辅助方法 ──────────────────────────────────────────────

func (s *BackupService) findAccountByKey(userID int64, key string, tx *sql.Tx) (int64, error) {
	parts := strings.SplitN(key, "::", 2)
	if len(parts) < 2 {
		return 0, nil
	}
	var id int64
	err := tx.QueryRow(
		"SELECT id FROM accounts WHERE user_id = ? AND provider_type = ? AND name = ?",
		userID, parts[0], parts[1],
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return id, err
}

func (s *BackupService) findDDNSToken(userID int64, tx *sql.Tx) (string, error) {
	var token string
	err := tx.QueryRow("SELECT token FROM ddns_tokens WHERE user_id = ?", userID).Scan(&token)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return token, err
}

func (s *BackupService) findDDNSTokenOwner(token string, tx *sql.Tx) (int64, error) {
	if strings.TrimSpace(token) == "" {
		return 0, nil
	}
	var userID int64
	err := tx.QueryRow("SELECT user_id FROM ddns_tokens WHERE token = ?", token).Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return userID, err
}

func (s *BackupService) findEmailConfig(userID int64, tx *sql.Tx) (bool, error) {
	var id int64
	err := tx.QueryRow("SELECT id FROM email_config WHERE user_id = ?", userID).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *BackupService) buildAccountKeyMap(userID int64) (map[int64]string, error) {
	accounts, err := s.accountService.List(userID)
	if err != nil {
		return nil, err
	}
	m := make(map[int64]string, len(accounts))
	for _, a := range accounts {
		m[a.ID] = fmt.Sprintf("%s::%s", a.ProviderType, a.Name)
	}
	return m, nil
}

type allDomainCache struct {
	AccountID         int64
	DomainID          string
	DomainName        string
	RenewalDate       string
	RenewalURL        string
	UsesDNSHEDNS      bool
	DeletedAt         *time.Time
	LastSyncAt        *time.Time
	ProviderUpdatedOn *time.Time
}

func (s *BackupService) listAllDomainCaches(userID int64) ([]allDomainCache, error) {
	rows, err := database.DB.Query(
		`SELECT account_id, domain_id, domain_name, COALESCE(renewal_date, ''), COALESCE(renewal_url, ''), uses_dnshe_dns, deleted_at, last_sync_at, provider_updated_on
		 FROM domain_cache WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var caches []allDomainCache
	for rows.Next() {
		var c allDomainCache
		var usesDNSHE int
		var deletedAt, lastSyncAt, providerUpdatedOn sql.NullTime
		if err := rows.Scan(&c.AccountID, &c.DomainID, &c.DomainName, &c.RenewalDate, &c.RenewalURL, &usesDNSHE, &deletedAt, &lastSyncAt, &providerUpdatedOn); err != nil {
			return nil, err
		}
		c.UsesDNSHEDNS = usesDNSHE != 0
		if deletedAt.Valid {
			c.DeletedAt = &deletedAt.Time
		}
		if lastSyncAt.Valid {
			c.LastSyncAt = &lastSyncAt.Time
		}
		if providerUpdatedOn.Valid {
			c.ProviderUpdatedOn = &providerUpdatedOn.Time
		}
		caches = append(caches, c)
	}
	return caches, rows.Err()
}

func (s *BackupService) getWHOISConfig(userID int64) (*backupWHOISConfig, error) {
	var apiKey string
	err := database.DB.QueryRow("SELECT api_key FROM whois_config WHERE user_id = ?", userID).Scan(&apiKey)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &backupWHOISConfig{APIKey: apiKey}, nil
}

func (s *BackupService) getDNSHEAutoRenew(userID int64) (*backupDNSHEAutoRenew, error) {
	var enabled int
	var daysBefore int
	var lastRunAt sql.NullTime
	err := database.DB.QueryRow("SELECT enabled, days_before, last_run_at FROM dnshe_auto_renew_config WHERE user_id = ?", userID).Scan(&enabled, &daysBefore, &lastRunAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	cfg := &backupDNSHEAutoRenew{Enabled: enabled != 0, DaysBefore: daysBefore}
	if lastRunAt.Valid {
		cfg.LastRunAt = lastRunAt.Time.UTC().Format(time.RFC3339)
	}
	return cfg, nil
}

func (s *BackupService) listCFOptimizeConfigs(userID int64, accountKeyMap map[int64]string) ([]backupCFOptimizeConfig, error) {
	rows, err := database.DB.Query(
		`SELECT account_id, zone_id, zone_name, origin_ip, COALESCE(origin_record_name, ''), COALESCE(origin_record_id, ''), COALESCE(cname_target, ''), COALESCE(cname_record_name, ''), COALESCE(cname_record_id, ''),
		 custom_hostname, COALESCE(custom_hostname_id, ''), COALESCE(status, ''), COALESCE(ssl_status, ''), COALESCE(intermediate_record_name, ''), COALESCE(intermediate_record_id, ''), COALESCE(validation_record_ids, '')
		 FROM cf_optimize WHERE user_id = ? ORDER BY created_at ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []backupCFOptimizeConfig
	for rows.Next() {
		var accountID int64
		var cfg backupCFOptimizeConfig
		if err := rows.Scan(&accountID, &cfg.ZoneID, &cfg.ZoneName, &cfg.OriginIP, &cfg.OriginRecordName, &cfg.OriginRecordID, &cfg.CnameTarget, &cfg.CnameRecordName, &cfg.CnameRecordID, &cfg.CustomHostname, &cfg.CustomHostnameID, &cfg.Status, &cfg.SSLStatus, &cfg.IntermediateRecordName, &cfg.IntermediateRecordID, &cfg.ValidationRecordIDs); err != nil {
			return nil, err
		}
		cfg.AccountKey = accountKeyMap[accountID]
		if cfg.AccountKey == "" {
			cfg.AccountKey = fmt.Sprintf("unknown::%d", accountID)
		}
		configs = append(configs, cfg)
	}
	return configs, rows.Err()
}

func (s *BackupService) importNotificationSetting(tx *sql.Tx, userID, accountID int64, dc backupDomainCache, now time.Time, overwrite bool) error {
	daysBefore := dc.DaysBefore
	if daysBefore <= 0 {
		daysBefore = 30
	}
	enabled := backupBoolToInt(dc.NotifyEnabled)
	lastNotifiedAt := parseNullableTime(dc.LastNotifiedAt)

	var notifID int64
	err := tx.QueryRow(
		"SELECT id FROM notification_settings WHERE user_id = ? AND account_id = ? AND domain_id = ?",
		userID, accountID, dc.DomainID,
	).Scan(&notifID)
	if err == sql.ErrNoRows {
		_, err = tx.Exec(
			`INSERT INTO notification_settings (user_id, domain_id, account_id, days_before, enabled, last_notified_at, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			userID, dc.DomainID, accountID, daysBefore, enabled, lastNotifiedAt, now, now,
		)
	} else if err == nil && overwrite {
		_, err = tx.Exec(
			"UPDATE notification_settings SET days_before=?, enabled=?, last_notified_at=?, updated_at=? WHERE id=?",
			daysBefore, enabled, lastNotifiedAt, now, notifID,
		)
	}
	if err != nil {
		return fmt.Errorf("import notification setting %s: %w", dc.DomainName, err)
	}
	return nil
}

func (s *BackupService) importWHOISConfig(tx *sql.Tx, userID int64, cfg *backupWHOISConfig, overwrite bool) (bool, bool, error) {
	var id int64
	err := tx.QueryRow("SELECT id FROM whois_config WHERE user_id = ?", userID).Scan(&id)
	if err == nil && !overwrite {
		return false, true, nil
	}
	now := time.Now()
	if err == sql.ErrNoRows {
		_, err = tx.Exec("INSERT INTO whois_config (user_id, api_key, created_at, updated_at) VALUES (?, ?, ?, ?)", userID, cfg.APIKey, now, now)
	} else if err == nil {
		_, err = tx.Exec("UPDATE whois_config SET api_key=?, updated_at=? WHERE user_id=?", cfg.APIKey, now, userID)
	}
	if err != nil {
		return false, false, fmt.Errorf("import whois config: %w", err)
	}
	return true, false, nil
}

func (s *BackupService) importDNSHEAutoRenew(tx *sql.Tx, userID int64, cfg *backupDNSHEAutoRenew, overwrite bool) (bool, bool, error) {
	var id int64
	err := tx.QueryRow("SELECT id FROM dnshe_auto_renew_config WHERE user_id = ?", userID).Scan(&id)
	if err == nil && !overwrite {
		return false, true, nil
	}
	enabled := backupBoolToInt(cfg.Enabled)
	daysBefore := cfg.DaysBefore
	if daysBefore <= 0 {
		daysBefore = 7
	}
	lastRunAt := parseNullableTime(cfg.LastRunAt)
	now := time.Now()
	if err == sql.ErrNoRows {
		_, err = tx.Exec("INSERT INTO dnshe_auto_renew_config (user_id, enabled, days_before, last_run_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", userID, enabled, daysBefore, lastRunAt, now, now)
	} else if err == nil {
		_, err = tx.Exec("UPDATE dnshe_auto_renew_config SET enabled=?, days_before=?, last_run_at=?, updated_at=? WHERE user_id=?", enabled, daysBefore, lastRunAt, now, userID)
	}
	if err != nil {
		return false, false, fmt.Errorf("import dnshe auto-renew config: %w", err)
	}
	return true, false, nil
}

func (s *BackupService) importCFOptimizeConfig(tx *sql.Tx, userID, accountID int64, cfg backupCFOptimizeConfig, overwrite bool) (bool, bool, error) {
	var id int64
	err := tx.QueryRow("SELECT id FROM cf_optimize WHERE user_id = ? AND account_id = ? AND custom_hostname = ?", userID, accountID, cfg.CustomHostname).Scan(&id)
	if err == nil && !overwrite {
		return false, true, nil
	}
	now := time.Now()
	if err == sql.ErrNoRows {
		_, err = tx.Exec(
			`INSERT INTO cf_optimize (user_id, account_id, zone_id, zone_name, origin_ip, origin_record_name, origin_record_id, cname_target, cname_record_name, cname_record_id,
			 custom_hostname, custom_hostname_id, status, ssl_status, intermediate_record_name, intermediate_record_id, validation_record_ids, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			userID, accountID, cfg.ZoneID, cfg.ZoneName, cfg.OriginIP, cfg.OriginRecordName, cfg.OriginRecordID, cfg.CnameTarget, cfg.CnameRecordName, cfg.CnameRecordID,
			cfg.CustomHostname, cfg.CustomHostnameID, defaultString(cfg.Status, "pending"), defaultString(cfg.SSLStatus, "pending"), cfg.IntermediateRecordName, cfg.IntermediateRecordID, cfg.ValidationRecordIDs, now, now,
		)
	} else if err == nil {
		_, err = tx.Exec(
			`UPDATE cf_optimize SET zone_id=?, zone_name=?, origin_ip=?, origin_record_name=?, origin_record_id=?, cname_target=?, cname_record_name=?, cname_record_id=?,
			 custom_hostname_id=?, status=?, ssl_status=?, intermediate_record_name=?, intermediate_record_id=?, validation_record_ids=?, updated_at=? WHERE id=?`,
			cfg.ZoneID, cfg.ZoneName, cfg.OriginIP, cfg.OriginRecordName, cfg.OriginRecordID, cfg.CnameTarget, cfg.CnameRecordName, cfg.CnameRecordID,
			cfg.CustomHostnameID, defaultString(cfg.Status, "pending"), defaultString(cfg.SSLStatus, "pending"), cfg.IntermediateRecordName, cfg.IntermediateRecordID, cfg.ValidationRecordIDs, now, id,
		)
	}
	if err != nil {
		return false, false, fmt.Errorf("import cf optimize config %s: %w", cfg.CustomHostname, err)
	}
	return true, false, nil
}

func formatNullTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func parseNullableTime(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
		return t
	}
	return nil
}

func backupBoolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
