# min-idp — Architecture Specification

A minimal, single-tenant, self-hosted Identity Provider written in Go on top of the [MinStack](https://www.go-minstack.com/) (Uber FX-based) framework. API-first. Multiple Service Providers, single user base, OIDC and SAML on day one, extension points for SMTP, LDAP, TOTP, and federation.

---

## 1. Goals

- **Single-tenant**: one organization, one user base, multiple SPs.
- **API-centred**: no admin GUI initially; admin UI is a future client of the same API.
- **Simple to bootstrap**: first run generates keys, schema, and root admin; prints credentials to stdout.
- **Pluggable storage**: SQLite, MySQL, PostgreSQL.
- **Pluggable cache (KV)**: DB-backed by default; Redis (or any external KV) via interface injection.
- **Protocols on day one**: OIDC + OAuth2, SAML 2.0.
- **Future-ready**: SMTP (notifications), LDAP (directory/federation), TOTP, external IdP federation.

## 2. Non-Goals (initial release)

- No admin GUI.
- No multi-tenancy.
- No HA/clustering specifics (DB-backed state means horizontal scaling is possible, but not a target here).
- No webhooks / event bus (audit log first; webhooks later if needed).
- No built-in rate limiting (add later as middleware).

---

## 3. Module Layout (MinStack DI)

Each directory is a self-contained FX module exposing a single `Module()` function. `extensions/*` plug in via interfaces; the core does not depend on them.

```
cmd/min-idp/main.go         # composes modules, calls fx.New()
internal/
  config/                   # ENV first, YAML later
  storage/                  # DB abstraction (sqlite/mysql/postgres)
    migrations/             # versioned SQL migrations
    repo/                   # per-entity repos (users, sps, sessions, ...)
  kvstore/                  # KVStore interface, DB-backed impl
  crypto/                   # key generation, JWK, X.509, signing, hashing
  keystore/                 # signing key lifecycle (rotation, current/previous)
  user/                     # user domain (CRUD, password, roles, groups)
  rbac/                     # roles, permissions, policy evaluation
  session/                  # IdP-level session (cookies, SSO state)
  audit/                    # append-only event log
  protocol/
    oidc/                   # OIDC + OAuth2 endpoints
    saml/                   # SAML 2.0 endpoints
  sp/                       # service provider registry (both protocols)
  api/                      # admin + auth REST handlers
  bootstrap/                # first-run: keys, root admin, schema
extensions/                 # future plug-ins (smtp, ldap, totp, federation)
```

---

## 4. Data Model

Core tables. All timestamps are UTC. All IDs are UUIDs unless otherwise noted.

```
users
  id, email, password_hash, status (active|disabled|locked),
  created_at, updated_at

user_credentials
  user_id, type (password|totp|webauthn), data, created_at
  -- TOTP/WebAuthn here when added; keeps users table clean.

roles
  id, name, description, system (bool)

permissions
  id, name        -- e.g. "system:admin", "sp:login", "sp:login:<slug>"

role_permissions
  role_id, permission_id

user_roles
  user_id, role_id

groups
  id, name, description

user_groups
  user_id, group_id

service_providers
  id, slug, name, protocol (oidc|saml), enabled, created_at

oidc_clients
  sp_id, client_id, client_secret_hash, redirect_uris,
  grant_types, response_types, scopes, token_endpoint_auth,
  pkce_required

saml_clients
  sp_id, entity_id, acs_urls, slo_url, name_id_format,
  sp_certificate, want_signed_requests, want_signed_assertions

sp_access_rules
  sp_id, rule_type (allow|deny),
  subject_type (group|role|user), subject_id, priority
  -- evaluated in priority order; default deny if no allow matches.

sessions
  id, user_id, created_at, expires_at, last_seen_at,
  ip, user_agent, revoked_at

sp_sessions
  session_id, sp_id, sub, created_at
  -- tracks which SPs participated in this SSO session (for SLO).

oauth_codes
  code_hash, client_id, user_id, session_id, redirect_uri,
  scope, code_challenge, code_challenge_method,
  nonce, expires_at, used_at

oauth_tokens
  id, type (access|refresh), token_hash, client_id,
  user_id, session_id, scope, expires_at, revoked_at,
  parent_id   -- refresh-rotation chain

saml_requests
  id, sp_id, request_id, relay_state, expires_at
  -- replay prevention + AuthnRequest tracking.

signing_keys
  id, protocol (oidc|saml), kid, algorithm,
  private_key_encrypted, public_key, certificate (saml only),
  status (active|previous|retired),
  created_at, activated_at, retired_at

audit_events
  id, ts, actor_user_id, action, target_type, target_id,
  sp_id, ip, user_agent, result, metadata_json

kv_store
  key (PK), value (blob), expires_at (nullable), created_at
  -- generic KV; transparent to callers via KVStore interface.

bootstrap_state
  key, value          -- "initialized=true", schema version, etc.
```

---

## 5. KVStore Interface

Cache and ephemeral state share a single abstraction. Default implementation is DB-backed using the `kv_store` table; a Redis implementation can be injected without touching callers.

```go
type KVStore interface {
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Get(ctx context.Context, key string) ([]byte, error)
    Delete(ctx context.Context, key string) error
    SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)
}
```

**DB-backed implementation**:
- Single `kv_store` table.
- `expires_at` indexed; janitor goroutine deletes expired rows every N seconds.
- `SetNX` via conditional INSERT.

**Redis implementation** (future): same interface, native commands.

Used for: OAuth auth codes (in addition to `oauth_codes` table — TBD whether KV or table is canonical; current plan keeps both because codes are auditable), PKCE challenges, SAML AuthnRequest IDs, nonces, replay guards, rate-limit counters.

---

## 6. Authentication Flows

### 6.1 IdP Session (the SSO anchor)

A successful login (any path) creates a row in `sessions` and sets a session cookie on the IdP domain. This cookie is what enables SSO across SPs: subsequent SP redirects to `/oauth2/authorize` or `/saml/sso` find the cookie and skip re-authentication.

### 6.2 Direct API login (no redirect — admin / API clients)

```
POST /api/auth/login
  body: { email, password }
  -> sets session cookie + optionally returns API bearer token
POST /api/auth/logout
POST /api/auth/register   (gated by feature flag)
GET  /api/me
PATCH /api/me
GET  /api/me/sessions
DELETE /api/me/sessions/{id}
```

### 6.3 OIDC / OAuth2

```
GET  /oauth2/authorize
POST /oauth2/token
GET  /oauth2/userinfo
POST /oauth2/revoke
POST /oauth2/introspect
GET  /oauth2/logout                         # RP-initiated logout
GET  /.well-known/openid-configuration
GET  /.well-known/jwks.json
```

`/oauth2/authorize` flow:
1. Validate client, redirect_uri, scopes, PKCE.
2. If IdP session exists and access rules allow: issue code, redirect.
3. Otherwise: stash request in KV (state cookie), render `/login`, on success resume.

### 6.4 SAML 2.0

```
GET/POST /saml/sso         # AuthnRequest (Redirect or POST binding)
GET/POST /saml/slo         # SLO request
GET      /saml/metadata    # IdP metadata (one global doc)
```

Same session-based gating as OIDC.

### 6.5 Login UI (minimal)

A single HTML template at `/login`. POSTs to `/api/auth/login`, then redirects back to the original protocol request (resumed from the KV-stashed state). No JS framework. Replaceable when a real GUI is built.

---

## 7. Access Control

### 7.1 Permissions (system-level)

- `system:admin` — full admin API access.
- `sp:login` — global gate: user can perform SSO to any SP (subject to per-SP rules).
- `sp:login:<slug>` — fine-grained: user can SSO only to specific SPs (future; same evaluation pipeline).
- `api:user` — direct API auth (if API login is enabled).

A user with only `system:admin` and no `sp:login` is admin-only and cannot SSO anywhere.

### 7.2 Per-SP rules (`sp_access_rules`)

When a user attempts SSO to SP-X:

1. Check user has `sp:login` (or `sp:login:<X.slug>`).
2. Walk `sp_access_rules` for SP-X in priority order; match against user's roles, groups, and ID.
3. Default deny if no allow matches; per-SP `default_action` may flip this to default-allow.

Examples:
- `allow group=engineering` → only engineering can use this SP.
- `deny group=contractors` + `allow *` → everyone except contractors.

---

## 8. Bootstrap (first run)

`bootstrap` module runs at startup, idempotent.

1. Run migrations.
2. If `bootstrap_state.initialized != true`:
   - Generate OIDC signing keypair (RS256 or ES256). Insert as `active`.
   - Generate SAML signing keypair + self-signed X.509 cert. Insert as `active`.
   - Create system roles: `system:admin`, `sp:login`, `api:user`.
   - Create root admin user with random password; assign `system:admin` + `sp:login`.
   - **Print to stdout**: email, password, message instructing the operator to change it.
   - Set `initialized=true`.
3. On every start: load active keys into `keystore`.

### 8.1 Master key

Private signing keys are encrypted at rest using AES-GCM with a master key from env (`MIN_IDP_MASTER_KEY`). Bootstrap fails loudly if the env var is missing.

---

## 9. Key Rotation

`signing_keys.status` enum:

- `active` — current signing key (one per protocol).
- `previous` — old key, still trusted for verification during a grace period.
- `retired` — no longer published.

JWKS endpoint publishes `active` + `previous`. SAML metadata publishes both certs.

Admin endpoint:

```
POST /api/admin/keys/{protocol}/rotate
```

generates a new `active`, demotes the old to `previous`, retires anything older. Future: scheduled automatic rotation.

---

## 10. Admin API

```
/api/admin/users                       CRUD
/api/admin/users/{id}/roles            POST, DELETE
/api/admin/users/{id}/groups           POST, DELETE
/api/admin/users/{id}/sessions         GET, DELETE     (force logout)
/api/admin/users/{id}/password         POST            (admin reset)

/api/admin/roles                       CRUD
/api/admin/roles/{id}/permissions      POST, DELETE
/api/admin/groups                      CRUD

/api/admin/sps                         CRUD
/api/admin/sps/{id}/oidc               GET, PUT
/api/admin/sps/{id}/saml               GET, PUT
/api/admin/sps/{id}/access-rules       CRUD

/api/admin/keys/{protocol}             GET             (list)
/api/admin/keys/{protocol}/rotate      POST

/api/admin/audit                       GET             (filter, paginate)
```

No version segment in the URL (`/api/admin/...`, not `/api/v1/admin/...`); versioning will track the `min-idp` release.

---

## 11. Extension Points

Interfaces are defined in v1; implementations come later. Adding an extension is a matter of registering a new module in `cmd/min-idp/main.go`.

```go
// internal/user
type CredentialVerifier interface {
    Type() string                              // "password", "totp", "ldap"
    Verify(ctx context.Context, user *User, input any) error
}

// internal/notification
type Notifier interface {
    Send(ctx context.Context, kind NotificationKind, recipient string, payload any) error
}
// Default: no-op notifier. SMTP module replaces it.

// internal/federation
type IdentityProvider interface {
    Authenticate(ctx context.Context, request *AuthnRequest) (*ExternalIdentity, error)
}

// internal/user
type Provisioner interface {
    Provision(ctx context.Context, identity *ExternalIdentity) (*User, error)
}
```

Login becomes a chain of `CredentialVerifier`s. TOTP = a second verifier plus policy `require_factors >= 2`. LDAP = an `IdentityProvider` plus a `Provisioner` that creates/updates local users from the directory.

---

## 12. Configuration

ENV-first. YAML support deferred.

Required:

```
MIN_IDP_LISTEN              ":8080"
MIN_IDP_EXTERNAL_URL        "https://idp.example.com"
MIN_IDP_DB_DRIVER           "sqlite" | "mysql" | "postgres"
MIN_IDP_DB_DSN              "..."
MIN_IDP_MASTER_KEY          "<base64 32 bytes>"
```

Optional:

```
MIN_IDP_KV_DRIVER           "db" (default) | "redis"
MIN_IDP_KV_REDIS_URL        "redis://..."
MIN_IDP_SESSION_COOKIE      "min_idp_session"
MIN_IDP_SESSION_TTL         "12h"
MIN_IDP_SESSION_IDLE        "1h"
MIN_IDP_FEATURE_API_REGISTRATION  "false"
MIN_IDP_FEATURE_API_LOGIN         "true"
```

SPs, users, keys, roles, groups: all live in the DB.

---

## 13. Audit Logging

Every state-changing action and every authentication attempt writes an `audit_events` row. Captured: actor, action, target, SP (if applicable), IP, user-agent, result, free-form metadata. Append-only; no UPDATE or DELETE in normal operation. Exposed via `/api/admin/audit`.

---

## 14. Open Items / Deferred

- Rate limiting (middleware, KV-backed counters).
- Webhooks / outbound event delivery.
- Scheduled automatic key rotation.
- YAML config layered over ENV.
- Multi-region / HA deployment notes.
- TOTP, WebAuthn, LDAP, SMTP, external federation (interfaces ready; implementations later).
- Canonical store for OAuth codes: `oauth_codes` table vs. `kv_store` (decide during implementation).

---

## 15. Naming Convention Summary

- Module name: `min-idp`.
- Routes: `/api/admin/...`, `/api/auth/...`, `/api/me`, `/oauth2/...`, `/saml/...`, `/.well-known/...`.
- DB tables: snake_case, plural where appropriate (`users`, `oidc_clients`, `kv_store`).
- Permissions: `domain:action` or `domain:action:scope` (`sp:login:my-app`).