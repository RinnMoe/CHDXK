import { Navigate, useSearchParams } from "react-router-dom"
import { AdminUserQueryTab } from "@/components/admin/admin-user-query-tab"
import { AdminUsersTab } from "@/components/admin/admin-users-tab"
import { SystemApiKeysTab } from "@/components/admin/system-api-keys-tab"
import { PageTitle } from "@/components/common/page-title"
import { PageShell } from "@/components/layout/page-shell"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useAuth } from "@/contexts/auth-context"

const adminTabs = ["user", "admin", "system-api-key"] as const
type AdminTab = (typeof adminTabs)[number]

function getAdminTab(value: string | null): AdminTab {
  return adminTabs.includes(value as AdminTab) ? (value as AdminTab) : "user"
}

export function UserAdminPage() {
  const { user, isLoading: authLoading } = useAuth()
  const [searchParams, setSearchParams] = useSearchParams()
  const activeTab = getAdminTab(searchParams.get("tab"))

  function setActiveTab(tab: string) {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev)
        next.set("tab", getAdminTab(tab))
        return next
      },
      { replace: true }
    )
  }

  if (authLoading) return null
  if (!user) return <Navigate to="/login" replace />
  if (user.role !== "admin") {
    return (
      <>
        <PageTitle>用户管理</PageTitle>
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
      <PageTitle>用户管理</PageTitle>
      <PageShell>
        <div className="space-y-6">
          <h1 className="text-2xl font-semibold">用户管理</h1>

          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList className="w-full justify-start overflow-x-auto sm:w-fit">
              <TabsTrigger value="user">用户查询</TabsTrigger>
              <TabsTrigger value="admin">管理员查询</TabsTrigger>
              <TabsTrigger value="system-api-key">系统 API Key</TabsTrigger>
            </TabsList>

            <TabsContent value="user" className="space-y-6">
              <AdminUserQueryTab currentUserID={user.id} />
            </TabsContent>

            <TabsContent value="admin" className="space-y-3">
              <AdminUsersTab currentUserID={user.id} />
            </TabsContent>

            <TabsContent value="system-api-key" className="space-y-4">
              <SystemApiKeysTab />
            </TabsContent>
          </Tabs>
        </div>
      </PageShell>
    </>
  )
}
