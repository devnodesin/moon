package main

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"time"
)

// EmailSender sends emails via SMTP.
type EmailSender struct {
	cfg    EmailConfig
	logger *Logger
}

// NewEmailSender creates an EmailSender from config.
func NewEmailSender(cfg EmailConfig, logger *Logger) *EmailSender {
	return &EmailSender{cfg: cfg, logger: logger}
}

// SendOTPEmail sends an OTP code to the given recipient.
func (e *EmailSender) SendOTPEmail(to, code string) error {
	if e == nil {
		return fmt.Errorf("email sender not configured")
	}

	addr := fmt.Sprintf("%s:%d", e.cfg.SMTPHost, e.cfg.SMTPPort)
	from := e.cfg.EmailSender
	subject := "Your login code"
	body := fmt.Sprintf("Your one-time login code is: %s\n\nIt expires in %d minutes.", code, OTPTTLMinutes)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nDate: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n",
		from, to, subject, time.Now().UTC().Format(time.RFC1123Z), body)

	var auth smtp.Auth
	if e.cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", e.cfg.SMTPUser, e.cfg.SMTPPassword, e.cfg.SMTPHost)
	}

	if e.cfg.UseTLS {
		tlsConn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: e.cfg.SMTPHost, InsecureSkipVerify: e.cfg.SkipTLSVerify})
		if err != nil {
			return err
		}
		defer tlsConn.Close()
		client, err := smtp.NewClient(tlsConn, e.cfg.SMTPHost)
		if err != nil {
			return err
		}
		defer client.Close()
		if auth != nil {
			if err := client.Auth(auth); err != nil {
				return err
			}
		}
		if err := client.Mail(from); err != nil {
			return err
		}
		if err := client.Rcpt(to); err != nil {
			return err
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		if _, err := w.Write([]byte(msg)); err != nil {
			w.Close()
			return err
		}
		if err := w.Close(); err != nil {
			return err
		}
		return client.Quit()
	}

	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}
