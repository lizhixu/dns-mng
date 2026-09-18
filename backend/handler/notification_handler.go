package handler

import (
	"database/sql"
	"fmt"
	"net/http"

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

	c.JSON(http.StatusOK, gin.H{"message": "test email sent successfully"})
}

// ListMessages lists notification messages for the current user
func (h *NotificationHandler) ListMessages(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := parseIntQuery(c, "page", 1)
	pageSize, _ := parseIntQuery(c, "per_page", 20)
	unreadOnly := c.Query("unread_only") == "1"

	messages, total, err := h.notificationService.ListMessages(userID, page, pageSize, unreadOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"total":    total,
		"page":     page,
		"per_page": pageSize,
	})
}

// MarkMessageRead marks a single message as read
func (h *NotificationHandler) MarkMessageRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	messageID, err := parseIntParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message id"})
		return
	}

	if err := h.notificationService.MarkMessageRead(userID, messageID); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "marked as read"})
}

// MarkAllMessagesRead marks all messages as read for the current user
func (h *NotificationHandler) MarkAllMessagesRead(c *gin.Context) {
	userID := middleware.GetUserID(c)

	if err := h.notificationService.MarkAllMessagesRead(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "all messages marked as read"})
}

// DeleteMessage deletes a notification message
func (h *NotificationHandler) DeleteMessage(c *gin.Context) {
	userID := middleware.GetUserID(c)
	messageID, err := parseIntParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message id"})
		return
	}

	if err := h.notificationService.DeleteMessage(userID, messageID); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// GetUnreadCount gets unread message count for the current user
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := middleware.GetUserID(c)

	count, err := h.notificationService.GetUnreadCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func parseIntQuery(c *gin.Context, key string, def int) (int, error) {
	val := c.Query(key)
	if val == "" {
		return def, nil
	}
	var n int
	_, err := fmt.Sscanf(val, "%d", &n)
	return n, err
}

func parseIntParam(c *gin.Context, key string) (int64, error) {
	val := c.Param(key)
	var n int64
	_, err := fmt.Sscanf(val, "%d", &n)
	return n, err
}
