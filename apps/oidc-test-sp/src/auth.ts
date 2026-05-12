import NextAuth from "next-auth"

const issuer = process.env.AUTH_MIN_IDP_ISSUER || "http://localhost:8080"

export const { handlers, signIn, signOut, auth } = NextAuth({
  providers: [
    {
      id: "min-idp",
      name: "Min-IDP",
      type: "oidc",
      issuer,
      clientId: process.env.AUTH_MIN_IDP_ID || "test-client",
      clientSecret: process.env.AUTH_MIN_IDP_SECRET || "test-secret",
      authorization: {
        params: {
          scope: "openid profile email",
          response_type: "code",
        },
      },
      client: {
        token_endpoint_auth_method: "client_secret_basic",
      },
      checks: ["pkce", "state", "nonce"],
      profile(profile) {
        // profile comes from the userinfo endpoint or ID token claims.
        // email is in both (we embed it when minting the session JWT).
        return {
          id:    profile.sub,
          email: profile.email ?? null,
          name:  profile.name  ?? profile.email ?? profile.sub,
        }
      },
    },
  ],
  callbacks: {
    async jwt({ token, account }) {
      if (account) {
        token.idToken           = account.id_token
        token.accessToken       = account.access_token
        token.providerAccountId = account.providerAccountId

        // Extract email from the ID token claims directly — no userinfo round-trip needed.
        if (account.id_token) {
          try {
            const payload = JSON.parse(
              Buffer.from(account.id_token.split(".")[1], "base64url").toString()
            )
            token.email = payload.email ?? token.email
            token.name  = payload.email ?? token.name
          } catch { /* ignore decode errors */ }
        }
      }
      return token
    },
    async session({ session, token }) {
      return {
        ...session,
        idToken: token.idToken,
        accessToken: token.accessToken,
        providerAccountId: token.providerAccountId,
      }
    },
  },
})

// Cache the discovery document in memory for the process lifetime.
let _discovery: Record<string, string> | null = null

async function getDiscovery(): Promise<Record<string, string>> {
  if (_discovery) return _discovery
  const res = await fetch(`${issuer}/.well-known/openid-configuration`, {
    next: { revalidate: 3600 },
  })
  if (!res.ok) throw new Error(`OIDC discovery failed: ${res.status}`)
  _discovery = await res.json()
  return _discovery!
}

/**
 * Builds the RP-initiated logout URL using end_session_endpoint from
 * the provider's discovery document — works with any OIDC provider.
 */
export async function buildLogoutUrl(idToken: string, baseUrl: string): Promise<string> {
  const discovery = await getDiscovery()
  const endSessionEndpoint = discovery["end_session_endpoint"]
  if (!endSessionEndpoint) {
    // Provider doesn't advertise end_session_endpoint — fall back to local signout only.
    return baseUrl
  }
  const url = new URL(endSessionEndpoint)
  url.searchParams.set("id_token_hint", idToken)
  url.searchParams.set("post_logout_redirect_uri", baseUrl)
  return url.toString()
}
