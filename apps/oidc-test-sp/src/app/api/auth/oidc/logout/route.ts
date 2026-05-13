import { NextRequest, NextResponse } from "next/server"
import { getConfig, buildEndSessionUrl } from "@/lib/oidc"
import { getSession } from "@/lib/session"

export async function POST(req: NextRequest) {
  const session = await getSession()
  const idToken = session.user?.idToken
  await session.destroy()

  if (idToken) {
    try {
      const config = await getConfig()
      const base = new URL("/", req.url).href
      const logoutUrl = buildEndSessionUrl(config, {
        id_token_hint: idToken,
        post_logout_redirect_uri: base,
      })
      return NextResponse.redirect(logoutUrl.href)
    } catch {
      // Provider doesn't support end_session_endpoint — fall through to local signout
    }
  }

  return NextResponse.redirect(new URL("/", req.url))
}
