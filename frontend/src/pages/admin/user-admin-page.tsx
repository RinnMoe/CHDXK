import { getRouteApi, useNavigate } from "@tanstack/react-router"
import { AuditLogsTab } from "@/components/admin/audit-logs-tab"
import { AnnouncementsTab } from "@/components/admin/announcements-tab"
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
        <PageTitle>管理</PageTitle>
        <PageShell>
          <p className="py-12 text-center text-muted-foreground">
            需要管理员权限
          </p>
        </PageShell>
      </>
    )
  }

  return (
    <>
      <PageTitle>管理</PageTitle>
      <PageShell>
        <div className="space-y-6">
          <h1 className="text-2xl font-semibold">管理</h1>

          <Card>
            <CardHeader>
              <CardTitle>用户和角色由主站管理</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm leading-6 text-muted-foreground">
                res-hub 是账号、角色、会话和账户状态的唯一来源。本站管理员权限会
                按当前主站会话生效，用户、角色和会话状态请在主站管理。
              </p>
            </CardContent>
          </Card>

          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList className="w-full justify-start overflow-x-auto sm:w-fit">
              <TabsTrigger value="system-api-key">系统 API Key</TabsTrigger>
              <TabsTrigger value="system-settings">系统设置</TabsTrigger>
              <TabsTrigger value="announcement">公告</TabsTrigger>
              <TabsTrigger value="audit-log">审计日志</TabsTrigger>
            </TabsList>

            <TabsContent value="system-api-key" className="space-y-4">
              <SystemApiKeysTab />
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
