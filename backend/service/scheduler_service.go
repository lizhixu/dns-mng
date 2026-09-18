package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"dns-mng/models"
)

type SchedulerService struct {
	notificationService   *NotificationService
	emailService          *EmailService
	schedulerLogService   *SchedulerLogService
	dnsheAutoRenewService *DNSHEAutoRenewService
	ticker                *time.Ticker
	done                  chan bool
}

func NewSchedulerService(notificationService *NotificationService, emailService *EmailService, schedulerLogService *SchedulerLogService, dnsheAutoRenewService *DNSHEAutoRenewService) *SchedulerService {
	return &SchedulerService{
		notificationService:   notificationService,
		emailService:          emailService,
		schedulerLogService:   schedulerLogService,
		dnsheAutoRenewService: dnsheAutoRenewService,
		done:                  make(chan bool),
	}
}

// Start starts the scheduler
func (s *SchedulerService) Start() {
	log.Println("Starting domain expiry notification scheduler...")

	// Schedule to run daily at 9:00 AM
	s.scheduleDaily()
}

// Stop stops the scheduler
func (s *SchedulerService) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.done <- true
	log.Println("Scheduler stopped")
}

// scheduleDaily schedules the task to run daily at 9:00 AM
func (s *SchedulerService) scheduleDaily() {
	now := time.Now()
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())

	// If it's past 9 AM today, schedule for tomorrow
	if now.After(nextRun) {
		nextRun = nextRun.Add(24 * time.Hour)
	}

	// Calculate duration until next run
	duration := nextRun.Sub(now)
	log.Printf("Next notification check scheduled at: %s (in %v)", nextRun.Format("2006-01-02 15:04:05"), duration)

	// Wait until the scheduled time
	time.AfterFunc(duration, func() {
		s.checkExpiringDomains()
		s.runDNSHEAutoRenew()
		// Schedule next run (every 24 hours)
		s.ticker = time.NewTicker(24 * time.Hour)
		go func() {
			for {
				select {
				case <-s.ticker.C:
					s.checkExpiringDomains()
					s.runDNSHEAutoRenew()
				case <-s.done:
					return
				}
			}
		}()
	})
}

// runDNSHEAutoRenew runs the DNSHE auto-renew job for all enabled users.
func (s *SchedulerService) runDNSHEAutoRenew() {
	if s.dnsheAutoRenewService == nil {
		return
	}
	s.dnsheAutoRenewService.RunAll(context.Background(), s.schedulerLogService)
}

// checkExpiringDomains checks for expiring domains and sends notifications
func (s *SchedulerService) checkExpiringDomains(triggerOpt ...string) {
	log.Println("Checking for expiring domains...")

	taskName := "domain_expiry_notification"
	trigger := "scheduled"
	if len(triggerOpt) > 0 && triggerOpt[0] != "" {
		trigger = triggerOpt[0]
	}

	details := map[string]interface{}{
		"trigger": trigger,
	}

	logID, err := s.schedulerLogService.StartTask(taskName, details)
	if err != nil {
		log.Printf("Failed to create scheduler log: %v", err)
	}

	domains, err := s.notificationService.GetExpiringDomains()
	if err != nil {
		log.Printf("Error getting expiring domains: %v", err)
		if logID > 0 {
			s.schedulerLogService.UpdateTaskWithDetails(logID, "error", fmt.Sprintf("获取到期域名失败: %v", err), map[string]interface{}{
				"trigger": trigger,
				"error":   err.Error(),
			})
		}
		return
	}

	triggerTag := "[定时任务]"
	if trigger == "manual" {
		triggerTag = "[手动触发]"
	}

	if len(domains) == 0 {
		log.Println("No domains need notification")
		if logID > 0 {
			s.schedulerLogService.UpdateTaskWithDetails(logID, "success", fmt.Sprintf("%s 检查完成: 暂无需要提醒的到期域名", triggerTag), map[string]interface{}{
				"trigger":       trigger,
				"total_domains": 0,
				"success_count": 0,
				"error_count":   0,
				"message":       "所有已配置域名均未达到到期提醒阈值，或今日已发送过提醒",
			})
		}
		return
	}

	log.Printf("Found %d domain(s) that need notification", len(domains))

	// Group domains by user and collect detailed result per domain
	userDomains := make(map[int64][]string)
	successCount := 0
	errorCount := 0
	errorDetails := make([]map[string]string, 0)
	var successSummary []string
	var failedSummary []string
	var items []map[string]interface{}

	for _, domain := range domains {
		item := map[string]interface{}{
			"user_id":        domain.UserID,
			"account_id":     domain.AccountID,
			"domain":         domain.DomainName,
			"days_remaining": domain.DaysRemaining,
			"renewal_date":   domain.RenewalDate,
			"renewal_url":    domain.RenewalURL,
			"to_email":       domain.ToEmail,
		}

		// Send email notification
		emailErr := s.emailService.SendExpiryNotification(domain.UserID, domain)
		if emailErr != nil {
			log.Printf("Failed to send notification for domain %s: %v", domain.DomainName, emailErr)
			errorCount++
			item["email_status"] = "failed: " + emailErr.Error()
			errorDetails = append(errorDetails, map[string]string{
				"domain": domain.DomainName,
				"error":  emailErr.Error(),
			})
			failedSummary = append(failedSummary, fmt.Sprintf("%s (发送失败: %v)", domain.DomainName, emailErr))
		} else {
			item["email_status"] = "success"
			successCount++
			successSummary = append(successSummary, fmt.Sprintf("%s (剩余%d天)", domain.DomainName, domain.DaysRemaining))

			// Update last notified timestamp
			err = s.notificationService.UpdateLastNotifiedAt(domain.UserID, domain.AccountID, domain.DomainID)
			if err != nil {
				log.Printf("Failed to update last notified timestamp for domain %s: %v", domain.DomainName, err)
			}
		}

		// Sync to message center
		lang := domain.Language
		if lang == "" {
			lang = "zh"
		}
		var title string
		if lang == "en" {
			title = fmt.Sprintf("Domain %s is expiring soon", domain.DomainName)
		} else {
			title = fmt.Sprintf("域名 %s 即将到期", domain.DomainName)
		}
		content := fmt.Sprintf("域名 %s 的到期日为 %s，还有 %d 天到期。", domain.DomainName, domain.RenewalDate, domain.DaysRemaining)
		if domain.RenewalURL != "" {
			content += "\n续费地址：" + domain.RenewalURL
		}
		if lang == "en" {
			content = fmt.Sprintf("Domain %s expires on %s, %d days remaining.", domain.DomainName, domain.RenewalDate, domain.DaysRemaining)
			if domain.RenewalURL != "" {
				content += "\nRenewal URL: " + domain.RenewalURL
			}
		}

		msgErr := s.notificationService.CreateMessage(&models.NotificationMessage{
			UserID:        domain.UserID,
			Type:          "domain_expiry",
			Title:         title,
			Content:       content,
			DomainName:    domain.DomainName,
			DomainID:      domain.DomainID,
			AccountID:     domain.AccountID,
			RenewalDate:   domain.RenewalDate,
			DaysRemaining: domain.DaysRemaining,
			RenewalURL:    domain.RenewalURL,
		})
		if msgErr != nil {
			log.Printf("Failed to create notification message for domain %s: %v", domain.DomainName, msgErr)
			item["message_status"] = "failed: " + msgErr.Error()
		} else {
			item["message_status"] = "synced"
		}

		items = append(items, item)
		userDomains[domain.UserID] = append(userDomains[domain.UserID], domain.DomainName)
		log.Printf("Processed expiry notification for domain: %s (expires in %d days, email=%v, message=%v)",
			domain.DomainName, domain.DaysRemaining, item["email_status"], item["message_status"])
	}

	// Log summary
	for userID, domainNames := range userDomains {
		log.Printf("User %d: Notified about %d domain(s): %v", userID, len(domainNames), domainNames)
	}

	// Update scheduler log
	if logID > 0 {
		completionDetails := map[string]interface{}{
			"trigger":       trigger,
			"total_domains": len(domains),
			"success_count": successCount,
			"error_count":   errorCount,
			"user_domains":  userDomains,
			"items":         items,
		}
		if len(errorDetails) > 0 {
			completionDetails["errors"] = errorDetails
		}

		status := "success"
		var message string
		if errorCount == 0 {
			message = fmt.Sprintf("%s 到期通知完成: 成功通知 %d 个域名 | %s", triggerTag, successCount, strings.Join(successSummary, ", "))
		} else if successCount > 0 {
			status = "partial_success"
			message = fmt.Sprintf("%s 到期通知部分失败: 成功 %d 个, 失败 %d 个 | 成功: %s | 失败: %s", triggerTag, successCount, errorCount, strings.Join(successSummary, ", "), strings.Join(failedSummary, ", "))
		} else {
			status = "error"
			message = fmt.Sprintf("%s 到期通知失败: 全部 %d 个域名通知均失败 | %s", triggerTag, errorCount, strings.Join(failedSummary, ", "))
		}

		s.schedulerLogService.UpdateTaskWithDetails(logID, status, message, completionDetails)
	}
}

// TriggerManualCheck manually triggers domain expiry check (for testing)
func (s *SchedulerService) TriggerManualCheck() {
	go s.checkExpiringDomains("manual")
}
