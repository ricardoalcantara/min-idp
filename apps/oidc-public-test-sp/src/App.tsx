import { useCallback, useState } from "react"
import { Callback } from "./Callback"
import { loadSession, logout, startLogin, issuer, clientId, redirectUri } from "./oidc"
import type { SessionData } from "./oidc"
import "./App.css"

export default function App() {
  const [session, setSession] = useState<SessionData>(() => loadSession())
  const [error, setError] = useState<string | null>(null)

  const onComplete = useCallback((data: SessionData) => {
    setSession(data)
    setError(null)
  }, [])

  const onError = useCallback((message: string) => {
    setError(message)
  }, [])

  if (window.location.pathname === "/callback") {
    return (
      <main className="card">
        <h1>OIDC Public Test SP</h1>
        <Callback onComplete={onComplete} onError={onError} />
        {error && <p className="error">{error}</p>}
      </main>
    )
  }

  return (
    <main className="card">
      <h1>OIDC Public Test SP</h1>
      <p className="sub">Browser-only · PKCE · no client secret</p>

      <dl className="config">
        <dt>Issuer</dt>
        <dd>{issuer()}</dd>
        <dt>Client ID</dt>
        <dd>{clientId()}</dd>
        <dt>Redirect URI</dt>
        <dd>{redirectUri()}</dd>
      </dl>

      {error && <p className="error">{error}</p>}

      {!session.tokens ? (
        <button type="button" onClick={() => startLogin().catch((e) => setError(String(e)))}>
          Sign in with min-idp
        </button>
      ) : (
        <>
          <section>
            <h2>Signed in</h2>
            {session.userinfo ? (
              <pre>{JSON.stringify(session.userinfo, null, 2)}</pre>
            ) : (
              <pre>{JSON.stringify(session.tokens, null, 2)}</pre>
            )}
          </section>
          <button type="button" className="secondary" onClick={logout}>
            Sign out
          </button>
        </>
      )}
    </main>
  )
}
