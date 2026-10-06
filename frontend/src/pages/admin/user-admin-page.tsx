import { UsersPanel } from "@/components/admin/users-panel"
import { copy } from "@/lib/copy"

import { getRouteApi, useNavigate } from "@tanstack/react-router"
import { AuditLogsTab } from "@/components/admin/audit-logs-tab"
import { AnnouncementsTab } from "@/components/admin/announcements-tab"
import { CourseManagementTab } from "@/components/admin/course-management-tab"
import { SystemApiKeysTab } from "@/components/admin/system-api-keys-tab"
import { SystemSettingsTab } from "@/components/admin/system-settings-tab"
import { CopySettingsPanel } from "@/components/admin/copy-settings-panel"
import { ThemeSettingsPanel } from "@/components/admin/theme-settings-panel"
import { ContentPagesPanel } from "@/components/admin/content-pages-panel"
import { TeacherProfilesPanel } from "@/components/admin/teacher-profiles-panel"
import { AdminNavigation } from "@/components/admin/admin-navigation"
import {
  getAdminTab,
  canAccessAdminTab,
  type AdminTab,
} from "@/components/admin/admin-navigation-model"
import { PageTitle } from "@/components/common/page-title"
import { PageShell } from "@/components/layout/page-shell"
import { useAuth } from "@/contexts/auth-context"

const routeApi = getRouteApi("/app/admin/user")

export function UserAdminPage() {
  const { user } = useAuth()
  const search = routeApi.useSearch()
  const navigate = useNavigate({ from: "/admin/user" })
  const requested = getAdminTab(search.tab)
  const active =
    user && canAccessAdminTab(requested, user.can) ? requested : "course-data"
  function setActive(tab: AdminTab) {
    void navigate({
      search: (previous) => ({ ...previous, tab }),
      replace: true,
      resetScroll: true,
    })
  }
  if (!user) return null
  return (
    <>
      <PageTitle>
        {copy("copy.admin.pages.admin.user-admin-page.bb6d995724f4")}
      </PageTitle>
      <PageShell>
        <h1 className="mb-6 text-2xl font-semibold">
          {copy("copy.admin.pages.admin.user-admin-page.bb6d995724f4")}
        </h1>
        {!user.is_admin() ? (
          <p className="py-12 text-center text-muted-foreground">
            {copy("copy.admin.pages.admin.user-admin-page.de06f8bc069d")}
          </p>
        ) : (
          <div className="grid items-start gap-6 lg:grid-cols-[12rem_minmax(0,1fr)]">
            <aside className="lg:sticky lg:top-20">
              <AdminNavigation
                active={active}
                onChange={setActive}
                can={user.can}
              />
            </aside>
            <div className="min-w-0">
              {active === "users" && <UsersPanel />}
              {active === "course-data" && <CourseManagementTab />}
              {active === "teacher-profiles" && <TeacherProfilesPanel />}
              {active === "announcement" && <AnnouncementsTab />}
              {active === "system-settings" && <SystemSettingsTab />}
              {active === "appearance" && <ThemeSettingsPanel />}
              {active === "site-copy" && <CopySettingsPanel />}
              {active === "content-pages" && <ContentPagesPanel />}
              {active === "system-api-key" && <SystemApiKeysTab />}
              {active === "audit-log" && <AuditLogsTab />}
            </div>
          </div>
        )}
      </PageShell>
    </>
  )
}
