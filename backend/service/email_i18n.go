package service

import "fmt"

// EmailTranslations holds all translatable strings for email templates
type EmailTranslations struct {
	// Expiry notification
	ExpirySubject    func(domain string, days int) string
	ExpiryHeader     string
	ExpiryGreeting   string
	ExpiryMessage    string
	DomainLabel      string
	ExpiryDateLabel  string
	DaysRemainingFmt func(days int) string
	RenewNowBtn      string
	RenewSoonMsg     string
	Footer           string
	// Test email
	TestSubject      string
	TestSuccessTitle string
	TestCongrats     string
	TestConfirmation string
	TestFooter       string
}

var emailTranslations = map[string]EmailTranslations{
	"zh": {
		ExpirySubject: func(domain string, days int) string {
			return fmt.Sprintf("域名续费提醒：%s 将在 %d 天后到期", domain, days)
		},
		ExpiryHeader:    "域名续费提醒",
		ExpiryGreeting:  "您好，",
		ExpiryMessage:   "您的域名即将到期，请及时续费以避免服务中断。",
		DomainLabel:     "域名：",
		ExpiryDateLabel: "到期日期：",
		DaysRemainingFmt: func(days int) string {
			return fmt.Sprintf("剩余天数：%d 天", days)
		},
		RenewNowBtn:  "立即续费",
		RenewSoonMsg: "请尽快完成续费，以确保您的域名服务不受影响。",
		Footer:       "此邮件由 DNS Manager 系统自动发送，请勿直接回复。",
		// Test email
		TestSubject:      "DNS Manager - 邮件配置测试",
		TestSuccessTitle: "邮件配置测试成功",
		TestCongrats:     "恭喜！您的邮件配置已正确设置。",
		TestConfirmation: "DNS Manager 现在可以向您发送域名到期提醒通知。",
		TestFooter:       "此邮件由 DNS Manager 系统发送。",
	},
	"en": {
		ExpirySubject: func(domain string, days int) string {
			return fmt.Sprintf("Domain Renewal Reminder: %s expires in %d days", domain, days)
		},
		ExpiryHeader:    "Domain Renewal Reminder",
		ExpiryGreeting:  "Hello,",
		ExpiryMessage:   "Your domain is about to expire. Please renew it in time to avoid service interruption.",
		DomainLabel:     "Domain:",
		ExpiryDateLabel: "Expiry Date:",
		DaysRemainingFmt: func(days int) string {
			return fmt.Sprintf("Days Remaining: %d", days)
		},
		RenewNowBtn:  "Renew Now",
		RenewSoonMsg: "Please complete the renewal as soon as possible to ensure your domain service is not affected.",
		Footer:       "This email was sent automatically by DNS Manager. Please do not reply directly.",
		// Test email
		TestSubject:      "DNS Manager - Email Configuration Test",
		TestSuccessTitle: "Email Configuration Test Successful",
		TestCongrats:     "Congratulations! Your email configuration is set up correctly.",
		TestConfirmation: "DNS Manager can now send you domain expiry reminder notifications.",
		TestFooter:       "This email was sent by DNS Manager.",
	},
}

// GetEmailTranslations returns translations for the given language.
// Falls back to the provided defaultLang, then to Chinese ("zh").
func GetEmailTranslations(lang string, defaultLang string) EmailTranslations {
	if lang != "" {
		if t, ok := emailTranslations[lang]; ok {
			return t
		}
	}
	if defaultLang != "" {
		if t, ok := emailTranslations[defaultLang]; ok {
			return t
		}
	}
	return emailTranslations["zh"]
}
