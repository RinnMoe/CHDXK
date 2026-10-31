import { useState } from "react"

import { selectClass } from "@/components/admin/canteen/canteen-admin-shared"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { useAdminCanteenStores } from "@/hooks/use-canteen-catalog"
import {
  useAdminCanteenDishes,
  useMergeAdminCanteenDish,
  useUpdateAdminCanteenDishStatus,
} from "@/hooks/use-canteen-dishes"

export function CanteenDishPanel() {
  const stores = useAdminCanteenStores({ page: 1, page_size: 100 })
  const [storeID, setStoreID] = useState(0)
  const [status, setStatus] = useState("")
  const [mergeTargets, setMergeTargets] = useState<Record<number, number>>({})
  const [error, setError] = useState<string | null>(null)

  const dishes = useAdminCanteenDishes({
    store_id: storeID || undefined,
    status: status || undefined,
  })
  const updateStatus = useUpdateAdminCanteenDishStatus()
  const mergeDish = useMergeAdminCanteenDish()

  async function changeStatus(dishID: number, nextStatus: string) {
    setError(null)
    try {
      await updateStatus.mutateAsync({ dishID, status: nextStatus })
    } catch (err) {
      setError(err instanceof Error ? err.message : "更新菜品状态失败")
    }
  }

  async function merge(dishID: number) {
    const targetID = mergeTargets[dishID]
    if (!targetID) {
      setError("请选择要合并到的目标菜品")
      return
    }
    setError(null)
    try {
      await mergeDish.mutateAsync({ dishID, targetID })
    } catch (err) {
      setError(err instanceof Error ? err.message : "合并菜品失败")
    }
  }

  return (
    <div className="space-y-4">
      {error && <p className="text-sm text-destructive">{error}</p>}
      <Card size="sm">
        <CardHeader>
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="text-base font-semibold">菜品管理</h2>
            <div className="ml-auto flex items-center gap-2">
              <select
                className={selectClass}
                value={storeID || ""}
                onChange={(event) => setStoreID(Number(event.target.value))}
              >
                <option value="">选择店铺</option>
                {(stores.data?.items ?? []).map((store) => (
                  <option key={store.id} value={store.id}>
                    {store.hall_name} · {store.name}
                  </option>
                ))}
              </select>
              <select
                className={selectClass}
                value={status}
                onChange={(event) => setStatus(event.target.value)}
              >
                <option value="">全部状态</option>
                <option value="pending">待审核</option>
                <option value="active">已上架</option>
                <option value="hidden">已隐藏</option>
                <option value="rejected">已拒绝</option>
                <option value="merged">已合并</option>
              </select>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-2">
          {storeID === 0 && (
            <p className="text-sm text-muted-foreground">
              先选择一个店铺查看菜品。
            </p>
          )}
          {(dishes.data?.items ?? []).map((dish) => (
            <div
              key={dish.id}
              className="space-y-2 rounded-md border p-3 text-sm"
            >
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <span className="font-medium">{dish.name}</span>
                  <span className="ml-2 text-xs text-muted-foreground">
                    {dish.review_count} 次评分 · {dish.rating_score.toFixed(1)} 分
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <Badge variant="outline">{dish.status}</Badge>
                  <Button
                    type="button"
                    size="xs"
                    variant="outline"
                    onClick={() => void changeStatus(dish.id, "active")}
                  >
                    上架
                  </Button>
                  <Button
                    type="button"
                    size="xs"
                    variant="outline"
                    onClick={() => void changeStatus(dish.id, "hidden")}
                  >
                    隐藏
                  </Button>
                </div>
              </div>
              {dish.status !== "merged" && (
                <div className="flex flex-wrap items-center gap-2">
                  <select
                    className={selectClass}
                    value={mergeTargets[dish.id] ?? ""}
                    onChange={(event) =>
                      setMergeTargets({
                        ...mergeTargets,
                        [dish.id]: Number(event.target.value),
                      })
                    }
                  >
                    <option value="">合并到…</option>
                    {(dishes.data?.items ?? [])
                      .filter((item) => item.id !== dish.id && item.status !== "merged")
                      .map((item) => (
                        <option key={item.id} value={item.id}>
                          {item.name}
                        </option>
                      ))}
                  </select>
                  <Button
                    type="button"
                    size="xs"
                    variant="destructive"
                    onClick={() => void merge(dish.id)}
                  >
                    合并
                  </Button>
                </div>
              )}
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  )
}
