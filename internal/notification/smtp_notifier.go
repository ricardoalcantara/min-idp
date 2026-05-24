package notification

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
	"strings"

	"github.com/ricardoalcantara/min-idp/internal/config"
)

type SMTPNotifier struct {
	cfg  *config.Config
	log  *slog.Logger
	addr string
	from string
}

func NewSMTPNotifier(cfg *config.Config, log *slog.Logger) (*SMTPNotifier, error) {
	if cfg.SMTPTLS == "none" {
		log.Warn("notification: SMTP TLS disabled — use only for local development")
	}
	return &SMTPNotifier{
		cfg:  cfg,
		log:  log,
		addr: fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort),
		from: formatAddress(cfg.SMTPFromName, cfg.SMTPFrom),
	}, nil
}

func (n *SMTPNotifier) Send(ctx context.Context, msg EmailMessage) error {
	body := buildMIMEBody(n.from, msg.To, msg.Subject, msg.Text, msg.HTML)

	var client *smtp.Client
	var err error

	switch n.cfg.SMTPTLS {
	case "ssl":
		client, err = n.dialSSL()
	case "starttls", "none":
		client, err = n.dialPlain()
	default:
		return fmt.Errorf("notification: unsupported SMTP TLS mode %q", n.cfg.SMTPTLS)
	}
	if err != nil {
		return err
	}
	defer client.Close()

	if n.cfg.SMTPTLS == "starttls" {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return fmt.Errorf("notification: SMTP server does not support STARTTLS")
		}
		if err := client.StartTLS(&tls.Config{ServerName: n.cfg.SMTPHost}); err != nil {
			return fmt.Errorf("notification: STARTTLS: %w", err)
		}
	}

	if n.cfg.SMTPUsername != "" {
		auth := smtp.PlainAuth("", n.cfg.SMTPUsername, n.cfg.SMTPPassword, n.cfg.SMTPHost)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("notification: SMTP auth: %w", err)
		}
	}

	if err := client.Mail(n.cfg.SMTPFrom); err != nil {
		return fmt.Errorf("notification: MAIL FROM: %w", err)
	}
	if err := client.Rcpt(msg.To); err != nil {
		return fmt.Errorf("notification: RCPT TO: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("notification: DATA: %w", err)
	}
	if _, err := w.Write(body); err != nil {
		return fmt.Errorf("notification: write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("notification: close body: %w", err)
	}

	return client.Quit()
}

func (n *SMTPNotifier) dialPlain() (*smtp.Client, error) {
	conn, err := net.Dial("tcp", n.addr)
	if err != nil {
		return nil, fmt.Errorf("notification: dial SMTP: %w", err)
	}
	client, err := smtp.NewClient(conn, n.cfg.SMTPHost)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("notification: SMTP client: %w", err)
	}
	return client, nil
}

func (n *SMTPNotifier) dialSSL() (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", n.addr, &tls.Config{ServerName: n.cfg.SMTPHost})
	if err != nil {
		return nil, fmt.Errorf("notification: dial SMTPS: %w", err)
	}
	client, err := smtp.NewClient(conn, n.cfg.SMTPHost)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("notification: SMTPS client: %w", err)
	}
	return client, nil
}

func formatAddress(name, email string) string {
	if name == "" {
		return email
	}
	return fmt.Sprintf("%s <%s>", name, email)
}

func buildMIMEBody(from, to, subject, textBody, htmlBody string) []byte {
	boundary := "min-idp-boundary"
	var buf bytes.Buffer
	buf.WriteString("From: " + from + "\r\n")
	buf.WriteString("To: " + to + "\r\n")
	buf.WriteString("Subject: " + subject + "\r\n")
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: multipart/alternative; boundary=" + boundary + "\r\n")
	buf.WriteString("\r\n")

	writePart := func(contentType, body string) {
		buf.WriteString("--" + boundary + "\r\n")
		buf.WriteString("Content-Type: " + contentType + "; charset=utf-8\r\n")
		buf.WriteString("Content-Transfer-Encoding: 8bit\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(body)
		if !strings.HasSuffix(body, "\r\n") {
			buf.WriteString("\r\n")
		}
	}

	writePart("text/plain", textBody)
	writePart("text/html", htmlBody)
	buf.WriteString("--" + boundary + "--\r\n")

	return buf.Bytes()
}
