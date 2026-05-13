import { getSession } from "@/lib/session"

const S = {
  page:      { minHeight: "100vh", display: "flex", alignItems: "center", justifyContent: "center", padding: "1rem" } as React.CSSProperties,
  wrap:      { width: "100%", maxWidth: "680px" } as React.CSSProperties,
  header:    { textAlign: "center" as const, marginBottom: "2.5rem" },
  logo:      { width: 44, height: 44, background: "linear-gradient(135deg,#4f46e5,#7c3aed)", borderRadius: 10, display: "inline-flex", alignItems: "center", justifyContent: "center", color: "#fff", fontWeight: 700, fontSize: 18, marginBottom: "1.25rem" } as React.CSSProperties,
  h1:        { fontSize: "1.75rem", fontWeight: 700, marginBottom: "0.375rem" } as React.CSSProperties,
  sub:       { fontSize: "0.875rem", color: "var(--muted)" } as React.CSSProperties,
  card:      { background: "var(--card)", border: "1px solid var(--card-border)", borderRadius: 14, boxShadow: "0 2px 12px rgba(0,0,0,.06)", overflow: "hidden", marginBottom: "1rem" } as React.CSSProperties,
  ch:        { background: "var(--card-header)", borderBottom: "1px solid var(--card-border)", padding: "0.75rem 1.25rem", display: "flex", alignItems: "center", justifyContent: "space-between" } as React.CSSProperties,
  label:     { fontSize: "0.7rem", fontWeight: 600, textTransform: "uppercase" as const, letterSpacing: "0.05em", color: "var(--muted)" },
  badge:     (color: string) => ({ fontSize: "0.7rem", padding: "0.2rem 0.6rem", borderRadius: 99, background: `rgba(${color},.1)`, color: `rgb(${color})`, fontWeight: 500 } as React.CSSProperties),
  pre:       (bg: string, fg: string) => ({ padding: "1.25rem", fontSize: "0.75rem", fontFamily: "monospace", overflow: "auto", background: `var(${bg})`, color: `var(${fg})`, whiteSpace: "pre-wrap" as const, wordBreak: "break-all" as const }),
  status:    { background: "var(--card)", border: "1px solid var(--card-border)", borderRadius: 14, padding: "1rem 1.25rem", display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: "1rem" } as React.CSSProperties,
  dot:       { width: 10, height: 10, borderRadius: "50%", background: "#10b981", boxShadow: "0 0 0 4px var(--status-ring)", marginRight: "0.75rem", flexShrink: 0 } as React.CSSProperties,
  loginBtn:  { display: "inline-flex", alignItems: "center", gap: "0.5rem", padding: "0.625rem 1.5rem", background: "#4f46e5", color: "#fff", border: "none", borderRadius: 8, fontSize: "0.9rem", fontWeight: 600, cursor: "pointer" } as React.CSSProperties,
  logoutBtn: { padding: "0.5rem 1rem", background: "var(--card-border)", color: "var(--text)", border: "none", borderRadius: 8, fontSize: "0.875rem", fontWeight: 500, cursor: "pointer" } as React.CSSProperties,
  toggle:    { position: "fixed" as const, top: 16, right: 16, width: 36, height: 36, borderRadius: "50%", background: "var(--card-border)", border: "none", cursor: "pointer", fontSize: 18 },
  grid:      { display: "grid", gridTemplateColumns: "repeat(auto-fit,minmax(280px,1fr))", gap: "1rem" } as React.CSSProperties,
}

export default async function Home() {
  const session = await getSession()
  const user = session.user

  return (
    <div style={S.page}>
      <div style={S.wrap}>
        <button style={S.toggle} id="dark-toggle" suppressHydrationWarning>🌙</button>
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
          <div style={S.logo}>O</div>
          <h1 style={S.h1}>OIDC Test SP</h1>
          <p style={S.sub}>OIDC · PKCE · State · Nonce</p>
        </div>

        {!user ? (
          <div style={{ ...S.card, textAlign: "center", padding: "2.5rem 2rem" }}>
            <p style={{ marginBottom: "0.5rem", fontWeight: 600 }}>Sign in to continue</p>
            <p style={{ ...S.sub, marginBottom: "1.5rem" }}>Authenticate via OIDC to inspect your tokens.</p>
            <a href="/api/auth/oidc/login" style={S.loginBtn}>
              <svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
              </svg>
              Sign in with OIDC
            </a>
          </div>
        ) : (
          <>
            <div style={S.status}>
              <div style={{ display: "flex", alignItems: "center" }}>
                <span style={S.dot} />
                <div>
                  <div style={{ fontWeight: 600, fontSize: "0.9rem" }}>Authenticated</div>
                  <div style={{ ...S.sub, fontSize: "0.8rem" }}>{user.email}</div>
                </div>
              </div>
              <form method="POST" action="/api/auth/oidc/logout">
                <button type="submit" style={S.logoutBtn}>Sign out</button>
              </form>
            </div>

            <div style={S.card}>
              <div style={S.ch}>
                <span style={S.label}>ID Token</span>
                <span style={S.badge("99,102,241")}>JWT</span>
              </div>
              <pre style={S.pre("--code-bg", "--code-text")}>
                {user.idToken || "No ID Token returned"}
              </pre>
            </div>

            <div style={S.grid}>
              <div style={S.card}>
                <div style={S.ch}>
                  <span style={S.label}>Session Profile</span>
                </div>
                <pre style={S.pre("--attr-bg", "--attr-text")}>
                  {JSON.stringify({ sub: user.sub, email: user.email, name: user.name, roles: user.roles }, null, 2)}
                </pre>
              </div>

              <div style={S.card}>
                <div style={S.ch}>
                  <span style={S.label}>Access Token</span>
                </div>
                <pre style={S.pre("--amber-bg", "--amber-text")}>
                  {user.accessToken || "No Access Token returned"}
                </pre>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
