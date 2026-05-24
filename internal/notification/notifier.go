package notification

import "context"

type NotificationKind string

const (
	KindTest          NotificationKind = "test"
	KindPasswordReset NotificationKind = "password_reset"
	KindMagicLink     NotificationKind = "magic_link"
	KindSecurityAlert NotificationKind = "security_alert"
)

type EmailMessage struct {
	To      string
	Subject string
	Text    string
	HTML    string
}

type Notifier interface {
	Send(ctx context.Context, msg EmailMessage) error
}
