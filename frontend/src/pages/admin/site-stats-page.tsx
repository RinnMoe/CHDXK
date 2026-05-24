import { useState } from "react"
import dayjs from "dayjs"
import { zhCN } from "date-fns/locale"
import { RiCalendarLine } from "@remixicon/react"
import { Navigate, useSearchParams } from "react-router-dom"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { Skeleton } from "@/components/ui/skeleton"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import { Label } from "@/components/ui/label"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { StatsCard } from "@/components/site-stats/stats-card"
import { DailyStatsChart } from "@/components/site-stats/daily-stats-chart"
import { DailyStatsTable } from "@/components/site-stats/daily-stats-table"
import { useDailyStats, useYesterdayStats } from "@/hooks/use-site-stats"
import { useAuth } from "@/contexts/auth-context"
import { formatDateInputValue, formatRelativeDateInputValue } from "@/lib/date"

const tablePageSize = 20
const chartPageSize = 10000

type DateRangeParams = {
  start_date?: string
  end_date?: string
}

function getDefaultDateRange() {
  return {
    startDate: formatRelativeDateInputValue(-29),
    endDate: formatDateInputValue(new Date()),
  }
}

function getQuickDateRange(days: number): DateRangeParams {
  return {
    start_date: formatRelativeDateInputValue(-(days - 1)),
    end_date: formatDateInputValue(new Date()),
  }
}

function parseDateInputValue(value: string): Date | undefined {
  const date = dayjs(value)
  return date.isValid() ? date.toDate() : undefined
}

type DatePickerProps = {
  id: string
  label: string
  value: string
  onChange: (value: string) => void
}

function DatePicker({ id, label, value, onChange }: DatePickerProps) {
  const [open, setOpen] = useState(false)
  const selectedDate = parseDateInputValue(value)

  return (
    <div className="space-y-1">
      <Label htmlFor={id} className="text-sm">
        {label}
      </Label>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            id={id}
            type="button"
            variant="outline"
            className="w-44 justify-between font-normal"
          >
            {value}
            <RiCalendarLine className="size-4 text-muted-foreground" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="start">
          <Calendar
            mode="single"
            selected={selectedDate}
            defaultMonth={selectedDate}
            onSelect={(date) => {
              if (!date) return
              onChange(formatDateInputValue(date))
              setOpen(false)
            }}
            locale={zhCN}
            captionLayout="dropdown"
          />
        </PopoverContent>
      </Popover>
    </div>
  )
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

  function updateDateRange(params: DateRangeParams) {
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
        <PageTitle>站点统计</PageTitle>
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
      <PageTitle>站点统计</PageTitle>
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
            <DatePicker
              id="start-date"
              label="开始日期"
              value={startDate}
              onChange={(value) => updateDateRange({ start_date: value })}
            />
            <DatePicker
              id="end-date"
              label="结束日期"
              value={endDate}
              onChange={(value) => updateDateRange({ end_date: value })}
            />
            <div className="flex flex-wrap items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => updateDateRange(getQuickDateRange(30))}
              >
                最近30天
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => updateDateRange(getQuickDateRange(365))}
              >
                最近一年
              </Button>
              {(searchParams.get("start_date") ||
                searchParams.get("end_date")) && (
                <Button variant="ghost" size="sm" onClick={clearDateRange}>
                  重置
                </Button>
              )}
            </div>
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
