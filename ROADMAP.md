# min-idp Roadmap

This document tracks planned features. Items are grouped by theme, not strict priority — implementation order will depend on what's most needed at the time.

---

## Tier 1 — Security & Auth Hardening

### Email notifications (foundation)
SMTP configuration, email templates, and a send-email service. This is the prerequisite for password recovery, magic link login, and security alerts (e.g. login from a new IP).

### Password recovery
"Forgot password" flow: user submits their email, receives a time-limited reset link, sets a new password. Depends on email foundation.

### Magic link / email OTP login
Passwordless login: user enters email and receives a single-use login link. An alternative to passwords — useful for apps where you want low-friction access without requiring users to remember credentials.

### TOTP / 2FA
Time-based one-time passwords compatible with standard authenticator apps (Google Authenticator, Authy, 1Password, etc.). Users enroll a TOTP device via a `otpauth://` QR code; subsequent logins require password + 6-digit code.

---

## Tier 2 — Developer Experience

### Swagger UI at /docs
Serve the `docs/openapi.yml` spec through a built-in Swagger UI endpoint so developers can explore and test the API directly from the IdP without any external tooling.

### Admin UI
A web interface for managing users, roles, service providers, and signing keys. Currently everything requires direct API calls. A simple admin panel would make min-idp usable without an API client.

### Access rules UI
Dedicated UI for managing per-SP allow/deny rules (currently API-only).

### User profile page
A browser page where a logged-in user can view and update their own profile (name, username, email) without using the API.

---

## Tier 3 — Social Identity Providers

Allow users to log into min-idp using an upstream provider (Google, GitHub, Microsoft, etc.) instead of a local password. min-idp handles the OIDC/OAuth2 exchange with the upstream provider and maps the identity to a local account.

This is upstream identity *federation* — not multi-tenancy. The identity record stays in min-idp; the upstream provider just replaces the password step.

---

## Tier 4 — Quality & Observability

### Unit test coverage
Increase service layer coverage to 75%. Priority targets: OIDC service, SAML service, session middleware, RBAC.

### E2E tests
Full SSO flow tests using `httptest` + real SQLite against both OIDC and SAML endpoints.

### Audit log improvements
Richer event types, filtering by user/SP/action in the API, pagination, and a configurable retention policy.
