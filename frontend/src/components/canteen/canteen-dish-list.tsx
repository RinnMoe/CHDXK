import { Link } from "@tanstack/react-router"
import { useState } from "react"

import { CanteenDishReviewForm } from "@/components/canteen/canteen-dish-review-form"
import { CanteenEmptyState } from "@/components/canteen/canteen-empty-state"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import {
  useCreateCanteenDish,
  useStoreDishes,
} from "@/hooks/use-canteen-dishes"

export function CanteenDishList({
  storeID,
  canContribute,
}: {
  storeID: number
  canContribute: boolean
}) {
  const dishes = useStoreDishes(storeID)
  const create = useCreateCanteenDish(storeID)
  const [name, setName] = useState("")
  const [message, setMessage] = useState<string | null>(null)
  const [activeDish, setActiveDish] = useState<number | null>(null)
  const [error, setError] = useState<string | null>(null)

  async function addDish() {
    const trimmed = name.trim()
    if (!trimmed) return
    setError(null)
    setMessage(null)
    try {
      const result = await create.mutateAsync(trimmed)
      setName("")
      setActiveDish(result.id)
      setMessage(
        result.existing
          ? "已有同名菜品，已为你选中"
          : result.status === "active"
            ? "菜品已添加"
            : "已提交，等待管理员审核"
      )
    } catch (err) {
      setError(err instanceof Error ? err.message : "添加菜品失败")
    }
  }

  const items = dishes.data?.items ?? []

  return (
    <div className="space-y-3">
      {items.length === 0 ? (
        <CanteenEmptyState
          title="还没有菜品"
          description="可以添加你吃到的菜品，审核后进入菜品图鉴。"
        />
      ) : (
        <div className="space-y-2">
          {items.map((dish) => (
            <Card key={dish.id} size="sm">
              <CardContent className="space-y-2 pt-6">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div className="flex items-center gap-3">
                    {dish.cover_url ? (
                      <img
                        src={dish.cover_url}
                        alt={dish.name}
                        className="size-14 rounded-md border object-cover"
                      />
                    ) : (
                      <div className="flex size-14 items-center justify-center rounded-md bg-muted text-xs text-muted-foreground">
                        暂无图
                      </div>
                    )}
                    <div className="space-y-1">
                      <Link
                        to="/canteen/stores/$storeID/dishes/$dishID"
                        params={{ storeID: String(storeID), dishID: String(dish.id) }}
                        className="font-medium hover:underline"
                      >
                        {dish.name}
                      </Link>
                      <div className="flex items-center gap-2">
                        {dish.review_count > 0 && (
                          <Badge variant="secondary">
                            {dish.rating_score.toFixed(1)} 分 · {dish.rating_count} 人
                          </Badge>
                        )}
                        {dish.photo_count > 0 && (
                          <span className="text-xs text-muted-foreground">
                            {dish.photo_count} 张实拍
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                  {canContribute && (
                    <Button
                      type="button"
                      size="xs"
                      variant="outline"
                      onClick={() =>
                        setActiveDish(activeDish === dish.id ? null : dish.id)
                      }
                    >
                      评分
                    </Button>
                  )}
                </div>
                {dish.review_count > 0 && (
                  <p className="text-xs text-muted-foreground">
                    口味 {dish.taste_avg.toFixed(1)} · 分量{" "}
                    {dish.portion_avg.toFixed(1)} · 性价比{" "}
                    {dish.value_avg.toFixed(1)}
                  </p>
                )}
                {activeDish === dish.id && (
                  <CanteenDishReviewForm
                    dishID={dish.id}
                    onDone={() => setActiveDish(null)}
                  />
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {canContribute && (
        <div className="flex flex-wrap items-center gap-2">
          <Input
            className="max-w-64"
            placeholder="添加菜品，如：牛肉面"
            value={name}
            maxLength={60}
            onChange={(event) => setName(event.target.value)}
          />
          <Button
            type="button"
            size="sm"
            disabled={create.isPending}
            onClick={() => void addDish()}
          >
            {create.isPending ? "提交中…" : "添加菜品"}
          </Button>
        </div>
      )}
      {message && <p className="text-sm text-muted-foreground">{message}</p>}
      {error && <p className="text-sm text-destructive">{error}</p>}
    </div>
  )
}
