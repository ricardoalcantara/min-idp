package notification

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/ricardoalcantara/min-idp/internal/config"
	notification_types "github.com/ricardoalcantara/min-idp/internal/notification/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockNotifier struct {
	lastMsg EmailMessage
	err     error
	called  bool
}

func (m *mockNotifier) Send(ctx context.Context, msg EmailMessage) error {
	m.called = true
	m.lastMsg = msg
	return m.err
}

func TestTemplateRenderer_Render_TestKind(t *testing.T) {
	r, err := NewTemplateRenderer()
	require.NoError(t, err)

	rendered, err := r.Render(KindTest, notification_types.TestTemplateData{
		AppName:     "min-idp",
		ExternalURL: "https://idp.example.com",
	})
	require.NoError(t, err)

	assert.Equal(t, "min-idp test email", rendered.Subject)
	assert.Contains(t, rendered.Text, "min-idp")
	assert.Contains(t, rendered.Text, "https://idp.example.com")
	assert.Contains(t, rendered.HTML, "https://idp.example.com")
}

func TestTemplateRenderer_Render_AllKinds(t *testing.T) {
	r, err := NewTemplateRenderer()
	require.NoError(t, err)

	cases := []struct {
		kind    NotificationKind
		data    any
		contain string
	}{
		{KindPasswordReset, notification_types.PasswordResetTemplateData{ResetURL: "https://idp.example.com/reset", ExpiresIn: "15m"}, "https://idp.example.com/reset"},
		{KindMagicLink, notification_types.MagicLinkTemplateData{LoginURL: "https://idp.example.com/login", ExpiresIn: "10m"}, "https://idp.example.com/login"},
		{KindSecurityAlert, notification_types.SecurityAlertTemplateData{IP: "1.2.3.4", UserAgent: "curl", Time: "now", RevokeSessionsURL: "https://idp.example.com/sessions"}, "1.2.3.4"},
	}

	for _, tc := range cases {
		t.Run(string(tc.kind), func(t *testing.T) {
			rendered, err := r.Render(tc.kind, tc.data)
			require.NoError(t, err)
			assert.NotEmpty(t, rendered.Subject)
			assert.Contains(t, rendered.Text, tc.contain)
			assert.Contains(t, rendered.HTML, tc.contain)
		})
	}
}

func TestTemplateRenderer_Render_UnknownKind(t *testing.T) {
	r, err := NewTemplateRenderer()
	require.NoError(t, err)

	_, err = r.Render(NotificationKind("unknown"), notification_types.TestTemplateData{})
	assert.ErrorIs(t, err, errUnknownKind)
}

func TestTemplateRenderer_Render_InvalidTemplateData(t *testing.T) {
	r, err := NewTemplateRenderer()
	require.NoError(t, err)

	_, err = r.Render(KindTest, notification_types.PasswordResetTemplateData{})
	assert.ErrorIs(t, err, errInvalidTemplateData)
}

func TestNotificationService_Send_SMTPEnabled(t *testing.T) {
	mock := &mockNotifier{}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	r, err := NewTemplateRenderer()
	require.NoError(t, err)

	svc := NewNotificationService(&config.Config{
		SMTPEnabled:  true,
		SMTPFromName: "min-idp",
		ExternalURL:  "https://idp.example.com",
	}, mock, r, log)

	err = svc.Send(context.Background(), KindTest, "user@example.com", notification_types.TestTemplateData{
		AppName:     "min-idp",
		ExternalURL: "https://idp.example.com",
	})
	require.NoError(t, err)
	assert.True(t, mock.called)
	assert.Equal(t, "user@example.com", mock.lastMsg.To)
	assert.Equal(t, "min-idp test email", mock.lastMsg.Subject)
	assert.NotEmpty(t, mock.lastMsg.Text)
	assert.NotEmpty(t, mock.lastMsg.HTML)
}

func TestNotificationService_Send_SMTPEnabled_NoOpNotifier(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	r, err := NewTemplateRenderer()
	require.NoError(t, err)

	svc := NewNotificationService(&config.Config{SMTPEnabled: false}, NewNoopNotifier(log), r, log)

	err = svc.Send(context.Background(), KindTest, "user@example.com", notification_types.TestTemplateData{
		AppName:     "min-idp",
		ExternalURL: "https://idp.example.com",
	})
	require.NoError(t, err)
}

func TestNotificationService_SendTest_NotConfigured(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	r, err := NewTemplateRenderer()
	require.NoError(t, err)

	svc := NewNotificationService(&config.Config{SMTPEnabled: false}, NewNoopNotifier(log), r, log)

	err = svc.SendTest(context.Background(), "user@example.com")
	assert.ErrorIs(t, err, errSMTPNotConfigured)
}

func TestNotificationService_Send_RenderError(t *testing.T) {
	mock := &mockNotifier{}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	r, err := NewTemplateRenderer()
	require.NoError(t, err)

	svc := NewNotificationService(&config.Config{SMTPEnabled: true}, mock, r, log)

	err = svc.Send(context.Background(), KindTest, "user@example.com", notification_types.PasswordResetTemplateData{})
	assert.ErrorIs(t, err, errInvalidTemplateData)
	assert.False(t, mock.called)
}
