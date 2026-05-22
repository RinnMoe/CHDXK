import { useState } from "react"
import { Navigate } from "react-router-dom"
import { PageShell } from "@/components/layout/page-shell"
import { Skeleton } from "@/components/ui/skeleton"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { StatsCard } from "@/components/site-stats/stats-card"
import { DailyStatsTable } from "@/components/site-stats/daily-stats-table"
import { useDailyStats, useYesterdayStats } from "@/hooks/use-site-stats"
import { useAuth } from "@/contexts/auth-context"

export function SiteStatsPage() {
  const { user, isLoading: authLoading } = useAuth()
  const [startDate, setStartDate] = useState("")
  const [endDate, setEndDate] = useState("")
  const [page, setPage] = useState(1)
  const pageSize = 20

  const { data: yesterday, isLoading: yLoading } = useYesterdayStats()
  const { data: daily, isLoading: dLoading } = useDailyStats({
    start_date: startDate || undefined,
    end_date: endDate || undefined,
    page,
    page_size: pageSize,
  })

  if (authLoading) return null
  if (!user) return <Navigate to="/login" replace />
  if (user.role !== "admin") {
    return (
      <>
        <title>站点统计 - JCourse</title>
        <PageShell>
          <p className="text-center text-muted-foreground py-12">需要管理员权限</p>
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
            <Label htmlFor="start-date" className="text-xs">开始日期</Label>
            <Input
              id="start-date"
              type="date"
              value={startDate}
              onChange={(e) => { setStartDate(e.target.value); setPage(1) }}
              className="w-44"
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="end-date" className="text-xs">结束日期</Label>
            <Input
              id="end-date"
              type="date"
              value={endDate}
              onChange={(e) => { setEndDate(e.target.value); setPage(1) }}
              className="w-44"
            />
          </div>
          {(startDate || endDate) && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => { setStartDate(""); setEndDate(""); setPage(1) }}
            >
              清除
            </Button>
          )}
        </div>

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

        {daily && daily.total > pageSize && (
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
              disabled={page * pageSize >= daily.total}
            >
              下一页
            </Button>
          </div>
        )}
      </div>
    </PageShell>
    </>
  )
}
