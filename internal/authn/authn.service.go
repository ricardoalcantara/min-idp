package authn

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/ricardoalcantara/min-idp/internal/config"
	"github.com/ricardoalcantara/min-idp/internal/crypto"
	"github.com/ricardoalcantara/min-idp/internal/db"
	"github.com/ricardoalcantara/min-idp/internal/kvstore"
	"github.com/ricardoalcantara/min-idp/internal/notification"
	notification_types "github.com/ricardoalcantara/min-idp/internal/notification/types"
	"github.com/ricardoalcantara/min-idp/internal/session"
	"github.com/ricardoalcantara/min-idp/internal/types"
	"github.com/ricardoalcantara/min-idp/internal/users"
	user_entities "github.com/ricardoalcantara/min-idp/internal/users/entities"
)

var errInvalidCredentials = errors.New("invalid credentials")
var errInvalidResetToken = errors.New("invalid or expired reset token")
var errUnsupportedPKCEMethod = errors.New("unsupported code_challenge_method")
var errWeakPassword = errors.New("password must be at least 8 characters")

const passwordResetKeyPrefix = "password_reset:"

type UserAuthenticator interface {
	Authenticate(email, password string) (*user_entities.User, error)
}

type userDirectory interface {
	FindByEmail(email string) (*user_entities.User, error)
	UpdatePassword(userID uint, newPassword string) error
}

type notifier interface {
	SendAsync(kind notification.NotificationKind, recipient string, data any)
}

type sessionRevoker interface {
	RevokeAll(ctx context.Context, userID uint) error
}

type passwordResetData struct {
	UserID              uint   `json:"user_id"`
	CodeChallenge       string `json:"code_challenge,omitempty"`
	CodeChallengeMethod string `json:"code_challenge_method,omitempty"`
}

type AuthnService struct {
	users        UserAuthenticator
	userDir      userDirectory
	kv           kvstore.KVStore
	notification notifier
	session      sessionRevoker
	cfg          *config.Config
}

func NewAuthnService(
	auth UserAuthenticator,
	userSvc *users.UserService,
	kv kvstore.KVStore,
	notif *notification.NotificationService,
	sessionSvc *session.SessionService,
	cfg *config.Config,
) *AuthnService {
	return &AuthnService{
		users:        auth,
		userDir:      userSvc,
		kv:           kv,
		notification: notif,
		session:      sessionSvc,
		cfg:          cfg,
	}
}

func (s *AuthnService) Authenticate(email, password string) (*user_entities.User, error) {
	u, err := s.users.Authenticate(email, password)
	if err != nil {
		if errors.Is(err, users.ErrInvalidCredentials) || errors.Is(err, users.ErrAccountNotActive) {
			return nil, errInvalidCredentials
		}
		return nil, err
	}
	return u, nil
}

// RequestPasswordReset stores a single-use reset token in KV and emails the
// user. To avoid account enumeration the method returns nil for unknown or
// inactive accounts — the caller (HTTP layer) always responds 200.
//
// When codeChallenge is non-empty, the caller is binding the reset to a PKCE
// verifier. codeChallengeMethod must be "S256" or "plain"; any other value
// (including empty) is rejected.
func (s *AuthnService) RequestPasswordReset(ctx context.Context, email, codeChallenge, codeChallengeMethod string) error {
	if codeChallenge != "" {
		switch codeChallengeMethod {
		case "S256", "plain":
		default:
			return errUnsupportedPKCEMethod
		}
	}

	u, err := s.userDir.FindByEmail(email)
	if err != nil {
		if errors.Is(err, db.ErrEntityNotFound) {
			return nil
		}
		return err
	}
	if u.Status != types.UserStatusActive {
		return nil
	}

	token, err := crypto.GenerateSecureToken(32)
	if err != nil {
		return err
	}

	payload := passwordResetData{
		UserID:              u.ID,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
	}
	val, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if err := s.kv.Set(ctx, passwordResetKey(token), val, s.cfg.PasswordResetTTL); err != nil {
		return err
	}

	resetURL := s.cfg.ExternalURL + "/reset-password?token=" + token
	s.notification.SendAsync(notification.KindPasswordReset, u.Email, notification_types.PasswordResetTemplateData{
		ResetURL:  resetURL,
		ExpiresIn: "15 minutes",
	})

	return nil
}

// ResetPassword consumes a reset token, updates the password, revokes existing
// sessions, and deletes the token. All token-validation failures collapse to
// errInvalidResetToken so the response leaks no information about token state.
func (s *AuthnService) ResetPassword(ctx context.Context, token, newPassword, codeVerifier string) error {
	if len(newPassword) < 8 {
		return errWeakPassword
	}
	if token == "" {
		return errInvalidResetToken
	}

	key := passwordResetKey(token)
	raw, err := s.kv.Get(ctx, key)
	if err != nil {
		return errInvalidResetToken
	}

	var data passwordResetData
	if err := json.Unmarshal(raw, &data); err != nil {
		return errInvalidResetToken
	}

	if data.CodeChallenge != "" {
		if codeVerifier == "" {
			return errInvalidResetToken
		}
		switch data.CodeChallengeMethod {
		case "S256":
			h := sha256.Sum256([]byte(codeVerifier))
			if base64.RawURLEncoding.EncodeToString(h[:]) != data.CodeChallenge {
				return errInvalidResetToken
			}
		case "plain":
			if codeVerifier != data.CodeChallenge {
				return errInvalidResetToken
			}
		default:
			return errInvalidResetToken
		}
	}

	if err := s.userDir.UpdatePassword(data.UserID, newPassword); err != nil {
		return err
	}
	if err := s.session.RevokeAll(ctx, data.UserID); err != nil {
		return err
	}
	_ = s.kv.Delete(ctx, key)

	return nil
}

func passwordResetKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return passwordResetKeyPrefix + base64.RawURLEncoding.EncodeToString(sum[:])
}
