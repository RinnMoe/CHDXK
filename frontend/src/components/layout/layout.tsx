import { Outlet } from "react-router-dom"
import { SiteHeader } from "@/components/layout/site-header"

export function Layout() {
  return (
    <div className="min-h-svh flex flex-col">
      <SiteHeader />
      <div className="flex-1">
        <Outlet />
      </div>
    </div>
  )
}
