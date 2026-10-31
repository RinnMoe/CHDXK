import type { CanteenStoreListItemDTO } from "@/api/canteen"
import { CanteenEmptyState } from "@/components/canteen/canteen-empty-state"
import { CanteenStoreCard } from "@/components/canteen/canteen-store-card"

export function CanteenStoreList({
  stores,
  emptyTitle = "暂无店铺",
  emptyDescription = "管理员添加店铺后，这里会展示真实数据。",
}: {
  stores: CanteenStoreListItemDTO[]
  emptyTitle?: string
  emptyDescription?: string
}) {
  if (stores.length === 0) {
    return (
      <CanteenEmptyState title={emptyTitle} description={emptyDescription} />
    )
  }
  return (
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
      {stores.map((store) => (
        <CanteenStoreCard key={store.id} store={store} />
      ))}
    </div>
  )
}
