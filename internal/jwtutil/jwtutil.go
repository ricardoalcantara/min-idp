package jwtutil

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

// PayloadClaims decodes the JWT payload without verifying the signature.
// Use only when the signature has already been verified (e.g. by middleware),
// or when only display data is needed (logout hint, info page).
func PayloadClaims(tokenStr string) (map[string]any, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid jwt format")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	return claims, nil
}
