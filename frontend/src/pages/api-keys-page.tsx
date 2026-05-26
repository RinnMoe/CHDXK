import { useState } from "react"
import { Navigate } from "react-router-dom"
import { RiAddLine } from "@remixicon/react"
import { ApiKeyCreateDialog } from "@/components/api-key/api-key-create-dialog"
import { ApiKeyTable } from "@/components/api-key/api-key-table"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { Button } from "@/components/ui/button"
import { useAuth } from "@/contexts/auth-context"
import { useLoginRedirectPath } from "@/hooks/use-login-redirect"
import {
  useApiKeys,
  useCreateApiKey,
  useDeleteApiKey,
} from "@/hooks/use-api-key"

export function ApiKeysPage() {
  const { user, isLoading: authLoading } = useAuth()
  const loginRedirectPath = useLoginRedirectPath()
  const { data: apiKeys = [], isLoading } = useApiKeys(!!user)
  const createMutation = useCreateApiKey()
  const deleteMutation = useDeleteApiKey()

  const [isCreateOpen, setIsCreateOpen] = useState(false)

  if (authLoading) return null
  if (!user) return <Navigate to={loginRedirectPath} replace />

  return (
    <>
      <PageTitle>API Keys</PageTitle>
      <PageShell>
        <div className="space-y-6">
          <div className="space-y-2">
            <h1 className="text-2xl font-semibold">API Keys</h1>
            <p className="text-sm text-muted-foreground">
              API Key
              可用于脚本或外部工具访问选课社区接口，请妥善保管并定期清理不再使用的
              Key。
            </p>
          </div>

          <ApiKeyCreateDialog
            open={isCreateOpen}
            onOpenChange={setIsCreateOpen}
            onCreate={(cmd) => createMutation.mutateAsync(cmd)}
            isCreating={createMutation.isPending}
            title="新建 API Key"
            description="创建成功后明文仅显示一次。"
            nameInputID="api-key-name"
            namePlaceholder="例如：本地脚本"
          />

          <section className="space-y-4">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
              <h2 className="text-lg font-medium">我的 API Keys</h2>
              <Button onClick={() => setIsCreateOpen(true)}>
                <RiAddLine data-icon="inline-start" />
                新建 API Key
              </Button>
            </div>

            <ApiKeyTable
              apiKeys={apiKeys}
              isLoading={isLoading}
              isDeleting={deleteMutation.isPending}
              emptyText="暂无 API Key"
              deleteAriaLabel="删除 API Key"
              deleteDialogTitle="删除 API Key"
              deleteDialogDescription="删除后使用该 key 的请求会立即失效。"
              deleteActionLabel="删除"
              onDelete={(id) => deleteMutation.mutateAsync(id)}
            />
          </section>
        </div>
      </PageShell>
    </>
  )
}
