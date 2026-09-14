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
	smtpUser := os.Getenv("SMTP_USERNAME")
	if smtpUser == "" {
		smtpUser = os.Getenv("SMTP_USER")
	}
	smtpPass := os.Getenv("SMTP_PASSWORD")
	if smtpPass == "" {
		smtpPass = os.Getenv("SMTP_PASS")
	}

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

// SendVerificationEmail sends the registration verification email
func SendVerificationEmail(toEmail string, token string) error {
	subject := "Verifikasi Email JurnalUmi"
	baseURL := os.Getenv("APP_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	body := fmt.Sprintf("Halo,\n\nTerima kasih telah mendaftar di JurnalUmi. Silakan verifikasi email Anda dengan mengklik tautan berikut:\n\n%s/verify-email?token=%s\n\nTautan ini akan kedaluwarsa.\n\nSalam,\nTim JurnalUmi", baseURL, token)
	return SendEmailNotification(toEmail, subject, body)
}

// SendResetPasswordEmail sends the password reset email with token link
func SendResetPasswordEmail(toEmail string, token string) error {
	subject := "Reset Password JurnalUmi"
	baseURL := os.Getenv("APP_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	body := fmt.Sprintf("Halo,\n\nKami menerima permintaan untuk mereset kata sandi akun JurnalUmi Anda. Silakan klik tautan berikut untuk membuat kata sandi baru:\n\n%s/reset-password?token=%s\n\nTautan ini hanya berlaku selama 1 jam.\n\nJika Anda tidak merasa meminta reset password, abaikan email ini.\n\nSalam,\nTim JurnalUmi", baseURL, token)
	return SendEmailNotification(toEmail, subject, body)
}

