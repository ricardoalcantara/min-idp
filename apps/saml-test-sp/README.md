# SAML Test SP

A minimal Next.js SAML 2.0 Service Provider for testing against Authentik (or min-idp).

## Setup

### 1. Generate SP key pair

```bash
openssl req -x509 -newkey rsa:2048 -keyout sp-key.pem -out sp-cert.pem \
  -days 365 -nodes -subj "/CN=saml-test-sp"
```

### 2. Configure environment

```bash
cp .env.local.example .env.local
```

Fill in `.env.local`:
- Paste the contents of `sp-key.pem` into `SAML_SP_PRIVATE_KEY` (replace newlines with `\n`)
- Paste the contents of `sp-cert.pem` into `SAML_SP_CERT`
- Fill in IdP values from the next step

### 3. Configure Authentik

1. Create a new **SAML Provider** in Authentik:
   - **ACS URL**: `http://localhost:3002/api/auth/saml/callback`
   - **Issuer**: `http://localhost:3002`
   - **Signing Certificate**: upload `sp-cert.pem`
   - **Service Provider Binding**: `Post`
   - **Assertion signing**: enabled

2. Copy from the Authentik provider:
   - **SSO URL** → `SAML_IDP_SSO_URL`
   - **Entity ID** → `SAML_IDP_ENTITY_ID`
   - **Certificate** (base64, no headers) → `SAML_IDP_CERT`

   Or point `SAML_IDP_SSO_URL` at the redirect binding URL from Authentik's metadata.

### 4. Run

```bash
npm install
npm run dev   # http://localhost:3002
```

Visit `/api/auth/saml/metadata` to get SP metadata XML to import into Authentik.

## Routes

| Route | Method | Description |
|-------|--------|-------------|
| `/` | GET | Home — shows assertion after login |
| `/api/auth/saml/login` | GET | Initiates SAML auth (redirects to IdP) |
| `/api/auth/saml/callback` | POST | ACS — receives SAMLResponse from IdP |
| `/api/auth/saml/metadata` | GET | SP metadata XML |
| `/api/auth/saml/logout` | POST | Clears SP session (local only) |
