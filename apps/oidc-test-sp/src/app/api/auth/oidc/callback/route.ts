import { NextRequest, NextResponse } from "next/server"
import { getConfig, oidcRedirectUri, authorizationCodeGrant } from "@/lib/oidc"
import { getSession } from "@/lib/session"

export async function GET(req: NextRequest) {
  const session = await getSession()
  const pending = session.pending

  if (!pending) {
    return NextResponse.redirect(new URL("/?error=oidc_failed", req.url))
  }

  try {
    const config = await getConfig()
    const tokens = await authorizationCodeGrant(config, new URL(req.url), {
      pkceCodeVerifier: pending.codeVerifier,
      expectedState: pending.state,
      expectedNonce: pending.nonce,
      redirectUri: oidcRedirectUri(),
    })

    const claims = tokens.claims()
    const sub = claims?.sub ?? ""
    const email = (claims?.email as string | undefined) ?? null
    const name = (claims?.name as string | undefined) ?? email
    const roles = (claims?.roles as string[] | undefined) ?? []

    session.pending = undefined
    session.user = {
      sub,
      email,
      name,
      roles,
      idToken: tokens.id_token ?? "",
      accessToken: tokens.access_token ?? "",
    }
    await session.save()

    return NextResponse.redirect(new URL(pending.redirectTo, req.url))
  } catch (err) {
    console.error("OIDC callback error", err)
    return NextResponse.redirect(new URL("/?error=oidc_failed", req.url))
  }
}
