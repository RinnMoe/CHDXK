import { copy } from "@/lib/copy"
import { getRouteApi, useNavigate } from "@tanstack/react-router"
import { AuditLogsTab } from "@/components/admin/audit-logs-tab"
import { AnnouncementsTab } from "@/components/admin/announcements-tab"
import { CourseManagementTab } from "@/components/admin/course-management-tab"
import { SystemApiKeysTab } from "@/components/admin/system-api-keys-tab"
import { SystemSettingsTab } from "@/components/admin/system-settings-tab"
import { PageTitle } from "@/components/common/page-title"
import { PageShell } from "@/components/layout/page-shell"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useAuth } from "@/contexts/auth-context"

const adminTabs = [
  "system-api-key",
  "system-settings",
  "announcement",
  "audit-log",
  "course-data",
] as const
type AdminTab = (typeof adminTabs)[number]
const routeApi = getRouteApi("/app/admin/user")

function getAdminTab(value: string | null): AdminTab {
  return adminTabs.includes(value as AdminTab)
    ? (value as AdminTab)
    : "system-settings"
}

export function UserAdminPage() {
  const { user } = useAuth()
  const search = routeApi.useSearch()
  const navigate = useNavigate({ from: "/admin/user" })
  const activeTab = getAdminTab(search.tab ?? null)

  function setActiveTab(tab: string) {
    const nextTab = getAdminTab(tab)
    void navigate({
      search: (prev) => ({ ...prev, tab: nextTab }),
      replace: true,
      resetScroll: false,
    })
  }

  if (!user) return null
  if (!user.is_admin()) {
    return (
      <>
        <PageTitle>
          {copy("copy.admin.pages.admin.user-admin-page.bb6d995724f4")}
        </PageTitle>
        <PageShell>
          <p className="py-12 text-center text-muted-foreground">
            {copy("copy.admin.pages.admin.user-admin-page.9228f9aa4c65")}
          </p>
        </PageShell>
      </>
    )
  }

  return (
    <>
      <PageTitle>
        {copy("copy.admin.pages.admin.user-admin-page.bb6d995724f4")}
      </PageTitle>
      <PageShell>
        <div className="space-y-6">
          <h1 className="text-2xl font-semibold">
            {copy("copy.admin.pages.admin.user-admin-page.bb6d995724f4")}
          </h1>

          <Card>
            <CardHeader>
              <CardTitle>
                {copy("copy.admin.pages.admin.user-admin-page.5cb81812564c")}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm leading-6 text-muted-foreground">
                {copy("copy.admin.pages.admin.user-admin-page.aee480d85347")}
              </p>
            </CardContent>
          </Card>

          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList className="w-full justify-start overflow-x-auto sm:w-fit">
              <TabsTrigger value="course-data">
                {copy("copy.admin.pages.admin.user-admin-page.b4d79dbaebef")}
              </TabsTrigger>
              <TabsTrigger value="system-api-key">
                {copy("copy.admin.pages.admin.user-admin-page.705fb8d9afda")}
              </TabsTrigger>
              <TabsTrigger value="system-settings">
                {copy("copy.admin.pages.admin.user-admin-page.68ea5dd4d7af")}
              </TabsTrigger>
              <TabsTrigger value="announcement">
                {copy("copy.admin.pages.admin.user-admin-page.069c0ecbf898")}
              </TabsTrigger>
              <TabsTrigger value="audit-log">
                {copy("copy.admin.pages.admin.user-admin-page.a0f79e91f13f")}
              </TabsTrigger>
            </TabsList>

            <TabsContent value="system-api-key" className="space-y-4">
              <SystemApiKeysTab />
            </TabsContent>

            <TabsContent value="course-data" className="space-y-4">
              <CourseManagementTab />
            </TabsContent>

            <TabsContent value="system-settings" className="space-y-4">
              <SystemSettingsTab />
            </TabsContent>

            <TabsContent value="announcement" className="space-y-4">
              <AnnouncementsTab />
            </TabsContent>

            <TabsContent value="audit-log" className="space-y-4">
              <AuditLogsTab />
            </TabsContent>
          </Tabs>
        </div>
      </PageShell>
    </>
  )
}
