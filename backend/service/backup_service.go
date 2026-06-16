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
	Accounts            []backupAccount       `json:"accounts"`
	DomainCaches        []backupDomainCache   `json:"domain_caches"`
	DDNSToken           *backupDDNSToken      `json:"ddns_token"`
	EmailConfig         *backupEmailConfig    `json:"email_config"`
}

type backupAccount struct {
	Name         string `json:"name"`
	ProviderType string `json:"provider_type"`
	APIKey       string `json:"api_key"`
}

type backupDomainCache struct {
	AccountKey   string `json:"account_key"` // "provider_type::name"
	DomainID     string `json:"domain_id"`
	DomainName   string `json:"domain_name"`
	RenewalDate  string `json:"renewal_date,omitempty"`
	RenewalURL   string `json:"renewal_url,omitempty"`
	DaysBefore   int    `json:"days_before,omitempty"`    // 到期前提醒天数
	NotifyEnabled bool  `json:"notify_enabled,omitempty"` // 是否启用到期提醒
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

type backupFile struct {
	Version    int       `json:"version"`
	ExportedAt string    `json:"exported_at"`
	Encrypted  bool      `json:"encrypted"`
	Data       backupData `json:"data"`
}

// ─── 导入结果 ────────────────────────────────────────────────────

type ImportResult struct {
	AccountsImported     int  `json:"accounts_imported"`
	AccountsSkipped      int  `json:"accounts_skipped"`
	DomainCachesImported int  `json:"domain_caches_imported"`
	DomainCachesSkipped  int  `json:"domain_caches_skipped"`
	DDNSTokenImported    bool `json:"ddns_token_imported"`
	DDNSTokenSkipped     bool `json:"ddns_token_skipped"`
	EmailConfigImported  bool `json:"email_config_imported"`
	EmailConfigSkipped   bool `json:"email_config_skipped"`
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
		log.Printf("Warning: build account key map: %v", err)
	}
	caches, err := s.domainCacheService.GetCacheByUser(userID)
	if err != nil {
		log.Printf("Warning: export domain caches: %v", err)
	}
	// 构建通知设置查找表: "accountID:domainID" → (days_before, enabled)
	notifMap := make(map[string]struct{ daysBefore int; enabled bool })
	notifications, err := s.notificationService.GetAllNotificationSettings(userID)
	if err != nil {
		log.Printf("Warning: export notification settings: %v", err)
	}
	for _, ns := range notifications {
		notifMap[fmt.Sprintf("%d:%s", ns.AccountID, ns.DomainID)] = struct{ daysBefore int; enabled bool }{ns.DaysBefore, ns.Enabled}
	}
	for _, c := range caches {
		accountKey := accountKeyMap[c.AccountID]
		if accountKey == "" {
			accountKey = fmt.Sprintf("unknown::%d", c.AccountID)
		}
		entry := backupDomainCache{
			AccountKey:  accountKey,
			DomainID:    c.DomainID,
			DomainName:  c.DomainName,
			RenewalDate: c.RenewalDate,
			RenewalURL:  c.RenewalURL,
		}
		if notif, ok := notifMap[fmt.Sprintf("%d:%s", c.AccountID, c.DomainID)]; ok {
			entry.DaysBefore = notif.daysBefore
			entry.NotifyEnabled = notif.enabled
		}
		data.DomainCaches = append(data.DomainCaches, entry)
	}

	// 3. DDNS Token
	token, err := s.ddnsTokenService.GetToken(userID)
	if err != nil {
		log.Printf("Warning: export ddns token: %v", err)
	}
	if token != nil {
		data.DDNSToken = &backupDDNSToken{
			Token:   token.Token,
			Enabled: token.Enabled,
		}
	}

	// 3. 邮件配置（含密码）
	emailCfg, err := s.emailService.getEmailConfigWithPassword(userID)
	if err != nil {
		log.Printf("Warning: export email config: %v", err)
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

	// 构建文件
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

	// 可选加密
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
	// 解密
	plainJSON, err := DecryptBackup(fileBytes, password)
	if err != nil {
		return nil, err
	}

	// 解析
	var file backupFile
	if err := json.Unmarshal(plainJSON, &file); err != nil {
		return nil, fmt.Errorf("解析备份文件失败: %w", err)
	}
	if file.Version != 1 {
		return nil, fmt.Errorf("不支持的备份版本: %d", file.Version)
	}

	result := &ImportResult{}

	// 使用事务保证原子性
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 临时把 DB 替换为 tx 以复用 service 方法 —— 但现有 service 直接用 database.DB。
	// 为简洁起见，直接在 backup_service 中操作 tx，不修改现有 service。

	// 1. 导入账户
	accountKeyToID := make(map[string]int64) // "provider_type::name" → 新 accountID
	for _, acc := range file.Data.Accounts {
		key := fmt.Sprintf("%s::%s", acc.ProviderType, acc.Name)
		existingID, _ := s.findAccountByKey(userID, key, tx)

		if existingID > 0 && !overwrite {
			result.AccountsSkipped++
			accountKeyToID[key] = existingID
			continue
		}

		if existingID > 0 && overwrite {
			// 更新
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

		// 新建
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

	// 2. 导入域名缓存（续费信息）
	for _, dc := range file.Data.DomainCaches {
		accountID, ok := accountKeyToID[dc.AccountKey]
		if !ok {
			foundID, _ := s.findAccountByKey(userID, dc.AccountKey, tx)
			if foundID > 0 {
				accountID = foundID
			} else {
				log.Printf("Warning: skip domain cache for unknown account key: %s", dc.AccountKey)
				result.DomainCachesSkipped++
				continue
			}
		}

		// 检查是否已存在
		var existingID int64
		err := tx.QueryRow(
			"SELECT id FROM domain_cache WHERE user_id = ? AND account_id = ? AND domain_id = ? AND deleted_at IS NULL",
			userID, accountID, dc.DomainID,
		).Scan(&existingID)

		now := time.Now()
		if err == sql.ErrNoRows {
			// 新建
			_, err = tx.Exec(
				`INSERT INTO domain_cache (user_id, account_id, domain_id, domain_name, renewal_date, renewal_url, created_at, updated_at)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				userID, accountID, dc.DomainID, dc.DomainName, dc.RenewalDate, dc.RenewalURL, now, now,
			)
		} else if err == nil {
			if overwrite {
				_, err = tx.Exec(
					"UPDATE domain_cache SET domain_name=?, renewal_date=?, renewal_url=?, updated_at=? WHERE id=?",
					dc.DomainName, dc.RenewalDate, dc.RenewalURL, now, existingID,
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

		// 同步导入通知设置（days_before + notify_enabled）
		if dc.DaysBefore > 0 || dc.NotifyEnabled {
			notifEnabled := boolToInt(dc.NotifyEnabled)
			var notifID int64
			err := tx.QueryRow(
				"SELECT id FROM notification_settings WHERE user_id = ? AND account_id = ? AND domain_id = ?",
				userID, accountID, dc.DomainID,
			).Scan(&notifID)
			if err == sql.ErrNoRows {
				_, _ = tx.Exec(
					`INSERT INTO notification_settings (user_id, domain_id, account_id, days_before, enabled, created_at, updated_at)
					 VALUES (?, ?, ?, ?, ?, ?, ?)`,
					userID, dc.DomainID, accountID, dc.DaysBefore, notifEnabled, now, now,
				)
			} else if err == nil && overwrite {
				_, _ = tx.Exec(
					"UPDATE notification_settings SET days_before=?, enabled=?, updated_at=? WHERE id=?",
					dc.DaysBefore, notifEnabled, now, notifID,
				)
			}
		}
	}

	// 3. 导入 DDNS Token
	if file.Data.DDNSToken != nil {
		existing, err := s.findDDNSToken(userID, tx)
		if err != nil {
			return nil, fmt.Errorf("check ddns token: %w", err)
		}

		if existing != "" && !overwrite {
			result.DDNSTokenSkipped = true
		} else {
			enabled := boolToInt(file.Data.DDNSToken.Enabled)
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

	// 3. 导入邮件配置
	if file.Data.EmailConfig != nil {
		ec := file.Data.EmailConfig
		existing, err := s.findEmailConfig(userID, tx)
		if err != nil {
			return nil, fmt.Errorf("check email config: %w", err)
		}

		if existing && !overwrite {
			result.EmailConfigSkipped = true
		} else {
			enabled := boolToInt(ec.Enabled)
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

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return result, nil
}

// ─── 内部辅助方法 ──────────────────────────────────────────────

func (s *BackupService) findAccountByKey(userID int64, key string, tx *sql.Tx) (int64, error) {
	// key = "provider_type::name"
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
