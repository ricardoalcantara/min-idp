package notification

import (
	"log/slog"

	"github.com/go-minstack/go-minstack/core"
	"github.com/ricardoalcantara/min-idp/internal/config"
)

func newNotifier(cfg *config.Config, log *slog.Logger) (Notifier, error) {
	if !cfg.SMTPEnabled {
		return NewNoopNotifier(log), nil
	}
	return NewSMTPNotifier(cfg, log)
}

func Register(app *core.App) {
	app.Provide(NewTemplateRenderer)
	app.Use(core.ProvideAs[Notifier](newNotifier))
	app.Provide(NewNotificationService)
	app.Provide(NewNotificationController)
}
