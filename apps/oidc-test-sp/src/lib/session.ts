import { getIronSession, SessionOptions } from "iron-session"
import { cookies } from "next/headers"

export interface OIDCSessionData {
  pending?: {
    codeVerifier: string
    state: string
    nonce: string
    redirectTo: string
  }
  user?: {
    sub: string
    email: string | null
    name: string | null
    roles: string[]
    idToken: string
    accessToken: string
  }
}

const opts: SessionOptions = {
  cookieName: "oidc_session",
  password:   process.env.SESSION_SECRET ?? "dev-secret-change-in-production!!",
  cookieOptions: { secure: process.env.NODE_ENV === "production" },
}

export async function getSession() {
  return getIronSession<OIDCSessionData>(await cookies(), opts)
}
