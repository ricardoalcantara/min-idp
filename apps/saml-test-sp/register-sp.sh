#!/usr/bin/env bash
# Register saml-test-sp in min-idp.
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
  "slug": "saml-test-sp",
  "name": "SAML Test SP",
  "protocol": "saml"
}
EOF
)

SP_ID=$(printf '%s' "$RESP" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$SP_ID" ]; then
  echo "Failed to create SP:" >&2
  echo "$RESP" >&2
  exit 1
fi

echo "Configuring SAML client..."
curl -fsS -X PUT "${IDP_URL}/api/admin/sps/${SP_ID}/saml" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d @- <<'EOF'
{
  "entity_id": "http://localhost:3002",
  "acs_urls": [
    "http://localhost:3002/api/auth/saml/callback"
  ],
  "slo_url": "http://localhost:3002/api/auth/saml/logout",
  "name_id_format": "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress",
  "want_signed_requests": false,
  "want_signed_assertions": true
}
EOF

echo "✓ saml-test-sp registered (id=${SP_ID})"
