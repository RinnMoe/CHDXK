import { Link } from "@tanstack/react-router"
import { useState } from "react"

import { CanteenEmptyState } from "@/components/canteen/canteen-empty-state"
import { PaginationComponent } from "@/components/common/pagination"
import { PageTitle } from "@/components/common/page-title"
import { PageShell } from "@/components/layout/page-shell"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import {
  useDeleteCanteenReview,
  useMyCanteenReviews,
} from "@/hooks/use-canteen-reviews"

const PAGE_SIZE = 20

export function CanteenMyReviewsPage() {
  const [page, setPage] = useState(1)
  const mine = useMyCanteenReviews(page, PAGE_SIZE)
  const remove = useDeleteCanteenReview()
  const [error, setError] = useState<string | null>(null)

  async function handleDelete(reviewID: number) {
    if (!window.confirm("确定删除这条到店评价吗？")) return
    setError(null)
    try {
      await remove.mutateAsync(reviewID)
    } catch (err) {
      setError(err instanceof Error ? err.message : "删除评价失败")
    }
  }

  return (
    <>
      <PageTitle>我的餐饮评价</PageTitle>
      <PageShell>
        <div className="space-y-4">
          <header className="space-y-2">
            <div className="flex flex-wrap items-center gap-3">
              <h1 className="text-2xl font-bold">我的餐饮评价</h1>
              <Button asChild type="button" size="xs" variant="outline">
                <Link to="/canteen/contributions">我的餐饮贡献</Link>
              </Button>
            </div>
            <p className="text-sm text-muted-foreground">
              每次到店记录都会独立保存，修改会同步更新店铺综合分。
            </p>
          </header>

          {error && <p className="text-sm text-destructive">{error}</p>}

          {mine.isLoading ? (
            <p className="text-sm text-muted-foreground">加载中…</p>
          ) : (mine.data?.items ?? []).length === 0 ? (
            <CanteenEmptyState
              title="还没有发布过餐饮评价"
              description="去食堂首页选择一家店铺，记录第一次到店体验。"
            />
          ) : (
            <div className="space-y-3">
              {(mine.data?.items ?? []).map((review) => (
                <Card key={review.id} size="sm">
                  <CardContent className="space-y-2 pt-6">
                    <div className="flex flex-wrap items-start justify-between gap-3">
                      <div>
                        <Link
                          to="/canteen/stores/$storeID"
                          params={{ storeID: String(review.store_id) }}
                          className="font-medium hover:underline"
                        >
                          {review.store_name}
                        </Link>
                        <p className="text-xs text-muted-foreground">
                          {new Date(review.visited_at).toLocaleDateString(
                            "zh-CN"
                          )}{" "}
                          到店 · 综合分 {review.overall_score.toFixed(1)}
                        </p>
                      </div>
                      <div className="flex items-center gap-2">
                        <Badge
                          variant={
                            review.status === "active" ? "secondary" : "outline"
                          }
                        >
                          {review.status === "active"
                            ? "正常"
                            : review.status === "hidden"
                              ? "已隐藏"
                              : "已删除"}
                        </Badge>
                        <Button asChild type="button" size="xs" variant="outline">
                          <Link
                            to="/canteen/reviews/$reviewID/edit"
                            params={{ reviewID: String(review.id) }}
                          >
                            编辑
                          </Link>
                        </Button>
                        <Button
                          type="button"
                          size="xs"
                          variant="destructive"
                          onClick={() => void handleDelete(review.id)}
                        >
                          删除
                        </Button>
                      </div>
                    </div>
                    {review.content && (
                      <p className="text-sm whitespace-pre-wrap">
                        {review.content}
                      </p>
                    )}
                  </CardContent>
                </Card>
              ))}
            </div>
          )}

          {mine.data && (
            <PaginationComponent
              page={mine.data.page}
              pageSize={mine.data.page_size}
              total={mine.data.total}
              onPageChange={setPage}
            />
          )}
        </div>
      </PageShell>
    </>
  )
}
