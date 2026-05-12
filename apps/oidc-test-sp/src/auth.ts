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
    },
  ],
  callbacks: {
    async jwt({ token, account }) {
      if (account) {
        token.idToken = account.id_token
        token.accessToken = account.access_token
        token.providerAccountId = account.providerAccountId
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

/**
 * Returns the IdP end_session URL for RP-initiated logout.
 * Passes id_token_hint so the IdP knows which session to terminate.
 */
export function buildLogoutUrl(idToken: string, baseUrl: string): string {
  const endSession = new URL(`${issuer}/oauth2/logout`)
  endSession.searchParams.set("id_token_hint", idToken)
  endSession.searchParams.set("post_logout_redirect_uri", baseUrl)
  return endSession.toString()
}
