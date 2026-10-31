import { Link, getRouteApi } from "@tanstack/react-router"
import { useState } from "react"

import { CanteenDimensionBars } from "@/components/canteen/canteen-dimension-bars"
import { CanteenDishList } from "@/components/canteen/canteen-dish-list"
import { CanteenReportDialog } from "@/components/canteen/canteen-report-dialog"
import { CanteenReviewList } from "@/components/canteen/canteen-review-list"
import { PageTitle } from "@/components/common/page-title"
import { PageShell } from "@/components/layout/page-shell"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { useAuth } from "@/contexts/auth-context"
import { useCanteenStore } from "@/hooks/use-canteen-catalog"
import {
  useCanteenStoreReviews,
  useVoteCanteenReview,
} from "@/hooks/use-canteen-reviews"

const routeApi = getRouteApi("/app/canteen/stores/$storeID")
const PAGE_SIZE = 20

const orderOptions = [
  { value: "latest", label: "最新" },
  { value: "best", label: "好评" },
  { value: "worst", label: "差评" },
] as const

const filterOptions = [
  { value: "all", label: "全部" },
  { value: "with_media", label: "有图" },
  { value: "high", label: "高分" },
  { value: "low", label: "低分" },
] as const

export function CanteenStoreDetailPage() {
  const { storeID } = routeApi.useParams()
  const storeIDNumber = Number(storeID)
  const store = useCanteenStore(storeIDNumber)
  const { user } = useAuth()
  const [order, setOrder] = useState<"latest" | "best" | "worst">("latest")
  const [filter, setFilter] = useState<"all" | "with_media" | "high" | "low">(
    "all"
  )
  const [history, setHistory] = useState(false)
  const [voteError, setVoteError] = useState<string | null>(null)
  const [reportTarget, setReportTarget] = useState<
    { type: "store" | "review"; id: number } | null
  >(null)
  const reviews = useCanteenStoreReviews(storeIDNumber, {
    order,
    filter,
    history,
    limit: PAGE_SIZE,
  })
  const reviewItems = reviews.data?.pages.flatMap((page) => page.items) ?? []
  const vote = useVoteCanteenReview()

  if (store.isLoading) {
    return (
      <PageShell>
        <p className="py-16 text-center text-sm text-muted-foreground">
          正在加载店铺…
        </p>
      </PageShell>
    )
  }

  if (store.isError || !store.data) {
    return (
      <PageShell>
        <div className="space-y-4 py-10 text-center">
          <p className="text-sm font-medium">店铺不存在或暂不可用</p>
          <Link to="/canteen" className="text-sm text-primary hover:underline">
            返回食堂首页
          </Link>
        </div>
      </PageShell>
    )
  }

  const data = store.data

  async function handleVote(reviewID: number, voteType: number) {
    setVoteError(null)
    try {
      await vote.mutateAsync({ reviewID, voteType })
    } catch (err) {
      setVoteError(err instanceof Error ? err.message : "投票失败")
    }
  }

  return (
    <>
      <PageTitle>{data.name}</PageTitle>
      <PageShell>
        <div className="space-y-6">
          <Link
            to="/canteen"
            className="text-sm text-muted-foreground hover:underline"
          >
            ← 返回食堂首页
          </Link>

          <header className="space-y-3">
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div className="space-y-1">
                <h1 className="text-2xl font-bold">{data.name}</h1>
                <p className="text-sm text-muted-foreground">
                  {data.hall_name}
                  {data.floor ? ` · ${data.floor}` : ""}
                </p>
              </div>
              <div className="text-right">
                <div className="text-3xl font-bold">
                  {data.rating.score.toFixed(1)}
                </div>
                <div className="text-xs text-muted-foreground">
                  {data.rating.count} 人评价 · {data.review_count} 次到店
                </div>
              </div>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              {data.categories.map((category) => (
                <Badge key={category} variant="secondary">
                  {category}
                </Badge>
              ))}
              {data.status === "active" && (
                <Button asChild size="sm" className="ml-auto">
                  <Link
                    to="/canteen/stores/$storeID/review/new"
                    params={{ storeID }}
                  >
                    写评价
                  </Link>
                </Button>
              )}
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={() =>
                  setReportTarget({ type: "store", id: storeIDNumber })
                }
              >
                举报店铺
              </Button>
            </div>
          </header>

          <div className="grid gap-6 lg:grid-cols-[2fr_1fr]">
            <Card size="sm">
              <CardHeader>
                <h2 className="text-base font-semibold">五维评分</h2>
              </CardHeader>
              <CardContent>
                <CanteenDimensionBars dimensions={data.dimensions} />
              </CardContent>
            </Card>

            <Card size="sm">
              <CardHeader>
                <h2 className="text-base font-semibold">我的记录</h2>
              </CardHeader>
              <CardContent className="space-y-2 text-sm">
                {data.my_summary ? (
                  <>
                    <p>
                      我的综合分：
                      <span className="font-semibold">
                        {data.my_summary.weighted_score.toFixed(2)}
                      </span>
                    </p>
                    <p className="text-muted-foreground">
                      最近 {data.my_summary.window_size} 次到店参与计算，累计{" "}
                      {data.my_summary.lifetime_event_count} 次
                    </p>
                    {data.my_latest_review && (
                      <p className="text-xs text-muted-foreground">
                        最近一次：
                        {new Date(
                          data.my_latest_review.visited_at
                        ).toLocaleDateString("zh-CN")}
                      </p>
                    )}
                  </>
                ) : (
                  <p className="text-muted-foreground">还没有到店记录。</p>
                )}
              </CardContent>
            </Card>
          </div>

          <section className="space-y-3">
            <h2 className="text-lg font-semibold">菜品</h2>
            <CanteenDishList
              storeID={storeIDNumber}
              canContribute={data.status === "active"}
            />
          </section>

          <section className="space-y-3">
            <div className="flex flex-wrap items-center gap-2">
              <h2 className="text-lg font-semibold">到店评价</h2>
              <span className="text-xs text-muted-foreground">
                {history ? "全部到店记录" : "每人最新一条"}
                {reviews.data?.pages[0]
                  ? ` · 共 ${reviews.data.pages[0].total} 条`
                  : ""}
              </span>
              <div className="ml-auto flex flex-wrap gap-2">
                <Button
                  type="button"
                  size="xs"
                  variant={history ? "default" : "outline"}
                  onClick={() => setHistory((value) => !value)}
                >
                  {history ? "收起历史" : "查看历史"}
                </Button>
                {orderOptions.map((option) => (
                  <Button
                    key={option.value}
                    type="button"
                    size="xs"
                    variant={order === option.value ? "default" : "outline"}
                    onClick={() => setOrder(option.value)}
                  >
                    {option.label}
                  </Button>
                ))}
              </div>
            </div>
            <div className="flex flex-wrap gap-2">
              {filterOptions.map((option) => (
                <Button
                  key={option.value}
                  type="button"
                  size="xs"
                  variant={filter === option.value ? "secondary" : "ghost"}
                  onClick={() => setFilter(option.value)}
                >
                  {option.label}
                </Button>
              ))}
            </div>
            {voteError && <p className="text-sm text-destructive">{voteError}</p>}
            {reviews.isLoading ? (
              <p className="text-sm text-muted-foreground">加载中…</p>
            ) : (
              <CanteenReviewList
                reviews={reviewItems}
                currentUserID={user?.id}
                onVote={(reviewID, voteType) =>
                  void handleVote(reviewID, voteType)
                }
                onReport={(reviewID) =>
                  setReportTarget({ type: "review", id: reviewID })
                }
                voting={vote.isPending}
              />
            )}
            {reviews.hasNextPage && (
              <Button
                type="button"
                size="sm"
                variant="outline"
                disabled={reviews.isFetchingNextPage}
                onClick={() => void reviews.fetchNextPage()}
              >
                {reviews.isFetchingNextPage ? "加载中…" : "加载更多评价"}
              </Button>
            )}
          </section>
        </div>
      </PageShell>
      <CanteenReportDialog
        open={reportTarget !== null}
        onOpenChange={(open) => {
          if (!open) setReportTarget(null)
        }}
        subjectType={reportTarget?.type ?? "review"}
        subjectID={reportTarget?.id ?? 0}
      />
    </>
  )
}
