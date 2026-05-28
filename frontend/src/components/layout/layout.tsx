import { Outlet } from "@tanstack/react-router"
import { SiteHeader } from "@/components/layout/site-header"
import { SiteFooter } from "@/components/layout/site-footer"
import { GlobalShortcuts } from "@/components/shortcuts/global-shortcuts"

export function Layout() {
  return (
    <div className="flex min-h-svh flex-col">
      <SiteHeader />
      <GlobalShortcuts />
      <div className="flex-1">
        <Outlet />
      </div>
      <SiteFooter />
    </div>
  )
}
