package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"
)

type EmailConfig struct {
	Host     string
	Port     int
	FromAddr string
	Password string
}

type EmailService struct {
	cfg               EmailConfig
	auth              smtp.Auth
	welcomeEmailQueue chan [2]string
}

func New(ctx context.Context, cfg EmailConfig) (chan [2]string, error) {
	const op = "pkg.email.New"

	emailService := &EmailService{
		cfg:               cfg,
		auth:              smtp.PlainAuth("", cfg.FromAddr, cfg.Password, cfg.Host),
		welcomeEmailQueue: make(chan [2]string, 20),
	}

	tmpl, err := template.ParseFiles("./pkg/email/templates/verify-email.html")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	go emailService.sendWelcomeMessages(ctx, tmpl)

	return emailService.welcomeEmailQueue, nil
}

type VerificationEmailVars struct {
	Token string
}

func (e *EmailService) sendWelcomeMessages(ctx context.Context, tmpl *template.Template) {
	rendered := new(bytes.Buffer)

	for {
		select {
		case <-ctx.Done():
			return
		case emailVars := <-e.welcomeEmailQueue:
			vars := VerificationEmailVars{
				Token: emailVars[1],
			}

			err := tmpl.Execute(rendered, vars)
			if err != nil {
				rendered.Reset()
				continue
			}

			headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";"

			err = smtp.SendMail(
				fmt.Sprintf("%s:%d", e.cfg.Host, e.cfg.Port),
				e.auth,
				e.cfg.FromAddr,
				[]string{emailVars[0]},
				fmt.Appendf(nil, "Subject: Email\n%s\n\n%s", headers, rendered.String()),
			)
			if err != nil {
				rendered.Reset()
				continue
			}
		}
	}
}
