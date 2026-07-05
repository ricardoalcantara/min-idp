#!/usr/bin/env bash
# Register oidc-public-test-sp in min-idp.
#
# Usage:
#   TOKEN=<admin-jwt> ./register-sp.sh
#   IDP_URL=http://192.168.1.107:8081 TOKEN=<jwt> ./register-sp.sh
#
# Set on min-idp: MINSTACK_CORS_ORIGIN=http://192.168.1.107:5173
set -euo pipefail

: "${TOKEN:?Set TOKEN to a min-idp admin API bearer token}"
IDP_URL="${IDP_URL:-http://192.168.1.107:8081}"

echo "Creating SP..."
RESP=$(curl -fsS -X POST "${IDP_URL}/api/admin/sps" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d @- <<'EOF'
{
  "slug": "oidc-public-test-sp",
  "name": "OIDC Public Test SP (Vite)",
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

echo "Configuring OIDC public client..."
curl -fsS -X PUT "${IDP_URL}/api/admin/sps/${SP_ID}/oidc" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d @- <<'EOF'
{
  "client_id": "oidc-public-test-sp",
  "redirect_uris": [
    "http://192.168.1.107:5173/callback"
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
  "token_endpoint_auth": "none"
}
EOF

ROOT="$(cd "$(dirname "$0")" && pwd)"
if [ ! -f "${ROOT}/.env" ]; then
  cp "${ROOT}/.env.example" "${ROOT}/.env"
  echo "Created ${ROOT}/.env from .env.example"
fi

echo "✓ oidc-public-test-sp registered (id=${SP_ID})"
