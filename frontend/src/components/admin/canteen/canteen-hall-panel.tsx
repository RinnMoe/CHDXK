import { useState } from "react"

import type { CanteenHallInput } from "@/api/canteen"
import {
  emptyHall,
  selectClass,
} from "@/components/admin/canteen/canteen-admin-shared"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import {
  useCanteenHalls,
  useCreateCanteenHall,
  useUpdateCanteenHall,
} from "@/hooks/use-canteen-catalog"

export function CanteenHallPanel() {
  const halls = useCanteenHalls()
  const createHall = useCreateCanteenHall()
  const updateHall = useUpdateCanteenHall()
  const [form, setForm] = useState<CanteenHallInput>(emptyHall)
  const [editingID, setEditingID] = useState<number | null>(null)
  const [error, setError] = useState<string | null>(null)

  async function submit() {
    setError(null)
    try {
      if (editingID) {
        await updateHall.mutateAsync({ hallID: editingID, input: form })
      } else {
        await createHall.mutateAsync(form)
      }
      setForm(emptyHall)
      setEditingID(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : "保存食堂失败")
    }
  }

  function edit(hall: NonNullable<typeof halls.data>[number]) {
    setEditingID(hall.id)
    setForm({
      campus_key: hall.campus_key,
      name: hall.name,
      short_name: hall.short_name,
      address: hall.address,
      status: hall.status,
      display_order: hall.display_order,
    })
  }

  return (
    <div className="grid gap-6 lg:grid-cols-[1fr_1.4fr]">
      {error && (
        <p className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive lg:col-span-2">
          {error}
        </p>
      )}
      <Card size="sm">
        <CardHeader>
          <h2 className="text-base font-semibold">
            {editingID ? "编辑食堂" : "新增食堂"}
          </h2>
        </CardHeader>
        <CardContent>
          <form
            className="space-y-3"
            onSubmit={(event) => {
              event.preventDefault()
              void submit()
            }}
          >
            <Input
              placeholder="校区标识（如 benbu）"
              value={form.campus_key}
              onChange={(event) =>
                setForm({ ...form, campus_key: event.target.value })
              }
            />
            <Input
              placeholder="食堂名称"
              value={form.name}
              onChange={(event) => setForm({ ...form, name: event.target.value })}
            />
            <Input
              placeholder="简称（可选）"
              value={form.short_name ?? ""}
              onChange={(event) =>
                setForm({ ...form, short_name: event.target.value })
              }
            />
            <Input
              placeholder="地址（可选）"
              value={form.address ?? ""}
              onChange={(event) =>
                setForm({ ...form, address: event.target.value })
              }
            />
            <div className="flex items-center gap-2">
              <select
                className={selectClass}
                value={form.status ?? "active"}
                onChange={(event) =>
                  setForm({ ...form, status: event.target.value })
                }
              >
                <option value="active">开放</option>
                <option value="offline">停用</option>
                <option value="hidden">隐藏</option>
              </select>
              <Input
                type="number"
                placeholder="排序"
                value={form.display_order ?? 0}
                onChange={(event) =>
                  setForm({ ...form, display_order: Number(event.target.value) })
                }
              />
            </div>
            <div className="flex gap-2">
              <Button type="submit" size="sm">
                {editingID ? "保存修改" : "新增食堂"}
              </Button>
              {editingID && (
                <Button
                  type="button"
                  size="sm"
                  variant="outline"
                  onClick={() => {
                    setEditingID(null)
                    setForm(emptyHall)
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
          <h2 className="text-base font-semibold">食堂列表</h2>
        </CardHeader>
        <CardContent className="space-y-2">
          {(halls.data ?? []).length === 0 && (
            <p className="text-sm text-muted-foreground">
              还没有食堂，先新增一个。
            </p>
          )}
          {(halls.data ?? []).map((hall) => (
            <div
              key={hall.id}
              className="flex items-center justify-between gap-3 rounded-md border p-3 text-sm"
            >
              <div>
                <p className="font-medium">{hall.name}</p>
                <p className="text-xs text-muted-foreground">
                  {hall.campus_key} · {hall.store_count} 家店铺
                </p>
              </div>
              <div className="flex items-center gap-2">
                <Badge variant={hall.status === "active" ? "secondary" : "outline"}>
                  {hall.status}
                </Badge>
                <Button
                  type="button"
                  size="xs"
                  variant="outline"
                  onClick={() => edit(hall)}
                >
                  编辑
                </Button>
              </div>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  )
}
