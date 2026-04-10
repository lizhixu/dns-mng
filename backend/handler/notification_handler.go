package handler

import (
	"net/http"

	"dns-mng/database"
	"dns-mng/middleware"
	"dns-mng/models"
	"dns-mng/service"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notificationService *service.NotificationService
	emailService        *service.EmailService
	logService          *service.LogService
}

func NewNotificationHandler(notificationService *service.NotificationService, emailService *service.EmailService, logService *service.LogService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		emailService:        emailService,
		logService:          logService,
	}
}

// GetNotificationSetting gets notification setting for a domain
func (h *NotificationHandler) GetNotificationSetting(c *gin.Context) {
	userID := middleware.GetUserID(c)
	accountID, err := middleware.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}
	domainID := c.Param("domainId")

	setting, err := h.notificationService.GetNotificationSetting(userID, accountID, domainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if setting == nil {
		// Return default settings
		c.JSON(http.StatusOK, gin.H{
			"days_before": 30,
			"enabled":     false,
		})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// UpdateNotificationSetting updates notification setting for a domain
func (h *NotificationHandler) UpdateNotificationSetting(c *gin.Context) {
	userID := middleware.GetUserID(c)
	accountID, err := middleware.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}
	domainID := c.Param("domainId")

	var req models.UpdateNotificationSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setting, err := h.notificationService.UpsertNotificationSetting(userID, accountID, domainID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get domain name from cache for logging
	domainName := domainID
	var cachedDomain models.DomainCache
	database.DB.QueryRow(`
		SELECT domain_name 
		FROM domain_cache 
		WHERE user_id = ? AND account_id = ? AND domain_id = ?
	`, userID, accountID, domainID).Scan(&cachedDomain.DomainName)
	if cachedDomain.DomainName != "" {
		domainName = cachedDomain.DomainName
	}

	// Log operation
	h.logService.CreateLog(userID, "update", "notification_setting", domainID, map[string]interface{}{
		"domain":      domainName,
		"days_before": req.DaysBefore,
		"enabled":     req.Enabled,
	}, c.ClientIP())

	c.JSON(http.StatusOK, setting)
}

// GetAllNotificationSettings gets all notification settings for a user
func (h *NotificationHandler) GetAllNotificationSettings(c *gin.Context) {
	userID := middleware.GetUserID(c)

	settings, err := h.notificationService.GetAllNotificationSettings(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if settings == nil {
		settings = []models.NotificationSetting{}
	}

	c.JSON(http.StatusOK, settings)
}

// GetEmailConfig gets email configuration
func (h *NotificationHandler) GetEmailConfig(c *gin.Context) {
	userID := middleware.GetUserID(c)

	config, err := h.emailService.GetEmailConfig(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if config == nil {
		c.JSON(http.StatusOK, gin.H{"configured": false})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateEmailConfig updates email configuration
func (h *NotificationHandler) UpdateEmailConfig(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req models.UpdateEmailConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := h.emailService.UpsertEmailConfig(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log operation
	h.logService.CreateLog(userID, "update", "email_config", "", map[string]interface{}{
		"smtp_host": req.SMTPHost,
		"smtp_port": req.SMTPPort,
		"enabled":   req.Enabled,
	}, c.ClientIP())

	c.JSON(http.StatusOK, config)
}

// TestEmailConfig tests email configuration
func (h *NotificationHandler) TestEmailConfig(c *gin.Context) {
	userID := middleware.GetUserID(c)

	err := h.emailService.TestEmailConfig(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log operation
	h.logService.CreateLog(userID, "test", "email_config", "", map[string]interface{}{}, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "test email sent successfully"})
}
