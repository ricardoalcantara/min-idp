"use client"

import { useEffect, useState } from "react"

export function ThemeToggle() {
  const [dark, setDark] = useState(false)

  useEffect(() => {
    const stored = localStorage.getItem("theme")
    const prefersDark = window.matchMedia("(prefers-color-scheme: dark)").matches
    const isDark = stored ? stored === "dark" : prefersDark
    setDark(isDark)
    document.documentElement.classList.toggle("dark", isDark)
  }, [])

  function toggle() {
    const next = !dark
    setDark(next)
    document.documentElement.classList.toggle("dark", next)
    localStorage.setItem("theme", next ? "dark" : "light")
  }

  return (
    <button
      onClick={toggle}
      aria-label="Toggle dark mode"
      className="fixed top-4 right-4 w-9 h-9 rounded-full flex items-center justify-center text-lg transition-colors"
      style={{ background: "var(--card-border)" }}
    >
      {dark ? "☀️" : "🌙"}
    </button>
  )
}
