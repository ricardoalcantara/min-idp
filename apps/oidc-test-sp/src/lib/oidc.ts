import * as client from "openid-client"

function requireEnv(key: string): string {
  const v = process.env[key]
  if (!v) throw new Error(`Missing required env var: ${key}`)
  return v
}

export const oidcRedirectUri = () =>
  process.env.OIDC_REDIRECT_URI ?? "http://localhost:3001/api/auth/oidc/callback"

let _config: client.Configuration | null = null

export async function getConfig(): Promise<client.Configuration> {
  if (_config) return _config
  const issuer = new URL(requireEnv("OIDC_ISSUER"))
  // openid-client v6 blocks HTTP by default; allow it for local development
  const opts = issuer.protocol === "http:" ? { execute: [client.allowInsecureRequests] } : undefined
  _config = await client.discovery(
    issuer,
    requireEnv("OIDC_CLIENT_ID"),
    { client_secret: requireEnv("OIDC_CLIENT_SECRET") },
    undefined,
    opts,
  )
  return _config
}

export {
  randomPKCECodeVerifier,
  calculatePKCECodeChallenge,
  randomState,
  randomNonce,
  buildAuthorizationUrl,
  authorizationCodeGrant,
  buildEndSessionUrl,
} from "openid-client"
