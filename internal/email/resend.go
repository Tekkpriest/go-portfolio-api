package email

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"time"

	"github.com/resend/resend-go/v3"
)

type ResendClient interface {
	SendWithContext(ctx context.Context, params *resend.SendEmailRequest) (*resend.SendEmailResponse, error)
}

type MailService struct {
	client    ResendClient
	emailFrom string
	emailTo   string
}

func NewMailService(apiKey, from, to string) *MailService {
	client := resend.NewClient(apiKey)
	return &MailService{
		client:    client.Emails,
		emailFrom: from,
		emailTo:   to,
	}
}

func (m *MailService) SendMail(ctx context.Context, name, email, subject, company, message string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	htmlBody := fmt.Sprintf(`
		<p><strong>Name:</strong> %s</p>
		<p><strong>E-Mail:</strong> %s</p>
		<p><strong>Firma:</strong> %s</p>
		<p><strong>Nachricht:</strong>%s<br></p>
	`,
		html.EscapeString(name),
		html.EscapeString(email),
		html.EscapeString(company),
		html.EscapeString(message),
	)

	params := &resend.SendEmailRequest{
		From:    m.emailFrom,
		To:      []string{m.emailTo},
		Subject: subject,
		Html:    htmlBody,
	}

	sent, err := m.client.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("resend api error: %w", err)
	}
	slog.Info("email sent successfully to", "email_id", sent.Id)

	return nil
}
