package session

import (
	"strings"
	"testing"

	"github.com/ricardoalcantara/min-idp/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCookieTokenService(t *testing.T) *CookieTokenService {
	t.Helper()
	// 32 zero bytes base64-encoded
	svc, err := NewCookieTokenService(&config.Config{
		MasterKey: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
	})
	require.NoError(t, err)
	return svc
}

func TestCookieTokenService_RoundTrip(t *testing.T) {
	svc := newTestCookieTokenService(t)
	original := "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0In0.signature"

	encoded, err := svc.Encode(original)
	require.NoError(t, err)
	assert.NotEqual(t, original, encoded)

	decoded, err := svc.Decode(encoded)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestCookieTokenService_OutputIsBase64URL(t *testing.T) {
	svc := newTestCookieTokenService(t)
	encoded, err := svc.Encode("some.jwt.value")
	require.NoError(t, err)
	// base64url must not contain +, /, or = padding
	assert.False(t, strings.ContainsAny(encoded, "+/="), "expected base64url encoding, got: %s", encoded)
}

func TestCookieTokenService_TamperedCookieFails(t *testing.T) {
	svc := newTestCookieTokenService(t)
	encoded, err := svc.Encode("original.jwt")
	require.NoError(t, err)

	// flip the last byte of the base64 string
	tampered := encoded[:len(encoded)-1] + "X"
	_, err = svc.Decode(tampered)
	assert.Error(t, err, "tampered cookie should fail decryption")
}

func TestCookieTokenService_InvalidBase64Fails(t *testing.T) {
	svc := newTestCookieTokenService(t)
	_, err := svc.Decode("not-valid-base64!!!")
	assert.Error(t, err)
}

func TestCookieTokenService_DifferentKeyCannotDecode(t *testing.T) {
	svc1 := newTestCookieTokenService(t)
	svc2, err := NewCookieTokenService(&config.Config{
		// different key: 32 bytes of 0x01
		MasterKey: "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE=",
	})
	require.NoError(t, err)

	encoded, err := svc1.Encode("some.jwt")
	require.NoError(t, err)

	_, err = svc2.Decode(encoded)
	assert.Error(t, err, "different key should not be able to decode")
}
