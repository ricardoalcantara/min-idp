package notification_types

type TestTemplateData struct {
	AppName     string
	ExternalURL string
}

type PasswordResetTemplateData struct {
	ResetURL  string
	ExpiresIn string
}

type MagicLinkTemplateData struct {
	LoginURL  string
	ExpiresIn string
}

type SecurityAlertTemplateData struct {
	IP                string
	UserAgent         string
	Time              string
	RevokeSessionsURL string
}
