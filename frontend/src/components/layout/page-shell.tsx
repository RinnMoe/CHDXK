import type { ReactNode } from "react"
import { AnnouncementBanner } from "@/components/announcement/announcement-banner"

interface PageShellProps {
  children: ReactNode
  showAnnouncements?: boolean
}

export function PageShell({ children, showAnnouncements = true }: PageShellProps) {
  return (
    <main className="max-w-5xl mx-auto px-4 py-6">
      {showAnnouncements && <AnnouncementBanner />}
      {children}
    </main>
  )
}
