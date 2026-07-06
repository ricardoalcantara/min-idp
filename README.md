# min-idp

A minimal, self-hosted Identity Provider supporting **OIDC/OAuth 2.0** and **SAML 2.0**.
Built with Go, [go-minstack](https://github.com/go-minstack), Gin, and GORM.

→ [Roadmap](ROADMAP.md)

---

## Prerequisites

- Go 1.22+
- SQLite (default) or PostgreSQL

---

## Configuration

Copy and edit the environment file:

```bash
cp .env.example .env
```

| Variable | Description | Default |
|---|---|---|
| `MINSTACK_PORT` | HTTP listen port | `8081` |
| `MINSTACK_LOG_LEVEL` | Log level (`debug`, `info`, `warn`, `error`) | `info` |
| `MIN_IDP_EXTERNAL_URL` | Public base URL (used in OIDC discovery, SAML metadata) | required |
| `MIN_IDP_DB_DRIVER` | `sqlite` or `postgres` | `sqlite` |
| `MIN_IDP_DB_DSN` | Database connection string | `./dev.db` |
| `MIN_IDP_MASTER_KEY` | AES-256 key for session encryption — generate with `openssl rand -base64 32` | required |

---

## Running

```bash
# Development
make run

# Production build
make build
./min-idp
```

## Docker

A pre-built image is published to GitHub Container Registry on every release tag.

```bash
docker pull ghcr.io/ricardoalcantara/min-idp:v0.1.0-alpha
```

**`docker-compose.yml`** provides a ready-to-use stack with a persistent SQLite volume.  
**`docker-compose.override.yml`** is the place to set secrets and environment-specific values — it is git-ignored and merged automatically by Docker Compose.

```yaml
# docker-compose.override.yml
services:
  min-idp:
    environment:
      MIN_IDP_EXTERNAL_URL: https://your-domain.example.com
      MIN_IDP_MASTER_KEY: "your-key-here"
```

Generate a key with:

```bash
openssl rand -base64 32
```

Then start:

```bash
docker compose up -d
```

On first start, min-idp bootstraps the database and creates a default admin account:

- **Email:** `admin@min-idp.local`
- **Password:** `admin` *(change immediately)*

---

## Admin API

All admin endpoints are under `/api/admin/` and require a Bearer token.

### 1. Obtain a token

```bash
curl -s -X POST http://localhost:8081/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@min-idp.local","password":"admin"}' \
  | jq -r .access_token
```

Use the returned token as `Authorization: Bearer <token>` on all subsequent admin requests.

---

## Setting up a Service Provider

### OIDC SP

**Step 1 — Create the SP**

```bash
curl -X POST http://localhost:8081/api/admin/sps \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"slug":"my-app","name":"My Application","protocol":"oidc"}'
```

Note the `id` in the response (e.g. `1`).

**Step 2 — Configure the OIDC client**

```bash
curl -X PUT http://localhost:8081/api/admin/sps/1/oidc \
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

OIDC discovery is available at:

```
http://localhost:8081/.well-known/openid-configuration
```

Configure your client library with:

| Setting | Value |
|---|---|
| Issuer | `http://localhost:8081` |
| Client ID | `my-app-client` |
| Client Secret | `super-secret` |
| Redirect URI | as registered above |

#### OIDC public client (SPA)

For browser-only apps (no client secret), register a **public** client with PKCE:

```bash
curl -X PUT http://localhost:8081/api/admin/sps/1/oidc \
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

---

### SAML SP

**Step 1 — Create the SP**

```bash
curl -X POST http://localhost:8081/api/admin/sps \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"slug":"saml-test-sp","name":"SAML Test SP","protocol":"saml"}'
```

Note the `id` in the response.

**Step 2 — Configure the SAML provider**

```bash
curl -X PUT http://localhost:8081/api/admin/sps/1/saml \
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
| IdP SSO (Redirect) | `http://localhost:8081/saml/sso` |
| IdP SLO | `http://localhost:8081/saml/slo` |
| SP Metadata | `http://localhost:8081/saml/metadata` |
| IdP Entity ID | `http://localhost:8081` |

Configure your SP library with these values and the IdP signing certificate from:

```
http://localhost:8081/api/admin/keys
```

---

## Access Control

By default, any authenticated user can SSO into any configured SP. To restrict access:

1. Create a role under `/api/admin/roles`
2. Assign users to the role under `/api/admin/users/:id/roles`
3. Create an access rule under `/api/admin/sps/:id/access-rules` targeting the role

---

## Signing Keys

min-idp manages signing keys for OIDC (JWTs) and SAML (assertions) automatically.
To rotate a key:

```bash
curl -X POST http://localhost:8081/api/admin/keys/oidc/rotate \
  -H "Authorization: Bearer <token>"
```

OIDC public keys are published at `/api/oidc/jwks` for client verification.

---

## Test Apps

Two test SP apps are included under `apps/`:

| App | Protocol | Port | README |
|---|---|---|---|
| `oidc-test-sp` | OIDC confidential (Next.js) | 3001 | `apps/oidc-test-sp/` |
| `oidc-public-test-sp` | OIDC public SPA (Vite + React) | 5173 | `apps/oidc-public-test-sp/` |
| `saml-test-sp` | SAML 2.0 (node-saml) | 3002 | `apps/saml-test-sp/` |

Each has a `.env.local.example` — copy it to `.env.local` and fill in the IdP values.
