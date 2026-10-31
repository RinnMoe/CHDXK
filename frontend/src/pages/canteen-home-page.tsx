import { Link, getRouteApi, useNavigate } from "@tanstack/react-router"
import { useState } from "react"

import type { CanteenHallDTO } from "@/api/canteen"
import { CanteenEmptyState } from "@/components/canteen/canteen-empty-state"
import { CanteenHallFilter } from "@/components/canteen/canteen-hall-filter"
import { CanteenReviewCard } from "@/components/canteen/canteen-review-card"
import { CanteenStoreList } from "@/components/canteen/canteen-store-list"
import { PaginationComponent } from "@/components/common/pagination"
import { PageTitle } from "@/components/common/page-title"
import { PageShell } from "@/components/layout/page-shell"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import {
  useCanteenHome,
  useCanteenSearch,
  useCanteenStores,
} from "@/hooks/use-canteen-catalog"
import type { CanteenListSearch } from "@/router-search"

const routeApi = getRouteApi("/app/canteen")
const PAGE_SIZE = 20

const feedLabels: Record<string, string> = {
  recommended_store: "推荐",
  trending: "近期热门",
  new_store: "新店",
  stable_choice: "稳定选择",
  recent_photo: "同学实拍",
}

export function CanteenHomePage() {
  const search = routeApi.useSearch()
  const navigate = useNavigate({ from: "/canteen" })
  const home = useCanteenHome()
  const [query, setQuery] = useState("")
  const [submittedQuery, setSubmittedQuery] = useState("")
  const searchResult = useCanteenSearch(submittedQuery)
  const stores = useCanteenStores({
    hall_id: search.hall_id,
    campus: search.campus,
    q: search.q,
    sort: search.sort,
    page: search.page ?? 1,
    page_size: PAGE_SIZE,
  })

  function selectHall(hall: CanteenHallDTO | null) {
    void navigate({
      search: (prev: CanteenListSearch) => ({
        ...prev,
        hall_id: hall?.id,
        campus: hall?.campus_key,
        page: 1,
      }),
      resetScroll: true,
    })
  }

  function changePage(page: number) {
    void navigate({
      search: (prev: CanteenListSearch) => ({ ...prev, page }),
      resetScroll: true,
    })
  }

  function submitSearch(event: { preventDefault: () => void }) {
    event.preventDefault()
    setSubmittedQuery(query.trim())
  }

  if (home.isLoading) {
    return (
      <PageShell>
        <p className="py-16 text-center text-sm text-muted-foreground">
          正在加载食堂数据…
        </p>
      </PageShell>
    )
  }

  if (home.isError) {
    return (
      <PageShell>
        <Card size="sm">
          <CardContent className="space-y-2 py-10 text-center">
            <p className="text-sm font-medium">食堂评价暂不可用</p>
            <p className="text-xs text-muted-foreground">
              功能可能尚未开放，或后端服务未启动。
            </p>
          </CardContent>
        </Card>
      </PageShell>
    )
  }

  const data = home.data
  if (!data) return null

  return (
    <>
      <PageTitle>食堂评价</PageTitle>
      <PageShell>
        <div className="space-y-8">
          <header className="space-y-2">
            <h1 className="text-2xl font-bold">食堂评价</h1>
            <p className="text-sm text-muted-foreground">
              按校区和食堂浏览店铺，查看真实到店评价与综合评分。
            </p>
            <div className="flex flex-wrap gap-4 pt-2 text-sm text-muted-foreground">
              <span>{data.stats.store_count} 家店铺</span>
              <span>{data.stats.review_count} 条到店评价</span>
              <span>{data.stats.dish_count} 道菜品</span>
            </div>
          </header>

          <form onSubmit={submitSearch} className="flex max-w-xl gap-2">
            <Input
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder="搜索店铺或菜品，如：牛肉面"
              aria-label="搜索店铺或菜品"
            />
            <Button type="submit" size="sm">
              搜索
            </Button>
            {submittedQuery && (
              <Button
                type="button"
                size="sm"
                variant="ghost"
                onClick={() => {
                  setQuery("")
                  setSubmittedQuery("")
                }}
              >
                清除
              </Button>
            )}
          </form>

          {submittedQuery && (
            <section className="space-y-4">
              <h2 className="text-lg font-semibold">
                “{submittedQuery}”的搜索结果
              </h2>
              {searchResult.isLoading ? (
                <p className="text-sm text-muted-foreground">搜索中…</p>
              ) : (
                <>
                  {(searchResult.data?.stores.length ?? 0) === 0 &&
                    (searchResult.data?.dishes.length ?? 0) === 0 && (
                      <CanteenEmptyState
                        title="没有找到相关店铺或菜品"
                        description="换一个关键词，或在店铺页提交新菜品。"
                      />
                    )}
                  {(searchResult.data?.stores.length ?? 0) > 0 && (
                    <div className="space-y-2">
                      <h3 className="text-sm font-medium text-muted-foreground">
                        店铺
                      </h3>
                      <CanteenStoreList stores={searchResult.data?.stores ?? []} />
                    </div>
                  )}
                  {(searchResult.data?.dishes.length ?? 0) > 0 && (
                    <div className="space-y-2">
                      <h3 className="text-sm font-medium text-muted-foreground">
                        菜品
                      </h3>
                      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                        {(searchResult.data?.dishes ?? []).map((dish) => (
                          <Link
                            key={dish.id}
                            to="/canteen/stores/$storeID/dishes/$dishID"
                            params={{
                              storeID: String(dish.store_id),
                              dishID: String(dish.id),
                            }}
                          >
                            <Card size="sm" className="h-full hover:bg-muted/40">
                              <CardContent className="flex items-center gap-3 pt-6">
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
                                <div>
                                  <p className="font-medium">{dish.name}</p>
                                  <p className="text-xs text-muted-foreground">
                                    {dish.store_name} ·{" "}
                                    {dish.rating_score.toFixed(1)} 分
                                  </p>
                                </div>
                              </CardContent>
                            </Card>
                          </Link>
                        ))}
                      </div>
                    </div>
                  )}
                </>
              )}
            </section>
          )}

          {!submittedQuery && data.hero && (
            <section className="space-y-3">
              <h2 className="text-lg font-semibold">今天吃什么</h2>
              <Link
                to="/canteen/stores/$storeID"
                params={{ storeID: String(data.hero.store_id) }}
              >
                <Card className="overflow-hidden transition-colors hover:bg-muted/40">
                  <CardContent className="flex flex-wrap items-center gap-4 pt-6">
                    {data.hero.image_url ? (
                      <img
                        src={data.hero.image_url}
                        alt={data.hero.store_name}
                        className="h-28 w-40 rounded-md border object-cover"
                      />
                    ) : (
                      <div className="flex h-28 w-40 items-center justify-center rounded-md bg-muted text-xs text-muted-foreground">
                        暂无图片
                      </div>
                    )}
                    <div className="min-w-0 flex-1 space-y-1">
                      <p className="text-lg font-semibold">
                        {data.hero.dish_name || data.hero.store_name}
                      </p>
                      <p className="text-sm text-muted-foreground">
                        {data.hero.hall_name} · {data.hero.reason}
                      </p>
                      <p className="text-sm">
                        综合分 {data.hero.rating_score.toFixed(1)} ·{" "}
                        {data.hero.rating_count} 人评价
                      </p>
                    </div>
                  </CardContent>
                </Card>
              </Link>
            </section>
          )}

          {!submittedQuery && data.feeds.length > 0 && (
            <section className="space-y-3">
              <h2 className="text-lg font-semibold">发现</h2>
              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
                {data.feeds.map((feed) => (
                  <Link
                    key={`${feed.type}-${feed.store_id}`}
                    to="/canteen/stores/$storeID"
                    params={{ storeID: String(feed.store_id) }}
                  >
                    <Card size="sm" className="h-full hover:bg-muted/40">
                      <CardContent className="space-y-2 pt-6">
                        <Badge variant="outline">
                          {feedLabels[feed.type] ?? feed.type}
                        </Badge>
                        <p className="font-medium">
                          {feed.dish_name || feed.store_name}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          {feed.store_name} · {feed.reason}
                        </p>
                        {feed.tags.length > 0 && (
                          <div className="flex flex-wrap gap-1">
                            {feed.tags.map((tag) => (
                              <Badge key={tag} variant="secondary">
                                {tag}
                              </Badge>
                            ))}
                          </div>
                        )}
                      </CardContent>
                    </Card>
                  </Link>
                ))}
              </div>
            </section>
          )}

          {!submittedQuery && data.hot_dishes.length > 0 && (
            <section className="space-y-3">
              <h2 className="text-lg font-semibold">热门菜品</h2>
              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {data.hot_dishes.map((dish) => (
                  <Link
                    key={dish.id}
                    to="/canteen/stores/$storeID/dishes/$dishID"
                    params={{
                      storeID: String(dish.store_id),
                      dishID: String(dish.id),
                    }}
                  >
                    <Card size="sm" className="h-full hover:bg-muted/40">
                      <CardContent className="flex items-center gap-3 pt-6">
                        {dish.cover_url ? (
                          <img
                            src={dish.cover_url}
                            alt={dish.name}
                            className="size-16 rounded-md border object-cover"
                          />
                        ) : (
                          <div className="flex size-16 items-center justify-center rounded-md bg-muted text-xs text-muted-foreground">
                            暂无图
                          </div>
                        )}
                        <div>
                          <p className="font-medium">{dish.name}</p>
                          <p className="text-xs text-muted-foreground">
                            {dish.store_name} · {dish.rating_score.toFixed(1)} 分
                          </p>
                          {dish.photo_count > 0 && (
                            <p className="text-xs text-muted-foreground">
                              {dish.photo_count} 张实拍
                            </p>
                          )}
                        </div>
                      </CardContent>
                    </Card>
                  </Link>
                ))}
              </div>
            </section>
          )}

          <CanteenHallFilter
            halls={data.halls}
            hallID={search.hall_id}
            onSelect={selectHall}
          />

          {!search.hall_id && !search.q && (
            <section className="space-y-3">
              <h2 className="text-lg font-semibold">热门店铺</h2>
              <CanteenStoreList
                stores={data.top_stores}
                emptyTitle="还没有评分数据"
                emptyDescription="等第一位同学发布到店评价后，排行榜会在这里出现。"
              />
            </section>
          )}

          <section className="space-y-3">
            <h2 className="text-lg font-semibold">全部店铺</h2>
            {stores.isLoading ? (
              <p className="text-sm text-muted-foreground">加载中…</p>
            ) : (
              <CanteenStoreList
                stores={stores.data?.items ?? []}
                emptyTitle="没有找到店铺"
                emptyDescription="可以换一个食堂，或等待管理员补充店铺目录。"
              />
            )}
            {stores.data && (
              <PaginationComponent
                page={stores.data.page}
                pageSize={stores.data.page_size}
                total={stores.data.total}
                onPageChange={changePage}
              />
            )}
          </section>

          <section className="space-y-3">
            <h2 className="text-lg font-semibold">最新评价</h2>
            {data.latest_reviews.length === 0 ? (
              <CanteenEmptyState
                title="还没有到店评价"
                description="发布第一条评价，帮助同学了解这家店。"
              />
            ) : (
              <div className="grid gap-3 md:grid-cols-2">
                {data.latest_reviews.map((review) => (
                  <CanteenReviewCard key={review.id} review={review} showStore />
                ))}
              </div>
            )}
          </section>
        </div>
      </PageShell>
    </>
  )
}
