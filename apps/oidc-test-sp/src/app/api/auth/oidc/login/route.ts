import { NextRequest, NextResponse } from "next/server"
import {
  getConfig,
  oidcRedirectUri,
  randomPKCECodeVerifier,
  calculatePKCECodeChallenge,
  randomState,
  randomNonce,
  buildAuthorizationUrl,
} from "@/lib/oidc"
import { getSession } from "@/lib/session"

export async function GET(req: NextRequest) {
  const config = await getConfig()
  const redirectTo = req.nextUrl.searchParams.get("next") ?? "/"

  const codeVerifier = randomPKCECodeVerifier()
  const codeChallenge = await calculatePKCECodeChallenge(codeVerifier)
  const state = randomState()
  const nonce = randomNonce()

  const session = await getSession()
  session.pending = { codeVerifier, state, nonce, redirectTo }
  await session.save()

  const authUrl = buildAuthorizationUrl(config, {
    redirect_uri: oidcRedirectUri(),
    scope: "openid profile email",
    code_challenge: codeChallenge,
    code_challenge_method: "S256",
    state,
    nonce,
  })

  return NextResponse.redirect(authUrl.href)
}
