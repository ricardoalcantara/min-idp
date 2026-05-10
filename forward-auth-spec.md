# Mini-IdP Forward-Auth Specification

A minimal Identity Provider supporting forward-auth (subrequest auth) for reverse proxies. This spec defines the per-proxy endpoint contracts and shared session/identity model.

## Overview

Forward-auth is a pattern where a reverse proxy makes a subrequest to an external service before forwarding the original request upstream. The auth service returns either a success (with identity headers) or an unauth response (with redirect or 401).

This IdP must expose **per-proxy endpoints** because each reverse proxy implements the pattern with a different request/response contract.

## Endpoint Contracts

### `/auth/traefik` — Traefik `forwardAuth`

**Request from proxy:**
- Method preserved (GET, POST, etc.)
- Original request metadata forwarded as headers:
  - `X-Forwarded-Method`
  - `X-Forwarded-Proto`
  - `X-Forwarded-Host`
  - `X-Forwarded-Uri`
  - `X-Forwarded-For`
- Cookies forwarded as-is
- Request body is **not** forwarded

**Response to proxy:**
| Scenario | Status | Notes |
|---|---|---|
| Authenticated | `200` | Return identity headers; proxy copies them upstream via `authResponseHeaders` |
| Unauthenticated | `302` | Redirect to login; Traefik follows the redirect |
| Forbidden | `403` | Hard block, no redirect |

**Key trait:** Traefik follows redirects, so the IdP handles the redirect-to-login flow directly.

---

### `/auth/nginx` — nginx `auth_request`

**Request from proxy:**
- Method always rewritten to `GET` (nginx limitation)
- Body is stripped (`proxy_pass_request_body off`)
- Original URL provided via `X-Original-URL` or `X-Original-URI`, set explicitly in nginx config
- Cookies forwarded

**Response to proxy:**
| Scenario | Status | Notes |
|---|---|---|
| Authenticated | `200` | Identity in response headers; nginx copies via `auth_request_set` + `proxy_set_header`. Include `Set-Cookie` here if refreshing the session |
| Unauthenticated | `401` | nginx config handles the redirect via `error_page 401 = @signin` |
| Forbidden | `403` | Blocks the request |

**Critical:** Do **not** return `302` from this endpoint. nginx cannot follow redirects from `auth_request` and will treat it as auth success.

---

### `/auth/caddy` — Caddy `forward_auth`

**Request from proxy:**
- Full original request forwarded (method, headers, body configurable)
- Cookies forwarded as-is

**Response to proxy:**
| Scenario | Status | Notes |
|---|---|---|
| Authenticated | `200` | Identity headers; Caddy picks up listed headers via `copy_headers` |
| Unauthenticated | `302` | Caddy follows redirects when configured to |
| Forbidden | `403` | Blocks |

Caddy is more flexible than nginx; either redirect or 401 mode can be configured.

---

### `/auth/envoy` — Envoy `ext_authz`

**Two modes:**

**HTTP mode:**
- `POST` request with original metadata in `X-Original-*` headers
- Body forwarding is opt-in
- Response distinguishes:
  - `headers_to_add` — sent upstream to the protected service
  - `headers_to_set` / `response_headers_to_add` — sent back downstream to the client

**gRPC mode (preferred in production Envoy):**
- Implements `envoy.service.auth.v3.Authorization` proto
- Completely different wire format from HTTP
- Returns `OkResponse` or `DeniedResponse` messages

**Response semantics:**
| Scenario | HTTP status | gRPC |
|---|---|---|
| Authenticated | `200` + `headers_to_add` | `OkResponse` |
| Unauthenticated | `401` or `403` | `DeniedResponse` with status code |

---

## Shared Requirements

### Session validation flow

All `/auth/*` endpoints share the same core logic:

1. Extract session cookie from request (configurable cookie name, e.g. `_midp_session`).
2. Look up session in session store.
3. If valid and unexpired, return identity claims as response headers.
4. If invalid or missing, return the proxy-specific unauthenticated response.

### Identity headers (sent upstream on success)

Standard set returned by all endpoints on success:

- `X-Auth-User` — subject identifier (username or stable ID)
- `X-Auth-Email`
- `X-Auth-Name` — display name
- `X-Auth-Groups` — comma-separated group list
- `X-Auth-Uid` — stable opaque user ID
- `X-Auth-Token` — optional signed JWT for downstream verification

These names should be configurable per deployment.

### Security: header stripping

**The IdP and/or proxy must strip all `X-Auth-*` headers from inbound client requests** before any auth logic runs. Otherwise a client can spoof identity by setting the headers themselves. This is a critical requirement, not optional.

### Login flow endpoints (unprotected)

- `GET /start?rd=<url>` — initiates login (OIDC/OAuth redirect, magic link, etc.). `rd` is the post-login redirect target, must be validated against an allowlist.
- `GET /callback` — OIDC/OAuth return endpoint. Exchanges code for tokens, creates session, sets cookie, redirects to `rd`.
- `GET /sign_out` — destroys session, clears cookie, optionally redirects to upstream IdP logout.
- `GET /healthz` — liveness check, no auth.

### Session store

- Backed by Redis, SQLite, or in-memory (configurable).
- Per-request validation: no proxy-side caching by default. Acceptable to add short TTL caching as opt-in, with explicit revocation trade-off documented.
- Sessions should be revocable server-side (logout-everywhere works).

### Cookies

- Signed and/or encrypted; never trust raw client-provided session data.
- Default flags: `HttpOnly`, `Secure`, `SameSite=Lax`.
- Domain scope configurable for SSO across subdomains.

### Trust boundaries

- `X-Forwarded-*` headers must only be trusted when the request comes from a configured proxy IP/CIDR.
- All other sources must have these headers treated as untrusted client input.

## Non-Goals

- Full OIDC provider implementation (token issuance, dynamic client registration). The IdP can either federate to an upstream OIDC provider or implement a minimal subset.
- LDAP outpost (out of scope for v1).
- Per-request caching at the IdP layer; that's the proxy's concern if needed.

## Reference Implementations to Study

- Authentik outpost: `internal/outpost/proxyv2/` in the goauthentik/authentik repo
- oauth2-proxy: `pkg/middleware/auth_*.go`
- Traefik forwardAuth source: `pkg/middlewares/auth/forward.go`