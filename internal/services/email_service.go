package services

import (
	"fmt"
	"net/smtp"
	"os"
)

// EmailService handles email operations
type EmailService struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
}

// NewEmailService creates a new email service
func NewEmailService() *EmailService {
	return &EmailService{
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		FromEmail:    os.Getenv("FROM_EMAIL"),
	}
}

// SendVerificationEmail sends an email verification link
func (s *EmailService) SendVerificationEmail(toEmail, verificationToken string) error {
	// Email configuration validation
	if s.SMTPHost == "" || s.SMTPPort == "" || s.SMTPUsername == "" || s.SMTPPassword == "" {
		// For development, just log the verification token
		fmt.Printf("\n=== EMAIL VERIFICATION (DEV MODE) ===\n")
		fmt.Printf("To: %s\n", toEmail)
		fmt.Printf("Verification Link: http://localhost:8080/api/v1/auth/verify-email?token=%s\n", verificationToken)
		fmt.Printf("=====================================\n\n")
		return nil
	}

	// Email content
	subject := "Verify Your Email Address"
	verificationURL := fmt.Sprintf("http://localhost:8080/api/v1/auth/verify-email?token=%s", verificationToken)

	body := fmt.Sprintf(`
Hello,

Thank you for registering! Please click the link below to verify your email address:

%s

This link will expire in 24 hours.

If you didn't create an account, please ignore this email.

Best regards,
Your App Team
`, verificationURL)

	// Email message
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		s.FromEmail, toEmail, subject, body)

	// SMTP authentication
	auth := smtp.PlainAuth("", s.SMTPUsername, s.SMTPPassword, s.SMTPHost)

	// Send email
	err := smtp.SendMail(
		s.SMTPHost+":"+s.SMTPPort,
		auth,
		s.FromEmail,
		[]string{toEmail},
		[]byte(message),
	)

	return err
}

// SendPasswordResetEmail sends a password reset email (for future use)
func (s *EmailService) SendPasswordResetEmail(toEmail, resetToken string) error {
	// Similar implementation for password reset
	if s.SMTPHost == "" || s.SMTPPort == "" || s.SMTPUsername == "" || s.SMTPPassword == "" {
		fmt.Printf("\n=== PASSWORD RESET (DEV MODE) ===\n")
		fmt.Printf("To: %s\n", toEmail)
		fmt.Printf("Reset Link: http://localhost:8080/api/v1/auth/reset-password?token=%s\n", resetToken)
		fmt.Printf("=================================\n\n")
		return nil
	}

	// Implementation for production email sending
	return nil
}
