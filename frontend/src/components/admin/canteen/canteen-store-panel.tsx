import { useState } from "react"

import type { CanteenStoreInput } from "@/api/canteen"
import {
  emptyStore,
  selectClass,
} from "@/components/admin/canteen/canteen-admin-shared"
import { PaginationComponent } from "@/components/common/pagination"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import {
  useAdminCanteenStores,
  useCanteenHalls,
  useCreateCanteenStore,
  useMergeCanteenStore,
  useUpdateCanteenStore,
  useUpdateCanteenStoreStatus,
} from "@/hooks/use-canteen-catalog"

export function CanteenStorePanel() {
  const halls = useCanteenHalls()
  const createStore = useCreateCanteenStore()
  const updateStore = useUpdateCanteenStore()
  const updateStatus = useUpdateCanteenStoreStatus()
  const mergeStore = useMergeCanteenStore()

  const [form, setForm] = useState<CanteenStoreInput>(emptyStore)
  const [editingID, setEditingID] = useState<number | null>(null)
  const [categoriesText, setCategoriesText] = useState("")
  const [error, setError] = useState<string | null>(null)

  const [queryDraft, setQueryDraft] = useState("")
  const [query, setQuery] = useState("")
  const [status, setStatus] = useState("")
  const [page, setPage] = useState(1)
  const [mergeTargets, setMergeTargets] = useState<Record<number, number>>({})

  const stores = useAdminCanteenStores({
    q: query || undefined,
    status: status || undefined,
    page,
    page_size: 20,
  })
  const hallOptions = halls.data ?? []

  function report(err: unknown, fallback: string) {
    setError(err instanceof Error ? err.message : fallback)
  }

  async function submit() {
    setError(null)
    const categories = categoriesText
      .split(/[,，]/)
      .map((item) => item.trim())
      .filter(Boolean)
    const input = { ...form, categories }
    try {
      if (editingID) {
        await updateStore.mutateAsync({ storeID: editingID, input })
      } else {
        await createStore.mutateAsync(input)
      }
      setForm(emptyStore)
      setCategoriesText("")
      setEditingID(null)
    } catch (err) {
      report(err, "保存店铺失败")
    }
  }

  function edit(store: {
    id: number
    hall_id: number
    name: string
    floor: string
    categories: string[]
    status: string
  }) {
    setEditingID(store.id)
    setForm({
      hall_id: store.hall_id,
      name: store.name,
      floor: store.floor,
      categories: store.categories,
      status: store.status,
    })
    setCategoriesText(store.categories.join("，"))
  }

  async function changeStatus(storeID: number, nextStatus: string) {
    setError(null)
    try {
      await updateStatus.mutateAsync({ storeID, status: nextStatus })
    } catch (err) {
      report(err, "更新店铺状态失败")
    }
  }

  async function merge(storeID: number) {
    const targetID = mergeTargets[storeID]
    if (!targetID) {
      setError("请选择要合并到的目标店铺")
      return
    }
    setError(null)
    try {
      await mergeStore.mutateAsync({ storeID, targetID })
    } catch (err) {
      report(err, "合并店铺失败")
    }
  }

  return (
    <div className="space-y-4">
      {error && (
        <p className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {error}
        </p>
      )}

      <Card size="sm">
        <CardHeader>
          <h2 className="text-base font-semibold">
            {editingID ? "编辑店铺" : "新增店铺"}
          </h2>
        </CardHeader>
        <CardContent>
          <form
            className="grid gap-3 md:grid-cols-2"
            onSubmit={(event) => {
              event.preventDefault()
              void submit()
            }}
          >
            <select
              className={selectClass}
              value={form.hall_id || ""}
              onChange={(event) =>
                setForm({ ...form, hall_id: Number(event.target.value) })
              }
            >
              <option value="">选择食堂</option>
              {hallOptions.map((hall) => (
                <option key={hall.id} value={hall.id}>
                  {hall.name}
                </option>
              ))}
            </select>
            <Input
              placeholder="店铺名称"
              value={form.name}
              onChange={(event) => setForm({ ...form, name: event.target.value })}
            />
            <Input
              placeholder="楼层/位置（可选）"
              value={form.floor ?? ""}
              onChange={(event) =>
                setForm({ ...form, floor: event.target.value })
              }
            />
            <Input
              placeholder="分类，用逗号分隔"
              value={categoriesText}
              onChange={(event) => setCategoriesText(event.target.value)}
            />
            <select
              className={selectClass}
              value={form.status ?? "pending"}
              onChange={(event) =>
                setForm({ ...form, status: event.target.value })
              }
            >
              <option value="pending">待审核</option>
              <option value="active">上架</option>
              <option value="offline">下架</option>
              <option value="rejected">已拒绝</option>
            </select>
            <div className="flex gap-2 md:col-span-2">
              <Button type="submit" size="sm">
                {editingID ? "保存修改" : "新增店铺"}
              </Button>
              {editingID && (
                <Button
                  type="button"
                  size="sm"
                  variant="outline"
                  onClick={() => {
                    setEditingID(null)
                    setForm(emptyStore)
                    setCategoriesText("")
                  }}
                >
                  取消编辑
                </Button>
              )}
            </div>
          </form>
        </CardContent>
      </Card>

      <Card size="sm">
        <CardHeader>
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="text-base font-semibold">店铺列表</h2>
            <div className="ml-auto flex items-center gap-2">
              <Input
                className="w-48"
                placeholder="搜索店铺"
                value={queryDraft}
                onChange={(event) => setQueryDraft(event.target.value)}
              />
              <select
                className={selectClass}
                value={status}
                onChange={(event) => {
                  setStatus(event.target.value)
                  setPage(1)
                }}
              >
                <option value="">全部状态</option>
                <option value="pending">待审核</option>
                <option value="active">上架</option>
                <option value="offline">下架</option>
                <option value="rejected">已拒绝</option>
                <option value="merged">已合并</option>
              </select>
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={() => {
                  setQuery(queryDraft)
                  setPage(1)
                }}
              >
                搜索
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-2">
          {(stores.data?.items ?? []).length === 0 && (
            <p className="text-sm text-muted-foreground">没有店铺。</p>
          )}
          {(stores.data?.items ?? []).map((store) => (
            <div
              key={store.id}
              className="space-y-2 rounded-md border p-3 text-sm"
            >
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <p className="font-medium">{store.name}</p>
                  <p className="text-xs text-muted-foreground">
                    {store.hall_name}
                    {store.floor ? ` · ${store.floor}` : ""} ·{" "}
                    {store.rating.count} 人评价
                  </p>
                </div>
                <div className="flex flex-wrap items-center gap-2">
                  <Badge variant="outline">{store.status}</Badge>
                  <Button
                    type="button"
                    size="xs"
                    variant="outline"
                    onClick={() => edit(store)}
                  >
                    编辑
                  </Button>
                  <Button
                    type="button"
                    size="xs"
                    variant="outline"
                    onClick={() => void changeStatus(store.id, "active")}
                  >
                    上架
                  </Button>
                  <Button
                    type="button"
                    size="xs"
                    variant="outline"
                    onClick={() => void changeStatus(store.id, "offline")}
                  >
                    下架
                  </Button>
                </div>
              </div>
              {store.status !== "merged" && (
                <div className="flex flex-wrap items-center gap-2">
                  <select
                    className={selectClass}
                    value={mergeTargets[store.id] ?? ""}
                    onChange={(event) =>
                      setMergeTargets({
                        ...mergeTargets,
                        [store.id]: Number(event.target.value),
                      })
                    }
                  >
                    <option value="">合并到…</option>
                    {(stores.data?.items ?? [])
                      .filter((item) => item.id !== store.id)
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
                    onClick={() => void merge(store.id)}
                  >
                    合并
                  </Button>
                </div>
              )}
            </div>
          ))}
          {stores.data && (
            <PaginationComponent
              page={stores.data.page}
              pageSize={stores.data.page_size}
              total={stores.data.total}
              onPageChange={setPage}
            />
          )}
        </CardContent>
      </Card>
    </div>
  )
}
