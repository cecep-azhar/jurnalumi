package services

import (
	"fmt"
	"net/smtp"
	"os"
)

// SendEmailNotification sends SMTP emails for Debt Reminders and Budget Alerts
func SendEmailNotification(toEmail string, subject string, body string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")

	if smtpHost == "" || smtpUser == "" {
		// Log/Mock if SMTP is not configured in local environment
		fmt.Printf("[SMTP MOCK] To: %s | Subject: %s | Body: %s\n", toEmail, subject, body)
		return nil
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", toEmail, subject, body))

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUser, []string{toEmail}, msg)
	return err
}
