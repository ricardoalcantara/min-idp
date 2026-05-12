import { auth, signIn, signOut, buildLogoutUrl } from "@/auth";
import { headers } from "next/headers";

export default async function Home() {
  const session = await auth();
  const hdrs = await headers();
  const baseUrl = `${hdrs.get("x-forwarded-proto") ?? "http"}://${hdrs.get("host") ?? "localhost:3000"}`;

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex items-center justify-center p-4 selection:bg-indigo-500/30">
      <div className="max-w-4xl w-full">
        {/* Header */}
        <div className="text-center mb-12">
          <div className="inline-flex items-center justify-center p-2 bg-indigo-500/10 rounded-2xl mb-6 ring-1 ring-indigo-500/20 shadow-[0_0_40px_-10px_rgba(99,102,241,0.3)]">
            <div className="h-12 w-12 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-xl flex items-center justify-center text-white font-bold text-xl shadow-lg">
              IDP
            </div>
          </div>
          <h1 className="text-5xl font-extrabold tracking-tight bg-gradient-to-br from-white to-slate-400 bg-clip-text text-transparent mb-4">
            Min-IDP Tester
          </h1>
          <p className="text-slate-400 text-lg max-w-xl mx-auto">
            A comprehensive Service Provider testing utility for Min-IDP using OIDC, PKCE, and state checking.
          </p>
        </div>

        {/* Content */}
        {!session ? (
          <div className="bg-slate-900/50 backdrop-blur-xl rounded-3xl p-8 border border-slate-800 shadow-2xl text-center transform transition-all hover:scale-[1.01] duration-300">
            <h2 className="text-2xl font-semibold mb-6 text-slate-200">Authenticate to Continue</h2>
            <form
              action={async () => {
                "use server"
                await signIn("min-idp")
              }}
            >
              <button
                type="submit"
                className="group relative inline-flex items-center justify-center px-8 py-4 font-bold text-white transition-all duration-200 bg-indigo-600 font-pj rounded-xl focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-600 hover:bg-indigo-500 overflow-hidden"
              >
                <div className="absolute inset-0 w-full h-full -mt-1 rounded-lg opacity-30 bg-gradient-to-b from-transparent via-transparent to-black"></div>
                <span className="relative flex items-center gap-2">
                  <svg className="w-5 h-5 transition-transform group-hover:rotate-12" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"></path></svg>
                  Sign in with Min-IDP
                </span>
              </button>
            </form>
          </div>
        ) : (
          <div className="space-y-6 animate-in fade-in slide-in-from-bottom-8 duration-700">
            <div className="bg-slate-900/50 backdrop-blur-xl rounded-3xl p-8 border border-slate-800 shadow-2xl">
              <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between mb-8 pb-6 border-b border-slate-800 gap-4">
                <div>
                  <h2 className="text-2xl font-bold text-white mb-1">Authentication Success</h2>
                  <p className="text-slate-400">You are securely signed in via Min-IDP.</p>
                </div>
                <form
                  action={async () => {
                    "use server"
                    // @ts-expect-error Session property extension
                    const idToken = session?.idToken as string | undefined
                    if (idToken) {
                      // RP-initiated logout: clear SP session then redirect to IdP
                      // so the IdP session is also terminated.
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
                    className="px-6 py-2.5 rounded-lg text-sm font-semibold text-slate-300 bg-slate-800 hover:bg-slate-700 hover:text-white transition-all duration-200 ring-1 ring-slate-700/50 whitespace-nowrap"
                  >
                    Sign Out
                  </button>
                </form>
              </div>

              <div className="space-y-6">
                <div className="group">
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="text-sm font-medium text-slate-400 uppercase tracking-wider">ID Token</h3>
                    <span className="text-xs px-2.5 py-1 rounded-full bg-indigo-500/10 text-indigo-400 ring-1 ring-indigo-500/20">JWT</span>
                  </div>
                  <div className="relative rounded-xl overflow-hidden bg-slate-950 ring-1 ring-slate-800/50 group-hover:ring-slate-700 transition-all duration-200">
                    <div className="absolute top-0 left-0 w-1 h-full bg-gradient-to-b from-indigo-500 to-purple-600"></div>
                    <pre className="p-4 overflow-x-auto text-sm font-mono text-indigo-300 whitespace-pre-wrap break-all">
                      {/* @ts-expect-error Session property extension */}
                      {session?.idToken || "No ID Token returned"}
                    </pre>
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="group">
                    <h3 className="text-sm font-medium text-slate-400 uppercase tracking-wider mb-2">Session Profile</h3>
                    <div className="relative rounded-xl overflow-hidden bg-slate-950 ring-1 ring-slate-800/50 group-hover:ring-slate-700 transition-all duration-200 h-full">
                      <div className="absolute top-0 left-0 w-1 h-full bg-gradient-to-b from-emerald-500 to-teal-600"></div>
                      <pre className="p-4 overflow-x-auto text-xs font-mono text-emerald-300 h-full">
                        {JSON.stringify(session?.user, null, 2)}
                      </pre>
                    </div>
                  </div>
                  
                  <div className="group">
                    <h3 className="text-sm font-medium text-slate-400 uppercase tracking-wider mb-2">Access Token</h3>
                    <div className="relative rounded-xl overflow-hidden bg-slate-950 ring-1 ring-slate-800/50 group-hover:ring-slate-700 transition-all duration-200 h-full">
                      <div className="absolute top-0 left-0 w-1 h-full bg-gradient-to-b from-amber-500 to-orange-600"></div>
                      <pre className="p-4 overflow-x-auto text-xs font-mono text-amber-300 h-full whitespace-pre-wrap break-all">
                        {/* @ts-expect-error Session property extension */}
                        {session?.accessToken || "No Access Token returned"}
                      </pre>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
