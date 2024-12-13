package mailer

import (
	"e-commerce-users/internal/config"
	"fmt"
	"net/smtp"
	"time"
)

type Mailer struct {
	cfg *config.SMTP
}

func New(cfg *config.SMTP) *Mailer {
	return &Mailer{
		cfg: cfg,
	}
}

func (m *Mailer) Send(email, code string) error {
	const op = "repositories.Mailer.Send"

	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)

	msg := []byte(fmt.Sprintf(
		`
		From: %s
		Your verification code: %s
		`, m.cfg.Username, code,
	))

	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", m.cfg.Host, m.cfg.Port),
		auth,
		m.cfg.Username,
		[]string{email},
		msg)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (m *Mailer) CodeTTL() time.Duration {
	return m.cfg.CodeTTL
}
