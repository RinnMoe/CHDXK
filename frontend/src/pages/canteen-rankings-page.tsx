import { getRouteApi, useNavigate } from "@tanstack/react-router"
import { useState } from "react"

import type { CanteenHallDTO } from "@/api/canteen"
import { CanteenHallFilter } from "@/components/canteen/canteen-hall-filter"
import { CanteenStoreList } from "@/components/canteen/canteen-store-list"
import { PaginationComponent } from "@/components/common/pagination"
import { PageTitle } from "@/components/common/page-title"
import { PageShell } from "@/components/layout/page-shell"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { useCanteenHalls, useCanteenRankings } from "@/hooks/use-canteen-catalog"
import type { CanteenListSearch } from "@/router-search"

const routeApi = getRouteApi("/app/canteen/rankings")

const sortOptions = [
  { value: "rating_score", label: "综合分" },
  { value: "rating_count", label: "评价人数" },
  { value: "latest", label: "最新收录" },
] as const

export function CanteenRankingsPage() {
  const search = routeApi.useSearch()
  const navigate = useNavigate({ from: "/canteen/rankings" })
  const halls = useCanteenHalls()
  const [showInfo, setShowInfo] = useState(false)
  const sort = search.sort ?? "rating_score"
  const rankings = useCanteenRankings({
    campus: search.campus,
    hall_id: search.hall_id,
    sort,
    page: search.page ?? 1,
    page_size: 20,
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

  function changeSort(next: (typeof sortOptions)[number]["value"]) {
    void navigate({
      search: (prev: CanteenListSearch) => ({ ...prev, sort: next, page: 1 }),
      resetScroll: true,
    })
  }

  function changePage(page: number) {
    void navigate({
      search: (prev: CanteenListSearch) => ({ ...prev, page }),
      resetScroll: true,
    })
  }

  return (
    <>
      <PageTitle>食堂排行榜</PageTitle>
      <PageShell>
        <div className="space-y-6">
          <header className="space-y-2">
            <h1 className="text-2xl font-bold">食堂排行榜</h1>
            <p className="text-sm text-muted-foreground">
              按每位同学最近三次到店体验加权后，再做贝叶斯平滑排名。
            </p>
          </header>

          <CanteenHallFilter
            halls={halls.data ?? []}
            hallID={search.hall_id}
            onSelect={selectHall}
          />

          <div className="flex flex-wrap items-center gap-2">
            {sortOptions.map((option) => (
              <Button
                key={option.value}
                type="button"
                size="sm"
                variant={sort === option.value ? "default" : "outline"}
                onClick={() => changeSort(option.value)}
              >
                {option.label}
              </Button>
            ))}
            <Button
              type="button"
              size="sm"
              variant="ghost"
              className="ml-auto"
              onClick={() => setShowInfo(true)}
            >
              综合分怎么算？
            </Button>
          </div>

          <CanteenStoreList
            stores={rankings.data?.items ?? []}
            emptyTitle="排行榜还是空的"
            emptyDescription="有店铺获得有效评分后会自动进入榜单。"
          />

          {rankings.data && (
            <PaginationComponent
              page={rankings.data.page}
              pageSize={rankings.data.page_size}
              total={rankings.data.total}
              onPageChange={changePage}
            />
          )}
        </div>
      </PageShell>
      <Dialog open={showInfo} onOpenChange={setShowInfo}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>综合分怎么算？</DialogTitle>
          </DialogHeader>
          <div className="space-y-3 text-sm text-muted-foreground">
            <p>
              综合分采用贝叶斯平滑：既看平均分，也看评价人数。评价人数少的高分会向全平台均值收缩，避免“1 人 5 分”霸榜。
            </p>
            <p>
              用户摘要按最近三次到店加权（0.6/0.3/0.1），店铺分数再按去重评价人数平滑。样本少于 3 人标为「样本较少」，3–9 人标为「参考」，10 人以上标为「较可信」。
            </p>
            <p>标签来自近 30 天评价的高频体验标签；菜品和实拍数量只作参考，不参与综合分计算。</p>
          </div>
        </DialogContent>
      </Dialog>
    </>
  )
}
