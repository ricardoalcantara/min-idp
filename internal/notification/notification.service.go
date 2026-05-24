package notification

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ricardoalcantara/min-idp/internal/config"
	notification_types "github.com/ricardoalcantara/min-idp/internal/notification/types"
)

var errSMTPNotConfigured = errors.New("smtp not configured")

type NotificationService struct {
	cfg      *config.Config
	notifier Notifier
	renderer *TemplateRenderer
	log      *slog.Logger
}

func NewNotificationService(
	cfg *config.Config,
	notifier Notifier,
	renderer *TemplateRenderer,
	log *slog.Logger,
) *NotificationService {
	return &NotificationService{
		cfg:      cfg,
		notifier: notifier,
		renderer: renderer,
		log:      log,
	}
}

func (s *NotificationService) Send(ctx context.Context, kind NotificationKind, recipient string, data any) error {
	if !s.cfg.SMTPEnabled {
		s.log.Debug("notification: skipped (smtp disabled)", "kind", kind, "to", recipient)
		return nil
	}

	rendered, err := s.renderer.Render(kind, data)
	if err != nil {
		s.log.Debug("notification: render failed", "kind", kind, "to", recipient, "err", err)
		return err
	}

	msg := EmailMessage{
		To:      recipient,
		Subject: rendered.Subject,
		Text:    rendered.Text,
		HTML:    rendered.HTML,
	}

	if err := s.notifier.Send(ctx, msg); err != nil {
		return err
	}

	if kind == KindTest {
		s.log.Info("notification: test email sent", "to", recipient)
	}

	return nil
}

func (s *NotificationService) SendAsync(kind NotificationKind, recipient string, data any) {
	go func() {
		if err := s.Send(context.Background(), kind, recipient, data); err != nil {
			s.log.Error("notification: send failed", "kind", kind, "to", recipient, "err", err)
		}
	}()
}

func (s *NotificationService) SendTest(ctx context.Context, to string) error {
	if !s.cfg.SMTPEnabled {
		return errSMTPNotConfigured
	}
	return s.Send(ctx, KindTest, to, notification_types.TestTemplateData{
		AppName:     s.cfg.SMTPFromName,
		ExternalURL: s.cfg.ExternalURL,
	})
}
