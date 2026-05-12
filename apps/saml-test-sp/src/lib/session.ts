import { getIronSession, SessionOptions } from "iron-session"
import { cookies } from "next/headers"

export interface SAMLSessionData {
  nameId?:     string
  nameIdFormat?: string
  attributes?: Record<string, string | string[]>
}

const opts: SessionOptions = {
  cookieName: "saml_session",
  password:   process.env.SESSION_SECRET ?? "dev-secret-change-in-production!!",
  cookieOptions: { secure: process.env.NODE_ENV === "production" },
}

export async function getSession() {
  return getIronSession<SAMLSessionData>(await cookies(), opts)
}
