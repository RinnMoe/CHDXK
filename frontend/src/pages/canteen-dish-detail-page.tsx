import { Link, getRouteApi } from "@tanstack/react-router"

import { CanteenDishReviewForm } from "@/components/canteen/canteen-dish-review-form"
import { CanteenEmptyState } from "@/components/canteen/canteen-empty-state"
import { PageTitle } from "@/components/common/page-title"
import { PageShell } from "@/components/layout/page-shell"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { useStoreDishDetail } from "@/hooks/use-canteen-dishes"

const routeApi = getRouteApi("/app/canteen/stores/$storeID/dishes/$dishID")

export function CanteenDishDetailPage() {
  const { storeID, dishID } = routeApi.useParams()
  const detail = useStoreDishDetail(Number(storeID), Number(dishID))

  if (detail.isLoading) {
    return (
      <PageShell>
        <p className="py-16 text-center text-sm text-muted-foreground">
          正在加载菜品…
        </p>
      </PageShell>
    )
  }

  if (detail.isError || !detail.data) {
    return (
      <PageShell>
        <div className="space-y-4 py-10 text-center">
          <p className="text-sm font-medium">菜品不存在或暂不可用</p>
          <Link
            to="/canteen/stores/$storeID"
            params={{ storeID }}
            className="text-sm text-primary hover:underline"
          >
            返回店铺
          </Link>
        </div>
      </PageShell>
    )
  }

  const { dish, photos, reviews } = detail.data

  return (
    <>
      <PageTitle>{dish.name}</PageTitle>
      <PageShell>
        <div className="space-y-6">
          <Link
            to="/canteen/stores/$storeID"
            params={{ storeID }}
            className="text-sm text-muted-foreground hover:underline"
          >
            ← 返回 {dish.store_name || "店铺"}
          </Link>

          <header className="flex flex-wrap items-start justify-between gap-4">
            <div className="space-y-1">
              <h1 className="text-2xl font-bold">{dish.name}</h1>
              <p className="text-sm text-muted-foreground">
                {dish.store_name}
                {dish.photo_count > 0 ? ` · ${dish.photo_count} 张实拍` : ""}
              </p>
            </div>
            <div className="text-right">
              <div className="text-3xl font-bold">
                {dish.rating_score.toFixed(1)}
              </div>
              <div className="text-xs text-muted-foreground">
                {dish.rating_count} 人评分 · {dish.review_count} 次评价
              </div>
            </div>
          </header>

          <div className="grid gap-3 sm:grid-cols-3">
            {[
              { label: "口味", value: dish.taste_avg },
              { label: "分量", value: dish.portion_avg },
              { label: "性价比", value: dish.value_avg },
            ].map((item) => (
              <Card key={item.label} size="sm">
                <CardContent className="flex items-center justify-between pt-6">
                  <span className="text-sm text-muted-foreground">
                    {item.label}
                  </span>
                  <span className="text-lg font-semibold">
                    {item.value.toFixed(1)}
                  </span>
                </CardContent>
              </Card>
            ))}
          </div>

          <section className="space-y-3">
            <h2 className="text-lg font-semibold">实拍</h2>
            {photos.length === 0 ? (
              <CanteenEmptyState
                title="还没有实拍"
                description="给这道菜评分时可以附上最多 3 张图片。"
              />
            ) : (
              <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4">
                {photos.map((photo) => (
                  <img
                    key={photo.id}
                    src={photo.url}
                    alt={`${dish.name} 实拍`}
                    className="h-40 w-full rounded-md border object-cover"
                  />
                ))}
              </div>
            )}
          </section>

          <section className="space-y-3">
            <h2 className="text-lg font-semibold">菜品评分</h2>
            {reviews.length === 0 ? (
              <CanteenEmptyState
                title="还没有菜品评分"
                description="成为第一个给这道菜打分的人。"
              />
            ) : (
              <div className="space-y-2">
                {reviews.map((review) => (
                  <Card key={review.id} size="sm">
                    <CardHeader>
                      <div className="flex items-center justify-between gap-2">
                        <span className="text-sm text-muted-foreground">
                          {new Date(review.created_at).toLocaleDateString(
                            "zh-CN"
                          )}
                        </span>
                        <Badge variant="secondary">
                          {review.overall_score.toFixed(1)} 分
                        </Badge>
                      </div>
                    </CardHeader>
                    <CardContent className="space-y-2">
                      {review.content && (
                        <p className="text-sm whitespace-pre-wrap">
                          {review.content}
                        </p>
                      )}
                      <p className="text-xs text-muted-foreground">
                        口味 {review.taste} · 分量 {review.portion} · 性价比{" "}
                        {review.value}
                      </p>
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
          </section>

          <section className="space-y-3">
            <h2 className="text-lg font-semibold">给这道菜打分</h2>
            <CanteenDishReviewForm dishID={dish.id} />
          </section>
        </div>
      </PageShell>
    </>
  )
}
