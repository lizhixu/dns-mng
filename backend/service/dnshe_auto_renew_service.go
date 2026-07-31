package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"dns-mng/database"
	"dns-mng/models"
)

// DNSHEAutoRenewService manages the DNSHE global auto-renew configuration and
// the scheduled renewal job.
type DNSHEAutoRenewService struct {
	dnsheService *DNSHEService
}

func NewDNSHEAutoRenewService(dnsheService *DNSHEService) *DNSHEAutoRenewService {
	return &DNSHEAutoRenewService{dnsheService: dnsheService}
}

// GetConfig returns the auto-renew config for a user (creates a default row if missing).
func (s *DNSHEAutoRenewService) GetConfig(userID int64) (*models.DNSHEAutoRenewConfig, error) {
	cfg, err := s.getConfig(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return defaults without persisting
			return &models.DNSHEAutoRenewConfig{
				UserID:     userID,
				Enabled:    false,
				DaysBefore: 7,
			}, nil
		}
		return nil, err
	}
	return cfg, nil
}

func (s *DNSHEAutoRenewService) getConfig(userID int64) (*models.DNSHEAutoRenewConfig, error) {
	var c models.DNSHEAutoRenewConfig
	var enabled int
	var lastRunAt sql.NullTime
	err := database.DB.QueryRow(
		`SELECT user_id, enabled, days_before, last_run_at, created_at, updated_at
		 FROM dnshe_auto_renew_config WHERE user_id = ?`,
		userID,
	).Scan(&c.UserID, &enabled, &c.DaysBefore, &lastRunAt, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	c.Enabled = enabled == 1
	if lastRunAt.Valid {
		c.LastRunAt = &lastRunAt.Time
	}
	return &c, nil
}

// UpdateConfig updates the auto-renew config for a user (upsert).
func (s *DNSHEAutoRenewService) UpdateConfig(userID int64, req *models.UpdateDNSHEAutoRenewConfigRequest) (*models.DNSHEAutoRenewConfig, error) {
	existing, err := s.getConfig(userID)
	now := time.Now()

	enabled := false
	daysBefore := 7
	if err == nil && existing != nil {
		enabled = existing.Enabled
		daysBefore = existing.DaysBefore
	}
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	if req.DaysBefore != nil {
		daysBefore = *req.DaysBefore
		if *req.DaysBefore < 1 {
			daysBefore = 1
		}
	}

	enabledInt := 0
	if enabled {
		enabledInt = 1
	}

	_, err = database.DB.Exec(
		`INSERT INTO dnshe_auto_renew_config (user_id, enabled, days_before, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(user_id) DO UPDATE SET enabled = excluded.enabled, days_before = excluded.days_before, updated_at = excluded.updated_at`,
		userID, enabledInt, daysBefore, now, now,
	)
	if err != nil {
		return nil, err
	}

	return s.getConfig(userID)
}

// TriggerRunForUser runs the auto-renew job once for a specific user (manual trigger).
func (s *DNSHEAutoRenewService) TriggerRunForUser(ctx context.Context, userID int64, schedulerLogService *SchedulerLogService) (*AutoRenewRunResult, error) {
	cfg, err := s.GetConfig(userID)
	if err != nil {
		return nil, err
	}

	var logID int64
	if schedulerLogService != nil {
		logID, _ = schedulerLogService.StartTask("dnshe_auto_renew", map[string]interface{}{"trigger": "manual", "user_id": userID})
	}

	result := s.runForUser(ctx, userID, cfg.DaysBefore)

	if logID > 0 {
		status := "success"
		if result.Failed > 0 && result.Renewed > 0 {
			status = "partial_success"
		} else if result.Failed > 0 && result.Renewed == 0 {
			status = "error"
		}
		message := fmt.Sprintf("手动触发: 检查 %d 个, 续期 %d 个, 失败 %d 个", result.Checked, result.Renewed, result.Failed)
		if len(result.RenewedDomains) > 0 {
			message += fmt.Sprintf(" | 成功: %s", strings.Join(result.RenewedDomains, ", "))
		}
		if len(result.FailedDomains) > 0 {
			message += fmt.Sprintf(" | 失败: %s", strings.Join(result.FailedDomains, ", "))
		}
		schedulerLogService.UpdateTask(logID, status, message)
	}

	return result, nil
}

// AutoRenewRunResult summarizes a single run.
type AutoRenewRunResult struct {
	Checked   int      `json:"checked"`
	Renewed   int      `json:"renewed"`
	Failed    int      `json:"failed"`
	RenewedDomains  []string `json:"renewed_domains"`
	FailedDomains   []string `json:"failed_domains"`
}

// runForUser iterates all DNSHE accounts/domains for a user and renews domains
// expiring within daysBefore days. Renewal does NOT check quota (per requirement).
func (s *DNSHEAutoRenewService) runForUser(ctx context.Context, userID int64, daysBefore int) *AutoRenewRunResult {
	result := &AutoRenewRunResult{RenewedDomains: []string{}, FailedDomains: []string{}}

	accounts, err := s.dnsheService.ListAccounts(userID)
	if err != nil || len(accounts) == 0 {
		return result
	}

	now := time.Now()
	threshold := now.AddDate(0, 0, daysBefore)

	for _, account := range accounts {
		// List domains via the DNSHE provider to inspect expiry.
		domains, err := s.dnsheService.domainListForAutoRenew(ctx, account)
		if err != nil {
			log.Printf("DNSHE auto-renew: failed to list domains for account %d (%s): %v", account.ID, account.Name, err)
			continue
		}

		for _, d := range domains {
			result.Checked++
			// Skip permanent / never-expires domains.
			if d.RenewalDate == "" || d.RenewalDate == "permanent" {
				continue
			}
			expiresAt, perr := time.Parse("2006-01-02", d.RenewalDate)
			if perr != nil {
				continue
			}
			if expiresAt.After(threshold) {
				continue
			}

			subdomainID, perr := strconv.Atoi(d.ID)
			if perr != nil {
				result.Failed++
				result.FailedDomains = append(result.FailedDomains, d.Name)
				continue
			}

			resp, rerr := s.dnsheService.RenewSubdomain(ctx, userID, account.ID, subdomainID)
			if rerr != nil {
				log.Printf("DNSHE auto-renew: failed to renew %s (account %d): %v", d.Name, account.ID, rerr)
				result.Failed++
				result.FailedDomains = append(result.FailedDomains, d.Name)
				continue
			}
			log.Printf("DNSHE auto-renew: renewed %s (charged %.2f, new expiry %s)", d.Name, resp.ChargedAmount, resp.NewExpiresAt)
			result.Renewed++
			result.RenewedDomains = append(result.RenewedDomains, d.Name)
		}
	}

	return result
}

// UpdateLastRunAt records the last run timestamp for a user.
func (s *DNSHEAutoRenewService) UpdateLastRunAt(userID int64) {
	now := time.Now()
	_, _ = database.DB.Exec(
		`UPDATE dnshe_auto_renew_config SET last_run_at = ?, updated_at = ? WHERE user_id = ?`,
		now, now, userID,
	)
}

// RunAll iterates all users with auto-renew enabled and runs the job. Used by the scheduler.
func (s *DNSHEAutoRenewService) RunAll(ctx context.Context, schedulerLogService *SchedulerLogService) {
	logID, _ := schedulerLogService.StartTask("dnshe_auto_renew", map[string]interface{}{"trigger": "scheduled"})

	rows, err := database.DB.Query(
		`SELECT user_id, days_before FROM dnshe_auto_renew_config WHERE enabled = 1`,
	)
	if err != nil {
		log.Printf("DNSHE auto-renew: failed to query enabled configs: %v", err)
		if logID > 0 {
			schedulerLogService.UpdateTask(logID, "error", err.Error())
		}
		return
	}
	type userCfg struct {
		userID     int64
		daysBefore int
	}
	var configs []userCfg
	for rows.Next() {
		var uc userCfg
		if err := rows.Scan(&uc.userID, &uc.daysBefore); err != nil {
			continue
		}
		configs = append(configs, uc)
	}
	rows.Close()

	totalRenewed, totalFailed := 0, 0
	var allRenewed, allFailed []string
	for _, uc := range configs {
		res := s.runForUser(ctx, uc.userID, uc.daysBefore)
		totalRenewed += res.Renewed
		totalFailed += res.Failed
		allRenewed = append(allRenewed, res.RenewedDomains...)
		allFailed = append(allFailed, res.FailedDomains...)
		s.UpdateLastRunAt(uc.userID)
	}

	status := "success"
	message := fmt.Sprintf("定时续期完成: 检查完成, 续期 %d 个, 失败 %d 个", totalRenewed, totalFailed)
	if len(allRenewed) > 0 {
		message += fmt.Sprintf(" | 成功: %s", strings.Join(allRenewed, ", "))
	}
	if len(allFailed) > 0 {
		message += fmt.Sprintf(" | 失败: %s", strings.Join(allFailed, ", "))
	}
	if totalFailed > 0 && totalRenewed == 0 {
		status = "error"
	} else if totalFailed > 0 {
		status = "partial_success"
	}
	if logID > 0 {
		schedulerLogService.UpdateTask(logID, status, message)
	}
	log.Printf("DNSHE auto-renew: %s", message)
}