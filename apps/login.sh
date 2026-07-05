#!/usr/bin/env bash
# Log in to min-idp and print an admin API token.
#
# Usage:
#   ./apps/login.sh
#   IDP_URL=http://192.168.1.107:8081 ./apps/login.sh
#
# Then:
#   export TOKEN=<printed-token>
#   ./apps/register-all-sps.sh
set -euo pipefail

IDP_URL="${IDP_URL:-http://192.168.1.107:8081}"

read -r -p "Login: " LOGIN
read -rs -p "Password: " PASSWORD
echo

if [ -z "$LOGIN" ] || [ -z "$PASSWORD" ]; then
  echo "Login and password are required." >&2
  exit 1
fi

BODY=$(python3 -c 'import json,sys; print(json.dumps({"login": sys.argv[1], "password": sys.argv[2]}))' \
  "$LOGIN" "$PASSWORD")

echo "Logging in to ${IDP_URL}..."
RESP=$(curl -fsS -X POST "${IDP_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "$BODY")

TOKEN=$(printf '%s' "$RESP" | grep -o '"access_token":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$TOKEN" ]; then
  echo "Login failed:" >&2
  echo "$RESP" >&2
  exit 1
fi

echo "✓ Login successful"
echo ""
echo "export TOKEN=${TOKEN}"
