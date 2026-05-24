package notification

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"strings"
	texttemplate "text/template"

	notification_templates "github.com/ricardoalcantara/min-idp/internal/notification/templates"
	notification_types "github.com/ricardoalcantara/min-idp/internal/notification/types"
)

var errUnknownKind = errors.New("unknown notification kind")
var errInvalidTemplateData = errors.New("invalid notification template data")

type RenderedEmail struct {
	Subject string
	Text    string
	HTML    string
}

type TemplateRenderer struct {
	html map[NotificationKind]*template.Template
	text map[NotificationKind]*texttemplate.Template
}

func NewTemplateRenderer() (*TemplateRenderer, error) {
	kinds := []struct {
		kind NotificationKind
		base string
	}{
		{KindTest, "test"},
		{KindPasswordReset, "password_reset"},
		{KindMagicLink, "magic_link"},
		{KindSecurityAlert, "security_alert"},
	}

	r := &TemplateRenderer{
		html: make(map[NotificationKind]*template.Template, len(kinds)),
		text: make(map[NotificationKind]*texttemplate.Template, len(kinds)),
	}

	for _, k := range kinds {
		htmlTmpl, err := template.ParseFS(notification_templates.Files, k.base+".html")
		if err != nil {
			return nil, fmt.Errorf("notification: parse html template %q: %w", k.base, err)
		}
		textTmpl, err := texttemplate.ParseFS(notification_templates.Files, k.base+".txt")
		if err != nil {
			return nil, fmt.Errorf("notification: parse text template %q: %w", k.base, err)
		}
		r.html[k.kind] = htmlTmpl
		r.text[k.kind] = textTmpl
	}

	return r, nil
}

func (r *TemplateRenderer) Render(kind NotificationKind, data any) (RenderedEmail, error) {
	htmlTmpl, ok := r.html[kind]
	if !ok {
		return RenderedEmail{}, errUnknownKind
	}
	textTmpl := r.text[kind]

	if err := validateTemplateData(kind, data); err != nil {
		return RenderedEmail{}, err
	}

	var htmlBuf, textBuf bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBuf, data); err != nil {
		return RenderedEmail{}, fmt.Errorf("notification: render html: %w", err)
	}
	if err := textTmpl.Execute(&textBuf, data); err != nil {
		return RenderedEmail{}, fmt.Errorf("notification: render text: %w", err)
	}

	return RenderedEmail{
		Subject: subjectForKind(kind),
		Text:    strings.TrimSpace(textBuf.String()),
		HTML:    strings.TrimSpace(htmlBuf.String()),
	}, nil
}

func subjectForKind(kind NotificationKind) string {
	switch kind {
	case KindTest:
		return "min-idp test email"
	case KindPasswordReset:
		return "Reset your password"
	case KindMagicLink:
		return "Sign in to min-idp"
	case KindSecurityAlert:
		return "New sign-in detected"
	default:
		return "Notification from min-idp"
	}
}

func validateTemplateData(kind NotificationKind, data any) error {
	switch kind {
	case KindTest:
		if _, ok := data.(notification_types.TestTemplateData); !ok {
			return errInvalidTemplateData
		}
	case KindPasswordReset:
		if _, ok := data.(notification_types.PasswordResetTemplateData); !ok {
			return errInvalidTemplateData
		}
	case KindMagicLink:
		if _, ok := data.(notification_types.MagicLinkTemplateData); !ok {
			return errInvalidTemplateData
		}
	case KindSecurityAlert:
		if _, ok := data.(notification_types.SecurityAlertTemplateData); !ok {
			return errInvalidTemplateData
		}
	default:
		return errUnknownKind
	}
	return nil
}
