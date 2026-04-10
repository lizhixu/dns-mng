package service

import (
	"crypto/tls"
	"database/sql"
	"dns-mng/database"
	"dns-mng/models"
	"fmt"
	"net/smtp"
	"time"
)

type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

// GetEmailConfig gets email configuration for a user
func (s *EmailService) GetEmailConfig(userID int64) (*models.EmailConfig, error) {
	var config models.EmailConfig
	var enabled int

	err := database.DB.QueryRow(
		`SELECT id, user_id, smtp_host, smtp_port, smtp_username, smtp_password, from_email, from_name, to_email, enabled, created_at, updated_at
		 FROM email_config WHERE user_id = ?`,
		userID,
	).Scan(&config.ID, &config.UserID, &config.SMTPHost, &config.SMTPPort,
		&config.SMTPUsername, &config.SMTPPassword, &config.FromEmail, &config.FromName,
		&config.ToEmail, &enabled, &config.CreatedAt, &config.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	config.Enabled = enabled == 1
	// Don't return password in response
	config.SMTPPassword = ""

	return &config, nil
}

// UpsertEmailConfig creates or updates email configuration
func (s *EmailService) UpsertEmailConfig(userID int64, req *models.UpdateEmailConfigRequest) (*models.EmailConfig, error) {
	now := time.Now()
	enabled := 0
	if req.Enabled {
		enabled = 1
	}

	// Check if config exists
	var existingID int64
	err := database.DB.QueryRow(`SELECT id FROM email_config WHERE user_id = ?`, userID).Scan(&existingID)

	switch err {
	case sql.ErrNoRows:
		// Insert new config
		_, err = database.DB.Exec(
			`INSERT INTO email_config (user_id, smtp_host, smtp_port, smtp_username, smtp_password, from_email, from_name, to_email, enabled, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			userID, req.SMTPHost, req.SMTPPort, req.SMTPUsername, req.SMTPPassword,
			req.FromEmail, req.FromName, req.ToEmail, enabled, now, now,
		)
	case nil:
		// Update existing config
		if req.SMTPPassword != "" {
			// Update with new password
			_, err = database.DB.Exec(
				`UPDATE email_config SET smtp_host = ?, smtp_port = ?, smtp_username = ?, smtp_password = ?, 
				 from_email = ?, from_name = ?, to_email = ?, enabled = ?, updated_at = ? WHERE user_id = ?`,
				req.SMTPHost, req.SMTPPort, req.SMTPUsername, req.SMTPPassword,
				req.FromEmail, req.FromName, req.ToEmail, enabled, now, userID,
			)
		} else {
			// Update without changing password
			_, err = database.DB.Exec(
				`UPDATE email_config SET smtp_host = ?, smtp_port = ?, smtp_username = ?, 
				 from_email = ?, from_name = ?, to_email = ?, enabled = ?, updated_at = ? WHERE user_id = ?`,
				req.SMTPHost, req.SMTPPort, req.SMTPUsername,
				req.FromEmail, req.FromName, req.ToEmail, enabled, now, userID,
			)
		}
	}

	if err != nil {
		return nil, err
	}

	return s.GetEmailConfig(userID)
}

// SendEmail sends an email using the user's configuration
func (s *EmailService) SendEmail(userID int64, to, subject, body string) error {
	// Get email config with password
	var config models.EmailConfig
	var enabled int

	err := database.DB.QueryRow(
		`SELECT smtp_host, smtp_port, smtp_username, smtp_password, from_email, from_name, enabled
		 FROM email_config WHERE user_id = ?`,
		userID,
	).Scan(&config.SMTPHost, &config.SMTPPort, &config.SMTPUsername, &config.SMTPPassword,
		&config.FromEmail, &config.FromName, &enabled)

	if err != nil {
		return fmt.Errorf("email configuration not found")
	}

	if enabled == 0 {
		return fmt.Errorf("email notifications are disabled")
	}

	// Build email message
	from := config.FromEmail
	if config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", config.FromName, config.FromEmail)
	}

	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	addr := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)

	// Try standard smtp.SendMail first (most reliable for most providers)
	auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPHost)
	err = smtp.SendMail(addr, auth, config.FromEmail, []string{to}, []byte(message))
	if err == nil {
		return nil
	}

	// If standard SendMail fails, try custom TLS approach
	fmt.Printf("Standard SendMail failed: %v, trying TLS approach...\n", err)

	tlsConfig := &tls.Config{
		ServerName:         config.SMTPHost,
		InsecureSkipVerify: false,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS connection failed: %v", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, config.SMTPHost)
	if err != nil {
		return fmt.Errorf("SMTP client create failed: %v", err)
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %v", err)
	}

	if err = client.Mail(config.FromEmail); err != nil {
		return fmt.Errorf("MAIL FROM failed: %v", err)
	}

	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO failed: %v", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %v", err)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("Write message failed: %v", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("Close writer failed: %v", err)
	}

	return client.Quit()
}

// SendExpiryNotification sends expiry notification email
func (s *EmailService) SendExpiryNotification(userID int64, domain models.ExpiringDomain) error {
	subject := fmt.Sprintf("域名续费提醒：%s 将在 %d 天后到期", domain.DomainName, domain.DaysRemaining)

	renewalLink := ""
	if domain.RenewalURL != "" {
		renewalLink = fmt.Sprintf(`<p><a href="%s" style="display: inline-block; padding: 10px 20px; background-color: #3b82f6; color: white; text-decoration: none; border-radius: 5px;">立即续费</a></p>`, domain.RenewalURL)
	}

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        .content { background-color: #ffffff; padding: 20px; border: 1px solid #e5e7eb; border-radius: 5px; }
        .warning { color: #dc2626; font-weight: bold; }
        .info { background-color: #f0f9ff; padding: 15px; border-left: 4px solid #3b82f6; margin: 15px 0; }
        .footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #e5e7eb; font-size: 12px; color: #6b7280; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2 style="margin: 0; color: #1f2937;">🔔 域名续费提醒</h2>
        </div>
        <div class="content">
            <p>您好，</p>
            <p>您的域名即将到期，请及时续费以避免服务中断。</p>
            <div class="info">
                <p style="margin: 5px 0;"><strong>域名：</strong>%s</p>
                <p style="margin: 5px 0;"><strong>到期日期：</strong>%s</p>
                <p style="margin: 5px 0;"><strong class="warning">剩余天数：%d 天</strong></p>
            </div>
            %s
            <p>请尽快完成续费，以确保您的域名服务不受影响。</p>
        </div>
        <div class="footer">
            <p>此邮件由 DNS Manager 系统自动发送，请勿直接回复。</p>
        </div>
    </div>
</body>
</html>
`, domain.DomainName, domain.RenewalDate, domain.DaysRemaining, renewalLink)

	err := s.SendEmail(userID, domain.ToEmail, subject, body)
	if err != nil {
		return fmt.Errorf("failed to send email to %s: %v", domain.ToEmail, err)
	}
	return nil
}

// TestEmailConfig tests email configuration by sending a test email
func (s *EmailService) TestEmailConfig(userID int64) error {
	// Get email config with to_email
	var toEmail string
	err := database.DB.QueryRow(
		`SELECT to_email FROM email_config WHERE user_id = ?`,
		userID,
	).Scan(&toEmail)

	if err != nil {
		return fmt.Errorf("email configuration not found")
	}

	subject := "DNS Manager - 邮件配置测试"
	body := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #3b82f6;">✅ 邮件配置测试成功</h2>
        <p>恭喜！您的邮件配置已正确设置。</p>
        <p>DNS Manager 现在可以向您发送域名到期提醒通知。</p>
        <hr style="border: none; border-top: 1px solid #e5e7eb; margin: 20px 0;">
        <p style="font-size: 12px; color: #6b7280;">此邮件由 DNS Manager 系统发送。</p>
    </div>
</body>
</html>
`
	return s.SendEmail(userID, toEmail, subject, body)
}
