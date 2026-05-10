Here's the full implementation order, simple to complex:

Tier 1 — Pure CRUD, no dependencies between modules

/api/admin/groups — list, create, get, update, delete (entity + repo already exist, just stubs now)
/api/admin/roles — complete CRUD + list permissions, update role (stubs exist)
/api/admin/users — CRUD, assign/remove role, assign/remove group, list roles, list groups, force sessions, admin password reset
/api/me — update profile, revoke single session (stubs exist)
Tier 2 — Service Providers registry (CRUD with sub-resources)

/api/admin/sps — SP CRUD
/api/admin/sps/:id/oidc — GET + PUT OIDC client config
/api/admin/sps/:id/saml — GET + PUT SAML client config
/api/admin/sps/:id/access-rules — CRUD access rules (allow/deny by role/group/user)
Tier 3 — Keys + Audit (no auth-flow dependencies)

/api/admin/keys/:protocol — list active+previous keys
/api/admin/keys/:protocol/rotate — generate new active, demote old to previous
/api/admin/audit — list with filtering + pagination
Tier 4 — Session-based SSO gating (requires SPs + RBAC ready)

Access rule evaluation — the sp:login + per-SP rule check engine used by both OIDC and SAML flows
Tier 5 — OIDC / OAuth2 flow

GET /oauth2/authorize — validate client, check session, stash in KV, redirect to /login, issue code on resume
POST /oauth2/token — auth code exchange, PKCE validation, issue access + refresh tokens
GET /oauth2/userinfo — introspect token, return claims
POST /oauth2/revoke — revoke access/refresh token
POST /oauth2/introspect — token introspection endpoint
GET/POST /oauth2/logout — RP-initiated logout, clear IdP session, SLO back-channel
Tier 6 — SAML 2.0 flow

GET/POST /saml/sso — parse AuthnRequest, check session, issue assertion response
GET/POST /saml/slo — single logout request handling, notify SP sessions
