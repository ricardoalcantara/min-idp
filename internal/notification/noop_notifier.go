package notification

import (
	"context"
	"log/slog"
)

type NoopNotifier struct {
	log *slog.Logger
}

func NewNoopNotifier(log *slog.Logger) *NoopNotifier {
	return &NoopNotifier{log: log}
}

func (n *NoopNotifier) Send(ctx context.Context, msg EmailMessage) error {
	n.log.Debug("notification: skipped (smtp disabled)", "to", msg.To, "subject", msg.Subject)
	return nil
}
