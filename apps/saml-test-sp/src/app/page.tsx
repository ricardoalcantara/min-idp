import { getSession } from "@/lib/session"
import { spEntityId } from "@/lib/saml"

const S = {
  page:    { minHeight: "100vh", display: "flex", alignItems: "center", justifyContent: "center", padding: "1rem" } as React.CSSProperties,
  wrap:    { width: "100%", maxWidth: "680px" } as React.CSSProperties,
  header:  { textAlign: "center" as const, marginBottom: "2.5rem" },
  logo:    { width: 44, height: 44, background: "linear-gradient(135deg,#dc2626,#9333ea)", borderRadius: 10, display: "inline-flex", alignItems: "center", justifyContent: "center", color: "#fff", fontWeight: 700, fontSize: 18, marginBottom: "1.25rem" } as React.CSSProperties,
  h1:      { fontSize: "1.75rem", fontWeight: 700, marginBottom: "0.375rem" } as React.CSSProperties,
  sub:     { fontSize: "0.875rem", color: "var(--muted)" } as React.CSSProperties,
  card:    { background: "var(--card)", border: "1px solid var(--card-border)", borderRadius: 14, boxShadow: "0 2px 12px rgba(0,0,0,.06)", overflow: "hidden", marginBottom: "1rem" } as React.CSSProperties,
  ch:      { background: "var(--card-header)", borderBottom: "1px solid var(--card-border)", padding: "0.75rem 1.25rem", display: "flex", alignItems: "center", justifyContent: "space-between" } as React.CSSProperties,
  label:   { fontSize: "0.7rem", fontWeight: 600, textTransform: "uppercase" as const, letterSpacing: "0.05em", color: "var(--muted)" },
  badge:   { fontSize: "0.7rem", padding: "0.2rem 0.6rem", borderRadius: 99, background: "rgba(220,38,38,.1)", color: "#dc2626", fontWeight: 500 } as React.CSSProperties,
  pre:     { padding: "1.25rem", fontSize: "0.75rem", fontFamily: "monospace", overflow: "auto", background: "var(--code-bg)", color: "var(--code-text)", whiteSpace: "pre-wrap" as const, wordBreak: "break-all" as const },
  preAttr: { padding: "1.25rem", fontSize: "0.75rem", fontFamily: "monospace", overflow: "auto", background: "var(--attr-bg)", color: "var(--attr-text)" },
  status:  { background: "var(--card)", border: "1px solid var(--card-border)", borderRadius: 14, padding: "1rem 1.25rem", display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: "1rem" } as React.CSSProperties,
  dot:     { width: 10, height: 10, borderRadius: "50%", background: "#10b981", boxShadow: "0 0 0 4px rgba(16,185,129,.15)", marginRight: "0.75rem", flexShrink: 0 } as React.CSSProperties,
  loginBtn:{ display: "inline-flex", alignItems: "center", gap: "0.5rem", padding: "0.625rem 1.5rem", background: "#dc2626", color: "#fff", border: "none", borderRadius: 8, fontSize: "0.9rem", fontWeight: 600, cursor: "pointer", textDecoration: "none" } as React.CSSProperties,
  logoutBtn:{ padding: "0.5rem 1rem", background: "var(--card-border)", color: "var(--text)", border: "none", borderRadius: 8, fontSize: "0.875rem", fontWeight: 500, cursor: "pointer" } as React.CSSProperties,
  toggle:  { position: "fixed" as const, top: 16, right: 16, width: 36, height: 36, borderRadius: "50%", background: "var(--card-border)", border: "none", cursor: "pointer", fontSize: 18 },
  meta:    { fontSize: "0.75rem", color: "var(--muted)", marginTop: "0.5rem" } as React.CSSProperties,
}

export default async function Home({ searchParams }: { searchParams: Promise<{ error?: string }> }) {
  const session = await getSession()
  const { error } = await searchParams
  const entityId = spEntityId()

  return (
    <div style={S.page}>
      <div style={S.wrap}>
        <button style={S.toggle} id="dark-toggle"
          onClick={undefined}
          suppressHydrationWarning
        >🌙</button>
        <script dangerouslySetInnerHTML={{ __html: `
          (function(){
            var btn=document.getElementById('dark-toggle');
            var s=localStorage.getItem('theme');
            var d=s?s==='dark':matchMedia('(prefers-color-scheme:dark)').matches;
            if(d){document.documentElement.classList.add('dark');btn.textContent='☀️';}
            btn.addEventListener('click',function(){
              var now=document.documentElement.classList.toggle('dark');
              localStorage.setItem('theme',now?'dark':'light');
              btn.textContent=now?'☀️':'🌙';
            });
          })();
        ` }} />

        <div style={S.header}>
          <div style={S.logo}>S</div>
          <h1 style={S.h1}>SAML Test SP</h1>
          <p style={S.sub}>SAML 2.0 · HTTP-Redirect · POST ACS</p>
        </div>

        {error && (
          <div style={{ ...S.card, borderColor: "#fecaca" }}>
            <div style={{ ...S.ch, background: "#fef2f2" }}>
              <span style={{ ...S.label, color: "#dc2626" }}>Authentication Error</span>
            </div>
            <div style={{ padding: "1rem 1.25rem", fontSize: "0.875rem", color: "#dc2626" }}>
              SAML validation failed. Check IdP configuration and SP certificate.
            </div>
          </div>
        )}

        {!session.nameId ? (
          <div style={{ ...S.card, textAlign: "center", padding: "2.5rem 2rem" }}>
            <p style={{ marginBottom: "0.5rem", fontWeight: 600 }}>Sign in to continue</p>
            <p style={{ ...S.sub, marginBottom: "1.5rem" }}>Authenticate via SAML to inspect your assertion.</p>
            <a href="/api/auth/saml/login" style={S.loginBtn}>
              <svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
              </svg>
              Sign in with SAML
            </a>
            <p style={{ ...S.meta, marginTop: "1.5rem" }}>
              SP metadata: <a href="/api/auth/saml/metadata" style={{ color: "var(--code-text)" }}>/api/auth/saml/metadata</a>
              <br />Entity ID: <code style={{ fontSize: "0.7rem" }}>{entityId}</code>
            </p>
          </div>
        ) : (
          <>
            <div style={S.status}>
              <div style={{ display: "flex", alignItems: "center" }}>
                <span style={S.dot} />
                <div>
                  <div style={{ fontWeight: 600, fontSize: "0.9rem" }}>Authenticated</div>
                  <div style={{ ...S.sub, fontSize: "0.8rem" }}>{session.nameId}</div>
                </div>
              </div>
              <form action="/api/auth/saml/logout" method="POST">
                <button style={S.logoutBtn} type="submit">Sign out</button>
              </form>
            </div>

            <div style={S.card}>
              <div style={S.ch}>
                <span style={S.label}>NameID</span>
                <span style={S.badge}>{session.nameIdFormat?.split(":").pop()}</span>
              </div>
              <pre style={S.pre}>{session.nameId}</pre>
            </div>

            <div style={S.card}>
              <div style={S.ch}>
                <span style={S.label}>Assertion Attributes</span>
                <span style={{ ...S.badge, background: "rgba(6,78,59,.1)", color: "#065f46" }}>
                  {Object.keys(session.attributes ?? {}).length} attrs
                </span>
              </div>
              <pre style={S.preAttr}>{JSON.stringify(session.attributes, null, 2)}</pre>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
