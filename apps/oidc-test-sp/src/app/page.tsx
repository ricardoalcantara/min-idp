import { auth, signIn, signOut, buildLogoutUrl } from "@/auth";
import { headers } from "next/headers";

export default async function Home() {
  const session = await auth();
  const hdrs = await headers();
  const baseUrl = `${hdrs.get("x-forwarded-proto") ?? "http"}://${hdrs.get("host") ?? "localhost:3000"}`;

  const card = {
    background: "var(--card)",
    border: "1px solid var(--card-border)",
  } as React.CSSProperties;

  const cardHeader = {
    background: "var(--card-header)",
    borderBottom: "1px solid var(--card-border)",
  } as React.CSSProperties;

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="max-w-3xl w-full">

        {/* Header */}
        <div className="text-center mb-10">
          <div className="inline-flex items-center justify-center w-12 h-12 bg-indigo-600 rounded-xl mb-5 shadow-sm">
            <span className="text-white font-bold text-lg">M</span>
          </div>
          <h1 className="text-3xl font-bold mb-2">Min-IDP Tester</h1>
          <p className="text-sm" style={{ color: "var(--muted)" }}>OIDC · PKCE · State · Nonce</p>
        </div>

        {!session ? (
          <div className="rounded-2xl shadow-sm p-8 text-center" style={card}>
            <h2 className="text-lg font-semibold mb-2">Sign in to continue</h2>
            <p className="text-sm mb-6" style={{ color: "var(--muted)" }}>Authenticate via Min-IDP to inspect your tokens.</p>
            <form
              action={async () => {
                "use server"
                await signIn("min-idp")
              }}
            >
              <button
                type="submit"
                className="inline-flex items-center gap-2 px-6 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-semibold rounded-lg transition-colors"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                </svg>
                Sign in with Min-IDP
              </button>
            </form>
          </div>
        ) : (
          <div className="space-y-4">
            {/* Status bar */}
            <div className="rounded-2xl shadow-sm p-5 flex items-center justify-between" style={card}>
              <div className="flex items-center gap-3">
                <span className="inline-flex w-2.5 h-2.5 rounded-full bg-emerald-500 ring-4" style={{ boxShadow: "0 0 0 4px var(--status-ring)" }} />
                <div>
                  <p className="text-sm font-semibold">Authenticated</p>
                  <p className="text-xs" style={{ color: "var(--muted)" }}>{session.user?.email}</p>
                </div>
              </div>
              <form
                action={async () => {
                  "use server"
                  // @ts-expect-error Session property extension
                  const idToken = session?.idToken as string | undefined
                  if (idToken) {
                    await signOut({ redirect: false })
                    const logoutUrl = buildLogoutUrl(idToken, baseUrl)
                    const { redirect } = await import("next/navigation")
                    redirect(logoutUrl)
                  } else {
                    await signOut()
                  }
                }}
              >
                <button
                  type="submit"
                  className="px-4 py-2 text-sm font-medium rounded-lg transition-colors"
                  style={{ background: "var(--card-border)", color: "var(--foreground)" }}
                >
                  Sign out
                </button>
              </form>
            </div>

            {/* ID Token */}
            <div className="rounded-2xl shadow-sm overflow-hidden" style={card}>
              <div className="flex items-center justify-between px-5 py-3" style={cardHeader}>
                <span className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted)" }}>ID Token</span>
                <span className="text-xs px-2 py-0.5 rounded-full bg-indigo-100 text-indigo-600 font-medium">JWT</span>
              </div>
              <pre className="p-5 text-xs font-mono overflow-x-auto whitespace-pre-wrap break-all"
                style={{ background: "var(--code-indigo-bg)", color: "var(--code-indigo-text)" }}>
                {/* @ts-expect-error Session property extension */}
                {session?.idToken || "No ID Token returned"}
              </pre>
            </div>

            {/* Profile + Access Token */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="rounded-2xl shadow-sm overflow-hidden" style={card}>
                <div className="px-5 py-3" style={cardHeader}>
                  <span className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted)" }}>Session Profile</span>
                </div>
                <pre className="p-5 text-xs font-mono overflow-x-auto"
                  style={{ background: "var(--code-emerald-bg)", color: "var(--code-emerald-text)" }}>
                  {JSON.stringify(session?.user, null, 2)}
                </pre>
              </div>

              <div className="rounded-2xl shadow-sm overflow-hidden" style={card}>
                <div className="px-5 py-3" style={cardHeader}>
                  <span className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted)" }}>Access Token</span>
                </div>
                <pre className="p-5 text-xs font-mono overflow-x-auto whitespace-pre-wrap break-all"
                  style={{ background: "var(--code-amber-bg)", color: "var(--code-amber-text)" }}>
                  {/* @ts-expect-error Session property extension */}
                  {session?.accessToken || "No Access Token returned"}
                </pre>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
