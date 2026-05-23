import { useState } from "react"
import { Navigate, useSearchParams } from "react-router-dom"
import { PageShell } from "@/components/layout/page-shell"
import { Skeleton } from "@/components/ui/skeleton"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { StatsCard } from "@/components/site-stats/stats-card"
import { DailyStatsChart } from "@/components/site-stats/daily-stats-chart"
import { DailyStatsTable } from "@/components/site-stats/daily-stats-table"
import { useDailyStats, useYesterdayStats } from "@/hooks/use-site-stats"
import { useAuth } from "@/contexts/auth-context"

const tablePageSize = 20
const chartPageSize = 10000

function formatDateInputValue(date: Date) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, "0")
  const day = String(date.getDate()).padStart(2, "0")
  return `${year}-${month}-${day}`
}

function getDefaultDateRange() {
  const endDate = new Date()
  const startDate = new Date(endDate)
  startDate.setDate(startDate.getDate() - 29)
  return {
    startDate: formatDateInputValue(startDate),
    endDate: formatDateInputValue(endDate),
  }
}

export function SiteStatsPage() {
  const { user, isLoading: authLoading } = useAuth()
  const [searchParams, setSearchParams] = useSearchParams()
  const defaultRange = getDefaultDateRange()
  const startDate = searchParams.get("start_date") || defaultRange.startDate
  const endDate = searchParams.get("end_date") || defaultRange.endDate
  const [page, setPage] = useState(1)
  const dateFilter = {
    start_date: startDate || undefined,
    end_date: endDate || undefined,
  }

  function updateDateRange(params: { start_date?: string; end_date?: string }) {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev)
        if (params.start_date !== undefined)
          next.set("start_date", params.start_date)
        if (params.end_date !== undefined) next.set("end_date", params.end_date)
        return next
      },
      { replace: true }
    )
    setPage(1)
  }

  function clearDateRange() {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev)
        next.delete("start_date")
        next.delete("end_date")
        return next
      },
      { replace: true }
    )
    setPage(1)
  }

  const { data: yesterday, isLoading: yLoading } = useYesterdayStats()
  const { data: daily, isLoading: dLoading } = useDailyStats({
    ...dateFilter,
    page,
    page_size: tablePageSize,
  })
  const { data: chartDaily, isLoading: chartLoading } = useDailyStats({
    ...dateFilter,
    page: 1,
    page_size: chartPageSize,
  })

  if (authLoading) return null
  if (!user) return <Navigate to="/login" replace />
  if (user.role !== "admin") {
    return (
      <>
        <title>站点统计 - JCourse</title>
        <PageShell>
          <p className="py-12 text-center text-muted-foreground">
            需要管理员权限
          </p>
        </PageShell>
      </>
    )
  }

  return (
    <>
      <title>站点统计 - JCourse</title>
      <PageShell>
        <div className="space-y-6">
          <h1 className="text-2xl font-semibold">站点统计</h1>

          {yLoading ? (
            <Skeleton className="h-40 w-full" />
          ) : yesterday ? (
            <StatsCard stat={yesterday} title="昨日数据" />
          ) : (
            <p className="text-sm text-muted-foreground">暂无昨日数据</p>
          )}

          <div className="flex flex-wrap items-end gap-2">
            <div className="space-y-1">
              <Label htmlFor="start-date" className="text-sm">
                开始日期
              </Label>
              <Input
                id="start-date"
                type="date"
                value={startDate}
                onChange={(e) =>
                  updateDateRange({ start_date: e.target.value })
                }
                className="w-44"
              />
            </div>
            <div className="space-y-1">
              <Label htmlFor="end-date" className="text-sm">
                结束日期
              </Label>
              <Input
                id="end-date"
                type="date"
                value={endDate}
                onChange={(e) => updateDateRange({ end_date: e.target.value })}
                className="w-44"
              />
            </div>
            {(searchParams.get("start_date") ||
              searchParams.get("end_date")) && (
              <Button variant="ghost" size="sm" onClick={clearDateRange}>
                重置
              </Button>
            )}
          </div>

          <Tabs defaultValue="chart">
            <TabsList>
              <TabsTrigger value="chart">图表</TabsTrigger>
              <TabsTrigger value="detail">明细</TabsTrigger>
            </TabsList>
            <TabsContent value="chart">
              {chartLoading ? (
                <Skeleton className="h-96 w-full" />
              ) : chartDaily && chartDaily.items.length > 0 ? (
                <DailyStatsChart stats={chartDaily.items} />
              ) : (
                <p className="text-sm text-muted-foreground">暂无统计数据</p>
              )}
            </TabsContent>
            <TabsContent value="detail" className="space-y-4">
              {dLoading ? (
                <Skeleton className="h-96 w-full" />
              ) : daily ? (
                <DailyStatsTable
                  stats={daily.items}
                  page={daily.page}
                  pageSize={daily.page_size}
                  total={daily.total}
                />
              ) : null}

              {daily && daily.total > tablePageSize && (
                <div className="flex justify-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPage((p) => Math.max(1, p - 1))}
                    disabled={page <= 1}
                  >
                    上一页
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPage((p) => p + 1)}
                    disabled={page * tablePageSize >= daily.total}
                  >
                    下一页
                  </Button>
                </div>
              )}
            </TabsContent>
          </Tabs>
        </div>
      </PageShell>
    </>
  )
}
