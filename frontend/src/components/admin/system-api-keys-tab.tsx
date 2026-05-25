import { useState } from "react"
import { RiAddLine } from "@remixicon/react"
import { ApiKeyCreateDialog } from "@/components/api-key/api-key-create-dialog"
import { ApiKeyTable } from "@/components/api-key/api-key-table"
import { Button } from "@/components/ui/button"
import {
  useCreateSystemApiKey,
  useDeleteSystemApiKey,
  useSystemApiKeys,
} from "@/hooks/use-api-key"
import { getErrorMessage } from "./admin-utils"

export function SystemApiKeysTab() {
  const [isCreateOpen, setIsCreateOpen] = useState(false)

  const systemKeysQuery = useSystemApiKeys()
  const createMutation = useCreateSystemApiKey()
  const deleteMutation = useDeleteSystemApiKey()

  return (
    <section className="space-y-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <h2 className="text-lg font-medium">系统 API Key</h2>
        <Button onClick={() => setIsCreateOpen(true)}>
          <RiAddLine data-icon="inline-start" />
          新建系统 Key
        </Button>
      </div>

      <ApiKeyCreateDialog
        open={isCreateOpen}
        onOpenChange={setIsCreateOpen}
        onCreate={(cmd) => createMutation.mutateAsync(cmd)}
        isCreating={createMutation.isPending}
        title="新建系统 API Key"
        description="创建成功后明文仅显示一次。"
        nameInputID="system-api-key-name"
        namePlaceholder="例如：外部积分查询服务"
        getErrorMessage={getErrorMessage}
      />

      <ApiKeyTable
        apiKeys={systemKeysQuery.data ?? []}
        isLoading={systemKeysQuery.isLoading}
        isDeleting={deleteMutation.isPending}
        emptyText="暂无系统 API Key"
        deleteAriaLabel="注销系统 API Key"
        deleteDialogTitle="注销系统 API Key"
        deleteDialogDescription="注销后使用该 key 的请求会立即失效。"
        deleteActionLabel="注销"
        onDelete={(id) => deleteMutation.mutateAsync(id)}
        getDeleteErrorMessage={getErrorMessage}
      />
    </section>
  )
}
