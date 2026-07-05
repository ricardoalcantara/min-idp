#!/usr/bin/env bash
# Register all test SP apps.
#
# Usage:
#   TOKEN=<jwt> ./apps/register-all-sps.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"

for app in oidc-test-sp oidc-public-test-sp saml-test-sp; do
  echo "--- ${app} ---"
  "${ROOT}/${app}/register-sp.sh"
  echo ""
done

echo "✓ All test SPs registered."
