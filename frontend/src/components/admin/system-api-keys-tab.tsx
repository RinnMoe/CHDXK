import { copy } from "@/lib/copy"
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
        <h2 className="text-lg font-medium">
          {copy("copy.admin.components.admin.system-api-keys-tab.705fb8d9afda")}
        </h2>
        <Button onClick={() => setIsCreateOpen(true)}>
          <RiAddLine data-icon="inline-start" />
          {copy("copy.admin.components.admin.system-api-keys-tab.51d906a9ecd8")}
        </Button>
      </div>

      <ApiKeyCreateDialog
        open={isCreateOpen}
        onOpenChange={setIsCreateOpen}
        onCreate={(cmd) => createMutation.mutateAsync(cmd)}
        isCreating={createMutation.isPending}
        title={copy(
          "copy.admin.components.admin.system-api-keys-tab.921f96d4e427"
        )}
        description={copy(
          "copy.admin.components.admin.system-api-keys-tab.98283370da3e"
        )}
        nameInputID="system-api-key-name"
        namePlaceholder={copy(
          "copy.admin.components.admin.system-api-keys-tab.5bae98259f10"
        )}
        getErrorMessage={getErrorMessage}
      />

      <ApiKeyTable
        apiKeys={systemKeysQuery.data ?? []}
        isLoading={systemKeysQuery.isLoading}
        isDeleting={deleteMutation.isPending}
        emptyText={copy(
          "copy.admin.components.admin.system-api-keys-tab.c6a9fed1e7d8"
        )}
        deleteAriaLabel={copy(
          "copy.admin.components.admin.system-api-keys-tab.7951c979a42d"
        )}
        deleteDialogTitle={copy(
          "copy.admin.components.admin.system-api-keys-tab.7951c979a42d"
        )}
        deleteDialogDescription={copy(
          "copy.admin.components.admin.system-api-keys-tab.0303d805e868"
        )}
        deleteActionLabel={copy(
          "copy.admin.components.admin.system-api-keys-tab.a48fa6108b57"
        )}
        onDelete={(id) => deleteMutation.mutateAsync(id)}
        getDeleteErrorMessage={getErrorMessage}
      />
    </section>
  )
}
