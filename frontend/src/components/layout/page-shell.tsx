import type { ReactNode } from "react"

interface PageShellProps {
  children: ReactNode
}

export function PageShell({ children }: PageShellProps) {
  return (
    <main className="max-w-5xl mx-auto px-4 py-6">
      {children}
    </main>
  )
}
