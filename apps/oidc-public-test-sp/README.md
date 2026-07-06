# OIDC Public Test SP

Browser-only OIDC client for [min-idp](../../README.md). Uses the authorization code flow with **PKCE** and **no client secret** (`token_endpoint_auth: none`).

## Prerequisites

1. min-idp running (default `http://localhost:8081`)
2. CORS enabled for this origin on the IdP:

   ```env
   MINSTACK_CORS_ORIGIN=http://localhost:5173
   ```

3. A public OIDC client registered in min-idp:

   ```bash
   # Create SP
   curl -X POST http://localhost:8081/api/admin/sps \
     -H "Authorization: Bearer <admin-token>" \
     -H "Content-Type: application/json" \
     -d '{"slug":"oidc-public-test-sp","name":"OIDC Public Test SP","protocol":"oidc"}'

   # Configure public client (use SP UUID from response)
   curl -X PUT http://localhost:8081/api/admin/sps/<sp-uuid>/oidc \
     -H "Authorization: Bearer <admin-token>" \
     -H "Content-Type: application/json" \
     -d '{
       "client_id": "oidc-public-test-sp",
       "redirect_uris": ["http://localhost:5173/callback"],
       "post_logout_redirect_uris": ["http://localhost:5173/"],
       "token_endpoint_auth": "none"
     }'
   ```

## Run

```bash
cp .env.example .env
npm install
npm run dev
```

Open [http://localhost:5173](http://localhost:5173) and click **Sign in with min-idp**.

## How it works

All OIDC logic runs in the browser:

1. Generate PKCE `code_verifier` / `code_challenge` (S256)
2. Redirect to `/oauth2/authorize`
3. On `/callback`, POST to `/oauth2/token` with `client_id` and `code_verifier` only
4. Fetch `/oauth2/userinfo` with the access token

Tokens are kept in `sessionStorage` for the demo.
