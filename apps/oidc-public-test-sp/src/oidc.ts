const STORAGE_KEY = "oidc_public_test_sp"
import { sha256 } from "./sha256"

export type OIDCDiscovery = {
  issuer: string
  authorization_endpoint: string
  token_endpoint: string
  userinfo_endpoint: string
}

export type TokenSet = {
  access_token: string
  id_token?: string
  refresh_token?: string
  token_type: string
  expires_in: number
}

export type UserInfo = {
  sub: string
  email?: string
  name?: string
  username?: string
  roles?: string[]
}

type PendingAuth = {
  codeVerifier: string
  state: string
  nonce: string
}

export type SessionData = {
  tokens?: TokenSet
  userinfo?: UserInfo
}

const DEFAULT_ISSUER = "http://192.168.1.107:8081"
const DEFAULT_CLIENT_ID = "oidc-public-test-sp"
const DEFAULT_REDIRECT_URI = "http://192.168.1.107:5173/callback"

export function issuer(): string {
  return import.meta.env.VITE_OIDC_ISSUER ?? DEFAULT_ISSUER
}

export function clientId(): string {
  return import.meta.env.VITE_OIDC_CLIENT_ID ?? DEFAULT_CLIENT_ID
}

export function redirectUri(): string {
  return import.meta.env.VITE_OIDC_REDIRECT_URI ?? DEFAULT_REDIRECT_URI
}

function randomString(bytes = 32): string {
  const buf = new Uint8Array(bytes)
  crypto.getRandomValues(buf)
  return base64UrlEncode(buf)
}

function base64UrlEncode(buf: Uint8Array): string {
  let str = ""
  for (const b of buf) {
    str += String.fromCharCode(b)
  }
  return btoa(str).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "")
}

export async function pkceChallenge(verifier: string): Promise<string> {
  const data = new TextEncoder().encode(verifier)
  const digest = await sha256(data)
  return base64UrlEncode(digest)
}

let discoveryCache: OIDCDiscovery | null = null

export async function discover(): Promise<OIDCDiscovery> {
  if (discoveryCache) return discoveryCache
  const url = `${issuer().replace(/\/$/, "")}/.well-known/openid-configuration`
  const res = await fetch(url)
  if (!res.ok) {
    throw new Error(`Discovery failed: ${res.status}`)
  }
  discoveryCache = await res.json()
  return discoveryCache!
}

function pendingKey(): string {
  return `${STORAGE_KEY}:pending`
}

function sessionKey(): string {
  return `${STORAGE_KEY}:session`
}

export function loadSession(): SessionData {
  const raw = sessionStorage.getItem(sessionKey())
  if (!raw) return {}
  try {
    return JSON.parse(raw) as SessionData
  } catch {
    return {}
  }
}

export function saveSession(data: SessionData): void {
  sessionStorage.setItem(sessionKey(), JSON.stringify(data))
}

export function clearSession(): void {
  sessionStorage.removeItem(sessionKey())
  sessionStorage.removeItem(pendingKey())
}

export async function startLogin(): Promise<void> {
  const doc = await discover()
  const codeVerifier = randomString(32)
  const codeChallenge = await pkceChallenge(codeVerifier)
  const state = randomString(16)
  const nonce = randomString(16)

  const pending: PendingAuth = { codeVerifier, state, nonce }
  sessionStorage.setItem(pendingKey(), JSON.stringify(pending))

  const params = new URLSearchParams({
    client_id: clientId(),
    redirect_uri: redirectUri(),
    response_type: "code",
    scope: "openid profile email",
    code_challenge: codeChallenge,
    code_challenge_method: "S256",
    state,
    nonce,
  })

  window.location.href = `${doc.authorization_endpoint}?${params}`
}

export async function handleCallback(search: string): Promise<SessionData> {
  const params = new URLSearchParams(search)
  const error = params.get("error")
  if (error) {
    throw new Error(params.get("error_description") ?? error)
  }

  const code = params.get("code")
  const state = params.get("state")
  if (!code || !state) {
    throw new Error("Missing code or state")
  }

  const pendingRaw = sessionStorage.getItem(pendingKey())
  if (!pendingRaw) {
    throw new Error("Missing PKCE session — start login again")
  }
  const pending = JSON.parse(pendingRaw) as PendingAuth
  if (pending.state !== state) {
    throw new Error("State mismatch")
  }

  const doc = await discover()
  const body = new URLSearchParams({
    grant_type: "authorization_code",
    client_id: clientId(),
    code,
    redirect_uri: redirectUri(),
    code_verifier: pending.codeVerifier,
  })

  const res = await fetch(doc.token_endpoint, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body,
  })

  if (!res.ok) {
    const text = await res.text()
    throw new Error(`Token exchange failed: ${res.status} ${text}`)
  }

  const tokens = (await res.json()) as TokenSet
  sessionStorage.removeItem(pendingKey())

  let userinfo: UserInfo | undefined
  if (tokens.access_token) {
    const uiRes = await fetch(doc.userinfo_endpoint, {
      headers: { Authorization: `Bearer ${tokens.access_token}` },
    })
    if (uiRes.ok) {
      userinfo = await uiRes.json()
    }
  }

  const session: SessionData = { tokens, userinfo }
  saveSession(session)
  return session
}

export function logout(): void {
  clearSession()
  window.location.href = "/"
}
