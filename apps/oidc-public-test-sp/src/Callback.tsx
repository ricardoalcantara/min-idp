import { useEffect, useState } from "react"
import { handleCallback } from "./oidc"
import type { SessionData } from "./oidc"

type Props = {
  onComplete: (session: SessionData) => void
  onError: (message: string) => void
}

export function Callback({ onComplete, onError }: Props) {
  const [working, setWorking] = useState(true)

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const session = await handleCallback(window.location.search)
        if (!cancelled) {
          onComplete(session)
          window.history.replaceState({}, "", "/")
        }
      } catch (err) {
        if (!cancelled) {
          onError(err instanceof Error ? err.message : "Callback failed")
        }
      } finally {
        if (!cancelled) setWorking(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [onComplete, onError])

  if (working) {
    return <p>Completing sign-in…</p>
  }
  return null
}
