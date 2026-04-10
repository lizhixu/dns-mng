package service

import (
	"log"
	"time"
)

type SchedulerService struct {
	notificationService *NotificationService
	emailService        *EmailService
	schedulerLogService *SchedulerLogService
	ticker              *time.Ticker
	done                chan bool
}

func NewSchedulerService(notificationService *NotificationService, emailService *EmailService, schedulerLogService *SchedulerLogService) *SchedulerService {
	return &SchedulerService{
		notificationService: notificationService,
		emailService:        emailService,
		schedulerLogService: schedulerLogService,
		done:                make(chan bool),
	}
}

// Start starts the scheduler
func (s *SchedulerService) Start() {
	log.Println("Starting domain expiry notification scheduler...")

	// Run immediately on start
	go s.checkExpiringDomains()

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
		// Schedule next run (every 24 hours)
		s.ticker = time.NewTicker(24 * time.Hour)
		go func() {
			for {
				select {
				case <-s.ticker.C:
					s.checkExpiringDomains()
				case <-s.done:
					return
				}
			}
		}()
	})
}

// checkExpiringDomains checks for expiring domains and sends notifications
func (s *SchedulerService) checkExpiringDomains() {
	log.Println("Checking for expiring domains...")

	taskName := "domain_expiry_notification"

	details := map[string]interface{}{
		"trigger": "scheduled",
	}

	logID, err := s.schedulerLogService.StartTask(taskName, details)
	if err != nil {
		log.Printf("Failed to create scheduler log: %v", err)
	}

	domains, err := s.notificationService.GetExpiringDomains()
	if err != nil {
		log.Printf("Error getting expiring domains: %v", err)
		if logID > 0 {
			s.schedulerLogService.UpdateTask(logID, "error", err.Error())
		}
		return
	}

	if len(domains) == 0 {
		log.Println("No domains need notification")
		if logID > 0 {
			s.schedulerLogService.UpdateTask(logID, "success", "No domains need notification")
		}
		return
	}

	log.Printf("Found %d domain(s) that need notification", len(domains))

	// Group domains by user
	userDomains := make(map[int64][]string)
	successCount := 0
	errorCount := 0
	errorDetails := make([]map[string]string, 0)

	for _, domain := range domains {
		// Send email notification
		err := s.emailService.SendExpiryNotification(domain.UserID, domain)
		if err != nil {
			log.Printf("Failed to send notification for domain %s: %v", domain.DomainName, err)
			errorCount++
			errorDetails = append(errorDetails, map[string]string{
				"domain": domain.DomainName,
				"error":  err.Error(),
			})
			continue
		}

		// Update last notified timestamp
		err = s.notificationService.UpdateLastNotifiedAt(domain.UserID, domain.AccountID, domain.DomainID)
		if err != nil {
			log.Printf("Failed to update last notified timestamp for domain %s: %v", domain.DomainName, err)
		}

		userDomains[domain.UserID] = append(userDomains[domain.UserID], domain.DomainName)
		log.Printf("Sent notification for domain: %s (expires in %d days)", domain.DomainName, domain.DaysRemaining)
		successCount++
	}

	// Log summary
	for userID, domainNames := range userDomains {
		log.Printf("User %d: Notified about %d domain(s): %v", userID, len(domainNames), domainNames)
	}

	// Update scheduler log
	if logID > 0 {
		completionDetails := map[string]interface{}{
			"total_domains": len(domains),
			"success_count": successCount,
			"error_count":   errorCount,
			"user_domains":  userDomains,
		}
		if len(errorDetails) > 0 {
			completionDetails["errors"] = errorDetails
		}

		status := "success"
		message := "Notifications sent successfully"
		if errorCount > 0 {
			status = "partial_success"
			message = "Some notifications failed to send"
		}

		s.schedulerLogService.UpdateTask(logID, status, message)
	}
}

// TriggerManualCheck manually triggers domain expiry check (for testing)
func (s *SchedulerService) TriggerManualCheck() {
	go s.checkExpiringDomains()
}
