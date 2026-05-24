package authn

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ricardoalcantara/min-idp/internal/config"
	"github.com/ricardoalcantara/min-idp/internal/db"
	"github.com/ricardoalcantara/min-idp/internal/kvstore"
	"github.com/ricardoalcantara/min-idp/internal/notification"
	"github.com/ricardoalcantara/min-idp/internal/types"
	"github.com/ricardoalcantara/min-idp/internal/users"
	user_entities "github.com/ricardoalcantara/min-idp/internal/users/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mocks ---

type mockUserAuthenticator struct {
	user *user_entities.User
	err  error
}

func (m *mockUserAuthenticator) Authenticate(_, _ string) (*user_entities.User, error) {
	return m.user, m.err
}

type mockUserDir struct {
	user       *user_entities.User
	findErr    error
	updateErr  error
	updatedID  uint
	updatedPwd string
}

func (m *mockUserDir) FindByEmail(string) (*user_entities.User, error) {
	return m.user, m.findErr
}
func (m *mockUserDir) UpdatePassword(userID uint, newPassword string) error {
	m.updatedID = userID
	m.updatedPwd = newPassword
	return m.updateErr
}

type kvEntry struct {
	value []byte
	ttl   time.Duration
}

type mockKV struct {
	store     map[string]kvEntry
	getErr    error
	setErr    error
	deleteErr error
	deleted   []string
}

func newMockKV() *mockKV { return &mockKV{store: map[string]kvEntry{}} }

func (m *mockKV) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.store[key] = kvEntry{value: value, ttl: ttl}
	return nil
}
func (m *mockKV) Get(_ context.Context, key string) ([]byte, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if v, ok := m.store[key]; ok {
		return v.value, nil
	}
	return nil, kvstore.ErrNotFound
}
func (m *mockKV) Delete(_ context.Context, key string) error {
	m.deleted = append(m.deleted, key)
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.store, key)
	return nil
}
func (m *mockKV) SetNX(context.Context, string, []byte, time.Duration) (bool, error) {
	return true, nil
}

type sentNotification struct {
	kind      notification.NotificationKind
	recipient string
	data      any
}

type mockNotifier struct {
	sent []sentNotification
}

func (m *mockNotifier) SendAsync(kind notification.NotificationKind, recipient string, data any) {
	m.sent = append(m.sent, sentNotification{kind: kind, recipient: recipient, data: data})
}

type mockSession struct {
	revokedFor []uint
	err        error
}

func (m *mockSession) RevokeAll(_ context.Context, userID uint) error {
	m.revokedFor = append(m.revokedFor, userID)
	return m.err
}

// --- helpers ---

func newTestSvc(auth UserAuthenticator, userDir userDirectory, kv kvstore.KVStore, notif notifier, sess sessionRevoker) *AuthnService {
	return &AuthnService{
		users:        auth,
		userDir:      userDir,
		kv:           kv,
		notification: notif,
		session:      sess,
		cfg:          &config.Config{ExternalURL: "https://idp.example.com", PasswordResetTTL: 15 * time.Minute},
	}
}

func activeUser() *user_entities.User {
	return &user_entities.User{Email: "alice@example.com", Status: types.UserStatusActive}
}

func s256Challenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// --- Authenticate ---

func TestAuthnService_Authenticate_Success(t *testing.T) {
	u := &user_entities.User{Email: "a@b.com"}
	svc := newTestSvc(&mockUserAuthenticator{user: u}, &mockUserDir{}, newMockKV(), &mockNotifier{}, &mockSession{})
	got, err := svc.Authenticate("a@b.com", "pass")
	require.NoError(t, err)
	assert.Equal(t, "a@b.com", got.Email)
}

func TestAuthnService_Authenticate_InvalidCredentials(t *testing.T) {
	svc := newTestSvc(&mockUserAuthenticator{err: users.ErrInvalidCredentials}, &mockUserDir{}, newMockKV(), &mockNotifier{}, &mockSession{})
	_, err := svc.Authenticate("a@b.com", "wrong")
	assert.ErrorIs(t, err, errInvalidCredentials)
}

func TestAuthnService_Authenticate_AccountNotActive(t *testing.T) {
	svc := newTestSvc(&mockUserAuthenticator{err: users.ErrAccountNotActive}, &mockUserDir{}, newMockKV(), &mockNotifier{}, &mockSession{})
	_, err := svc.Authenticate("a@b.com", "pass")
	assert.ErrorIs(t, err, errInvalidCredentials)
}

func TestAuthnService_Authenticate_InfraError(t *testing.T) {
	infraErr := errors.New("db down")
	svc := newTestSvc(&mockUserAuthenticator{err: infraErr}, &mockUserDir{}, newMockKV(), &mockNotifier{}, &mockSession{})
	_, err := svc.Authenticate("a@b.com", "pass")
	assert.ErrorIs(t, err, infraErr)
	assert.NotErrorIs(t, err, errInvalidCredentials)
}

// --- RequestPasswordReset ---

func TestRequestPasswordReset_UnknownEmail_Silent(t *testing.T) {
	kv := newMockKV()
	notif := &mockNotifier{}
	svc := newTestSvc(&mockUserAuthenticator{}, &mockUserDir{findErr: db.ErrEntityNotFound}, kv, notif, &mockSession{})

	err := svc.RequestPasswordReset(context.Background(), "ghost@example.com", "", "")

	require.NoError(t, err)
	assert.Empty(t, kv.store, "no KV write for unknown email")
	assert.Empty(t, notif.sent, "no notification for unknown email")
}

func TestRequestPasswordReset_InactiveUser_Silent(t *testing.T) {
	kv := newMockKV()
	notif := &mockNotifier{}
	u := activeUser()
	u.Status = types.UserStatusDisabled
	svc := newTestSvc(&mockUserAuthenticator{}, &mockUserDir{user: u}, kv, notif, &mockSession{})

	err := svc.RequestPasswordReset(context.Background(), u.Email, "", "")

	require.NoError(t, err)
	assert.Empty(t, kv.store)
	assert.Empty(t, notif.sent)
}

func TestRequestPasswordReset_FindError_Propagated(t *testing.T) {
	infraErr := errors.New("db down")
	svc := newTestSvc(&mockUserAuthenticator{}, &mockUserDir{findErr: infraErr}, newMockKV(), &mockNotifier{}, &mockSession{})

	err := svc.RequestPasswordReset(context.Background(), "x@y.com", "", "")
	assert.ErrorIs(t, err, infraErr)
}

func TestRequestPasswordReset_Active_NoPKCE_StoresAndNotifies(t *testing.T) {
	kv := newMockKV()
	notif := &mockNotifier{}
	u := activeUser()
	u.ID = 42
	svc := newTestSvc(&mockUserAuthenticator{}, &mockUserDir{user: u}, kv, notif, &mockSession{})

	err := svc.RequestPasswordReset(context.Background(), u.Email, "", "")
	require.NoError(t, err)

	require.Len(t, kv.store, 1, "exactly one KV entry")
	for _, entry := range kv.store {
		var data passwordResetData
		require.NoError(t, json.Unmarshal(entry.value, &data))
		assert.Equal(t, uint(42), data.UserID)
		assert.Empty(t, data.CodeChallenge)
		assert.Empty(t, data.CodeChallengeMethod)
		assert.Equal(t, 15*time.Minute, entry.ttl)
	}

	require.Len(t, notif.sent, 1)
	assert.Equal(t, notification.KindPasswordReset, notif.sent[0].kind)
	assert.Equal(t, u.Email, notif.sent[0].recipient)
}

func TestRequestPasswordReset_Active_S256_StoresChallenge(t *testing.T) {
	kv := newMockKV()
	notif := &mockNotifier{}
	u := activeUser()
	u.ID = 7
	svc := newTestSvc(&mockUserAuthenticator{}, &mockUserDir{user: u}, kv, notif, &mockSession{})

	challenge := s256Challenge("a-very-secret-verifier-string-1234")
	err := svc.RequestPasswordReset(context.Background(), u.Email, challenge, "S256")
	require.NoError(t, err)

	require.Len(t, kv.store, 1)
	for _, entry := range kv.store {
		var data passwordResetData
		require.NoError(t, json.Unmarshal(entry.value, &data))
		assert.Equal(t, challenge, data.CodeChallenge)
		assert.Equal(t, "S256", data.CodeChallengeMethod)
	}
}

func TestRequestPasswordReset_Active_Plain_StoresChallenge(t *testing.T) {
	kv := newMockKV()
	u := activeUser()
	u.ID = 9
	svc := newTestSvc(&mockUserAuthenticator{}, &mockUserDir{user: u}, kv, &mockNotifier{}, &mockSession{})

	err := svc.RequestPasswordReset(context.Background(), u.Email, "plain-challenge-value", "plain")
	require.NoError(t, err)

	require.Len(t, kv.store, 1)
	for _, entry := range kv.store {
		var data passwordResetData
		require.NoError(t, json.Unmarshal(entry.value, &data))
		assert.Equal(t, "plain-challenge-value", data.CodeChallenge)
		assert.Equal(t, "plain", data.CodeChallengeMethod)
	}
}

func TestRequestPasswordReset_Challenge_EmptyMethod_Rejected(t *testing.T) {
	kv := newMockKV()
	notif := &mockNotifier{}
	svc := newTestSvc(&mockUserAuthenticator{}, &mockUserDir{user: activeUser()}, kv, notif, &mockSession{})

	err := svc.RequestPasswordReset(context.Background(), "a@b.com", "some-challenge", "")
	assert.ErrorIs(t, err, errUnsupportedPKCEMethod)
	assert.Empty(t, kv.store)
	assert.Empty(t, notif.sent)
}

func TestRequestPasswordReset_Challenge_UnknownMethod_Rejected(t *testing.T) {
	kv := newMockKV()
	notif := &mockNotifier{}
	svc := newTestSvc(&mockUserAuthenticator{}, &mockUserDir{user: activeUser()}, kv, notif, &mockSession{})

	err := svc.RequestPasswordReset(context.Background(), "a@b.com", "some-challenge", "SHA512")
	assert.ErrorIs(t, err, errUnsupportedPKCEMethod)
	assert.Empty(t, kv.store)
	assert.Empty(t, notif.sent)
}

// --- ResetPassword ---

func primeKV(t *testing.T, kv *mockKV, data passwordResetData) string {
	t.Helper()
	token := "raw-token-" + t.Name()
	val, err := json.Marshal(data)
	require.NoError(t, err)
	kv.store[passwordResetKey(token)] = kvEntry{value: val, ttl: 15 * time.Minute}
	return token
}

func TestResetPassword_WeakPassword_Rejected(t *testing.T) {
	kv := newMockKV()
	userDir := &mockUserDir{}
	svc := newTestSvc(&mockUserAuthenticator{}, userDir, kv, &mockNotifier{}, &mockSession{})

	err := svc.ResetPassword(context.Background(), "any-token", "short", "")
	assert.ErrorIs(t, err, errWeakPassword)
	assert.Zero(t, userDir.updatedID, "password not updated on weak input")
}

func TestResetPassword_EmptyToken_Rejected(t *testing.T) {
	svc := newTestSvc(&mockUserAuthenticator{}, &mockUserDir{}, newMockKV(), &mockNotifier{}, &mockSession{})

	err := svc.ResetPassword(context.Background(), "", "longenoughpwd", "")
	assert.ErrorIs(t, err, errInvalidResetToken)
}

func TestResetPassword_UnknownToken_Rejected(t *testing.T) {
	userDir := &mockUserDir{}
	svc := newTestSvc(&mockUserAuthenticator{}, userDir, newMockKV(), &mockNotifier{}, &mockSession{})

	err := svc.ResetPassword(context.Background(), "no-such-token", "longenoughpwd", "")
	assert.ErrorIs(t, err, errInvalidResetToken)
	assert.Zero(t, userDir.updatedID)
}

func TestResetPassword_NonPKCE_Success(t *testing.T) {
	kv := newMockKV()
	userDir := &mockUserDir{}
	sess := &mockSession{}
	svc := newTestSvc(&mockUserAuthenticator{}, userDir, kv, &mockNotifier{}, sess)

	token := primeKV(t, kv, passwordResetData{UserID: 11})

	err := svc.ResetPassword(context.Background(), token, "newpassword123", "")
	require.NoError(t, err)
	assert.Equal(t, uint(11), userDir.updatedID)
	assert.Equal(t, "newpassword123", userDir.updatedPwd)
	assert.Equal(t, []uint{11}, sess.revokedFor)
	assert.Contains(t, kv.deleted, passwordResetKey(token))
	assert.Empty(t, kv.store, "KV entry consumed")
}

func TestResetPassword_PKCE_MissingVerifier_Rejected(t *testing.T) {
	kv := newMockKV()
	userDir := &mockUserDir{}
	svc := newTestSvc(&mockUserAuthenticator{}, userDir, kv, &mockNotifier{}, &mockSession{})

	token := primeKV(t, kv, passwordResetData{
		UserID:              22,
		CodeChallenge:       s256Challenge("verifier-abc"),
		CodeChallengeMethod: "S256",
	})

	err := svc.ResetPassword(context.Background(), token, "newpassword123", "")
	assert.ErrorIs(t, err, errInvalidResetToken)
	assert.Zero(t, userDir.updatedID, "no password change when verifier missing")
}

func TestResetPassword_PKCE_S256_WrongVerifier_Rejected(t *testing.T) {
	kv := newMockKV()
	userDir := &mockUserDir{}
	svc := newTestSvc(&mockUserAuthenticator{}, userDir, kv, &mockNotifier{}, &mockSession{})

	token := primeKV(t, kv, passwordResetData{
		UserID:              22,
		CodeChallenge:       s256Challenge("correct-verifier"),
		CodeChallengeMethod: "S256",
	})

	err := svc.ResetPassword(context.Background(), token, "newpassword123", "wrong-verifier")
	assert.ErrorIs(t, err, errInvalidResetToken)
	assert.Zero(t, userDir.updatedID)
}

func TestResetPassword_PKCE_S256_CorrectVerifier_Success(t *testing.T) {
	kv := newMockKV()
	userDir := &mockUserDir{}
	sess := &mockSession{}
	svc := newTestSvc(&mockUserAuthenticator{}, userDir, kv, &mockNotifier{}, sess)

	verifier := "the-actual-secret-verifier-value"
	token := primeKV(t, kv, passwordResetData{
		UserID:              33,
		CodeChallenge:       s256Challenge(verifier),
		CodeChallengeMethod: "S256",
	})

	err := svc.ResetPassword(context.Background(), token, "newpassword123", verifier)
	require.NoError(t, err)
	assert.Equal(t, uint(33), userDir.updatedID)
	assert.Equal(t, []uint{33}, sess.revokedFor)
	assert.Empty(t, kv.store)
}

func TestResetPassword_PKCE_Plain_CorrectVerifier_Success(t *testing.T) {
	kv := newMockKV()
	userDir := &mockUserDir{}
	sess := &mockSession{}
	svc := newTestSvc(&mockUserAuthenticator{}, userDir, kv, &mockNotifier{}, sess)

	token := primeKV(t, kv, passwordResetData{
		UserID:              44,
		CodeChallenge:       "shared-secret",
		CodeChallengeMethod: "plain",
	})

	err := svc.ResetPassword(context.Background(), token, "newpassword123", "shared-secret")
	require.NoError(t, err)
	assert.Equal(t, uint(44), userDir.updatedID)
	assert.Equal(t, []uint{44}, sess.revokedFor)
}

func TestResetPassword_PKCE_Plain_WrongVerifier_Rejected(t *testing.T) {
	kv := newMockKV()
	userDir := &mockUserDir{}
	svc := newTestSvc(&mockUserAuthenticator{}, userDir, kv, &mockNotifier{}, &mockSession{})

	token := primeKV(t, kv, passwordResetData{
		UserID:              44,
		CodeChallenge:       "shared-secret",
		CodeChallengeMethod: "plain",
	})

	err := svc.ResetPassword(context.Background(), token, "newpassword123", "different-secret")
	assert.ErrorIs(t, err, errInvalidResetToken)
	assert.Zero(t, userDir.updatedID)
}

func TestResetPassword_PKCE_UnknownMethod_Rejected(t *testing.T) {
	kv := newMockKV()
	userDir := &mockUserDir{}
	svc := newTestSvc(&mockUserAuthenticator{}, userDir, kv, &mockNotifier{}, &mockSession{})

	token := primeKV(t, kv, passwordResetData{
		UserID:              55,
		CodeChallenge:       "anything",
		CodeChallengeMethod: "weird",
	})

	err := svc.ResetPassword(context.Background(), token, "newpassword123", "anything")
	assert.ErrorIs(t, err, errInvalidResetToken)
}
