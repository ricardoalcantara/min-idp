#!/usr/bin/env bash
# Register oidc-test-sp in min-idp.
#
# Usage:
#   TOKEN=<admin-jwt> ./register-sp.sh
#   IDP_URL=http://localhost:8081 TOKEN=<jwt> ./register-sp.sh
set -euo pipefail

: "${TOKEN:?Set TOKEN to a min-idp admin API bearer token}"
IDP_URL="${IDP_URL:-http://localhost:8081}"

echo "Creating SP..."
RESP=$(curl -fsS -X POST "${IDP_URL}/api/admin/sps" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d @- <<'EOF'
{
  "slug": "oidc-test-sp",
  "name": "OIDC Test SP (Next.js)",
  "protocol": "oidc"
}
EOF
)

SP_ID=$(printf '%s' "$RESP" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$SP_ID" ]; then
  echo "Failed to create SP:" >&2
  echo "$RESP" >&2
  exit 1
fi

echo "Configuring OIDC client..."
curl -fsS -X PUT "${IDP_URL}/api/admin/sps/${SP_ID}/oidc" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d @- <<'EOF'
{
  "client_id": "oidc-test-sp",
  "client_secret": "super-secret",
  "redirect_uris": [
    "http://localhost:3001/api/auth/oidc/callback"
  ],
  "grant_types": [
    "authorization_code"
  ],
  "response_types": [
    "code"
  ],
  "scopes": [
    "openid",
    "email",
    "profile"
  ],
  "token_endpoint_auth": "client_secret_basic",
  "pkce_required": true
}
EOF

echo "✓ oidc-test-sp registered (id=${SP_ID})"
