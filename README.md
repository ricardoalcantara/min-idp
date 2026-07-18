# min-idp

A minimal, self-hosted Identity Provider supporting **OIDC/OAuth 2.0** and **SAML 2.0**.
Built with Go, [go-minstack](https://github.com/go-minstack), Gin, and GORM.

→ [Roadmap](ROADMAP.md)

---

## Prerequisites

- Go 1.22+
- SQLite (default), MySQL, or PostgreSQL

---

## Quick start (local)

### 1. Configure

```bash
cp .env.example .env
```

Generate and set a master key (required):

```bash
openssl rand -base64 32
```

Put the value in `.env`:

```bash
MIN_IDP_MASTER_KEY=<paste-here>
```

Optional but useful for local work — pin the admin password so it is stable across DB resets:

```bash
MIN_IDP_ADMIN_PASSWORD=change-me
```

`.env.example` already pins a default OIDC app for local predictability:

| Variable | Example value |
|---|---|
| `MIN_IDP_BOOTSTRAP_SP_CLIENT_ID` | `default` |
| `MIN_IDP_BOOTSTRAP_SP_CLIENT_SECRET` | `dev-client-secret-change-me` |
| `MIN_IDP_BOOTSTRAP_SP_REDIRECT_URIS` | `http://localhost:3000/callback` |

Adjust the redirect URI(s) to match your app before the first run (comma-separated for multiple).

### 2. Run

```bash
make run
```

Or:

```bash
make build
./min-idp
```

By default (from `.env.example`) the IdP listens on **http://localhost:8080**.

### 3. First-run bootstrap

On a fresh database, min-idp automatically:

1. Creates signing keys (OIDC + SAML)
2. Creates system roles (`system:admin`, `sp:login`, `api:user`)
3. Creates the admin user
4. Creates a default OIDC application (`slug=default`)

Credentials are printed **once** in the startup log:

```json
{
  "message": "bootstrap complete — change the admin password immediately",
  "email": "admin@min-idp.local",
  "password": "<admin-password>",
  "sp_slug": "default",
  "client_id": "default",
  "client_secret": "dev-client-secret-change-me"
}
```

- Admin email is always `admin@min-idp.local`
- Admin password comes from `MIN_IDP_ADMIN_PASSWORD`, or a random value if unset
- Client id/secret come from `MIN_IDP_BOOTSTRAP_SP_*`, or random secret if the secret env is unset
- For SPAs, set `MIN_IDP_BOOTSTRAP_SP_PUBLIC=true` — the app is created as a public client (`token_endpoint_auth=none`, PKCE required, no secret) and the log shows `client_type` instead of a secret

Bootstrap runs only once per database (tracked in `bootstrap_states`). Deleting the DB (or the SQLite file) and restarting re-runs it.

### 4. Point your app at min-idp

You can use the bootstrapped client immediately — no admin API calls required.

| Setting | Value |
|---|---|
| Issuer / discovery | `http://localhost:8080` |
| Discovery URL | `http://localhost:8080/.well-known/openid-configuration` |
| JWKS | `http://localhost:8080/.well-known/jwks.json` |
| Client ID | `default` (or your `MIN_IDP_BOOTSTRAP_SP_CLIENT_ID`) |
| Client Secret | value from log / `MIN_IDP_BOOTSTRAP_SP_CLIENT_SECRET` |
| Redirect URI | must match `MIN_IDP_BOOTSTRAP_SP_REDIRECT_URIS` |
| Scopes | `openid` (profile/email as needed) |

Authorize endpoint: `http://localhost:8080/oauth2/authorize`  
Token endpoint: `http://localhost:8080/oauth2/token`

Sign in at the IdP login page with the admin user (or any user that has the `sp:login` role). With no access rules on the SP, anyone with `sp:login` can SSO.

### 5. Admin login (API)

```bash
curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"login":"admin@min-idp.local","password":"<admin-password>"}' \
  | jq -r .access_token
```

Use the token as `Authorization: Bearer <token>` on `/api/admin/*` routes.

---

## Docker

A pre-built image is published to GitHub Container Registry on every release tag.

```bash
docker pull ghcr.io/ricardoalcantara/min-idp:v0.1.0-alpha
```

**`docker-compose.yml`** provides a ready-to-use stack with a persistent SQLite volume (port **8081**).  
**`docker-compose.override.yml`** is the place to set secrets and environment-specific values — it is git-ignored and merged automatically by Docker Compose.

```yaml
# docker-compose.override.yml
services:
  min-idp:
    environment:
      MIN_IDP_EXTERNAL_URL: https://your-domain.example.com
      MIN_IDP_MASTER_KEY: "your-key-here"
      # Optional: pin bootstrap credentials
      MIN_IDP_ADMIN_PASSWORD: "change-me"
      MIN_IDP_BOOTSTRAP_SP_CLIENT_ID: default
      MIN_IDP_BOOTSTRAP_SP_CLIENT_SECRET: "change-me-too"
      MIN_IDP_BOOTSTRAP_SP_REDIRECT_URIS: https://your-app.example.com/callback
```

Generate a key with:

```bash
openssl rand -base64 32
```

Then start:

```bash
docker compose up -d
docker compose logs -f min-idp   # look for the bootstrap complete line
```

---

## Configuration

Copy and edit `.env` (see [`.env.example`](.env.example) for the full list).

| Variable | Description | Default |
|---|---|---|
| `MINSTACK_PORT` | HTTP listen port | `8080` (example) / `8081` (Docker) |
| `MINSTACK_CORS_ORIGIN` | Allowed CORS origins (needed for public/SPA clients) | unset |
| `MINSTACK_LOG_LEVEL` | `debug`, `info`, `warn`, `error` | `info` |
| `MIN_IDP_EXTERNAL_URL` | Public base URL (OIDC issuer, SAML metadata) | **required** |
| `MIN_IDP_DB_DRIVER` | `sqlite`, `mysql`, or `postgres` | `sqlite` |
| `MIN_IDP_DB_DSN` | Database connection string | `./dev.db` |
| `MIN_IDP_MASTER_KEY` | AES key for encrypting signing keys at rest (`openssl rand -base64 32`) | **required** |
| `MIN_IDP_ADMIN_PASSWORD` | Pin admin password on first bootstrap; empty → random | empty |
| `MIN_IDP_BOOTSTRAP_SP_NAME` | Display name of the default OIDC app | `Default App` |
| `MIN_IDP_BOOTSTRAP_SP_REDIRECT_URIS` | Comma-separated redirect URIs | `http://localhost:3000/callback` |
| `MIN_IDP_BOOTSTRAP_SP_CLIENT_ID` | Client id; empty → `default` | empty → `default` |
| `MIN_IDP_BOOTSTRAP_SP_CLIENT_SECRET` | Pin client secret; empty → random (recommended in production) | empty |
| `MIN_IDP_BOOTSTRAP_SP_PUBLIC` | `true` = public client for SPAs (PKCE, no secret; secret var ignored) | `false` |

---

## Adding more Service Providers

The default OIDC app is enough to get started. Use the admin API when you need additional apps or SAML.

All admin endpoints are under `/api/admin/` and require a Bearer token from `/api/auth/login`.

SP identifiers in API paths are **UUIDs** (the `id` field returned when you create an SP), not numeric IDs.

### OIDC SP (confidential)

**Step 1 — Create the SP**

```bash
curl -X POST http://localhost:8080/api/admin/sps \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"slug":"my-app","name":"My Application","protocol":"oidc"}'
```

**Step 2 — Configure the OIDC client**

```bash
curl -X PUT http://localhost:8080/api/admin/sps/<sp-uuid>/oidc \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "my-app-client",
    "client_secret": "super-secret",
    "redirect_uris": ["http://localhost:3001/api/auth/callback"],
    "post_logout_redirect_uris": ["http://localhost:3001/"],
    "grant_types": ["authorization_code"],
    "response_types": ["code"],
    "scopes": ["openid","email","profile"],
    "token_endpoint_auth": "client_secret_basic",
    "pkce_required": true
  }'
```

**Step 3 — Point your app at min-idp**

| Setting | Value |
|---|---|
| Issuer | `http://localhost:8080` |
| Client ID | `my-app-client` |
| Client Secret | `super-secret` |
| Redirect URI | as registered above |

#### OIDC public client (SPA)

For browser-only apps (no client secret), register a **public** client with PKCE:

```bash
curl -X PUT http://localhost:8080/api/admin/sps/<sp-uuid>/oidc \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "my-spa",
    "redirect_uris": ["http://localhost:5173/callback"],
    "post_logout_redirect_uris": ["http://localhost:5173/"],
    "grant_types": ["authorization_code"],
    "response_types": ["code"],
    "scopes": ["openid","email","profile"],
    "token_endpoint_auth": "none"
  }'
```

`token_endpoint_auth: "none"` forces PKCE and does not allow a `client_secret`.

The browser must call the token endpoint directly, so add the SPA origin to CORS:

```
MINSTACK_CORS_ORIGIN=http://localhost:5173
```

See `apps/oidc-public-test-sp/` for a working Vite + React example.

### SAML SP

**Step 1 — Create the SP**

```bash
curl -X POST http://localhost:8080/api/admin/sps \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"slug":"saml-test-sp","name":"SAML Test SP","protocol":"saml"}'
```

**Step 2 — Configure the SAML provider**

```bash
curl -X PUT http://localhost:8080/api/admin/sps/<sp-uuid>/saml \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "http://localhost:3002",
    "acs_urls": ["http://localhost:3002/api/auth/saml/callback"],
    "slo_url": "http://localhost:3002/api/auth/saml/logout",
    "name_id_format": "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress",
    "want_signed_assertions": true,
    "want_signed_requests": false
  }'
```

**Step 3 — Point your SP at min-idp**

| Endpoint | URL |
|---|---|
| IdP SSO (Redirect) | `http://localhost:8080/saml/sso` |
| IdP SLO | `http://localhost:8080/saml/slo` |
| SP Metadata | `http://localhost:8080/saml/metadata` |
| IdP Entity ID | `http://localhost:8080` |

Configure your SP library with these values and the IdP signing certificate from:

```
http://localhost:8080/api/admin/keys
```

---

## Access Control

Users need the `sp:login` role (or `sp:login:<slug>` for a specific app) to SSO.

The bootstrapped admin already has `sp:login`. With **no** access rules on an SP, anyone with that role is allowed (minimum-effort default). Once you add rules, unmatched subjects are denied.

To restrict access:

1. Create a role under `/api/admin/roles`
2. Assign users under `/api/admin/users/:id/roles`
3. Create an access rule under `/api/admin/sps/:id/access-rules` targeting the role (or user)

---

## Signing Keys

min-idp manages signing keys for OIDC (JWTs) and SAML (assertions) automatically on bootstrap.
To rotate a key:

```bash
curl -X POST http://localhost:8080/api/admin/keys/oidc/rotate \
  -H "Authorization: Bearer <token>"
```

OIDC public keys are published at `/.well-known/jwks.json`.

---

## Test Apps

Sample SPs under `apps/`:

| App | Protocol | Port | README |
|---|---|---|---|
| `oidc-test-sp` | OIDC confidential (Next.js) | 3001 | `apps/oidc-test-sp/` |
| `oidc-public-test-sp` | OIDC public SPA (Vite + React) | 5173 | `apps/oidc-public-test-sp/` |
| `saml-test-sp` | SAML 2.0 (node-saml) | 3002 | `apps/saml-test-sp/` |

Each has a `.env.local.example` — copy it to `.env.local` and fill in the IdP values (issuer, client id/secret, redirect URI). You can either use the bootstrapped `default` client (update its redirect URI via env before first run, or create a dedicated SP via the admin API as shown above).
