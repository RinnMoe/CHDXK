import { Link } from "@tanstack/react-router"

import { canteenMediaURL } from "@/api/canteen-community"
import { CanteenEmptyState } from "@/components/canteen/canteen-empty-state"
import { PageTitle } from "@/components/common/page-title"
import { PageShell } from "@/components/layout/page-shell"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { useMyCanteenContributions } from "@/hooks/use-canteen-dishes"

function statusLabel(status: string) {
  switch (status) {
    case "approved":
      return "已通过"
    case "pending":
      return "待审核"
    case "rejected":
      return "未通过"
    default:
      return status
  }
}

export function CanteenContributionsPage() {
  const contributions = useMyCanteenContributions()
  const dishes = contributions.data?.dishes ?? []
  const media = contributions.data?.media ?? []

  return (
    <>
      <PageTitle>我的餐饮贡献</PageTitle>
      <PageShell>
        <div className="space-y-8">
          <header className="space-y-2">
            <h1 className="text-2xl font-bold">我的餐饮贡献</h1>
            <p className="text-sm text-muted-foreground">
              菜品评分与实拍记录。图片通过审核后才会出现在店铺页。
            </p>
          </header>

          <section className="space-y-3">
            <h2 className="text-lg font-semibold">菜品评分</h2>
            {dishes.length === 0 ? (
              <CanteenEmptyState
                title="还没有菜品评分"
                description="在店铺详情页的菜品列表里给吃过的菜打分。"
              />
            ) : (
              <div className="space-y-2">
                {dishes.map((dish) => (
                  <Card key={dish.id} size="sm">
                    <CardContent className="space-y-1 pt-6 text-sm">
                      <div className="flex flex-wrap items-center justify-between gap-2">
                        <div>
                          <span className="font-medium">{dish.dish_name}</span>
                          <span className="ml-2 text-xs text-muted-foreground">
                            {dish.store_name} ·{" "}
                            {new Date(dish.created_at).toLocaleDateString(
                              "zh-CN"
                            )}
                          </span>
                        </div>
                        <Badge variant="secondary">
                          {dish.overall_score.toFixed(1)} 分
                        </Badge>
                      </div>
                      {dish.content && (
                        <p className="text-muted-foreground">{dish.content}</p>
                      )}
                      <Link
                        to="/canteen/stores/$storeID"
                        params={{ storeID: String(dish.store_id) }}
                        className="text-xs text-primary hover:underline"
                      >
                        查看店铺
                      </Link>
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
          </section>

          <section className="space-y-3">
            <h2 className="text-lg font-semibold">实拍图片</h2>
            {media.length === 0 ? (
              <CanteenEmptyState
                title="还没有上传图片"
                description="发布到店评价时可以附上最多 3 张实拍。"
              />
            ) : (
              <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4">
                {media.map((asset) => (
                  <Card key={asset.id} size="sm">
                    <CardHeader className="pb-2">
                      <Badge variant="outline">{statusLabel(asset.status)}</Badge>
                    </CardHeader>
                    <CardContent>
                      <img
                        src={canteenMediaURL(asset.id)}
                        alt="实拍"
                        className="h-32 w-full rounded-md object-cover"
                      />
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
          </section>
        </div>
      </PageShell>
    </>
  )
}
