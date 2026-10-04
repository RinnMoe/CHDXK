import { copy } from "@/lib/copy"
import { useState } from "react"
import { getRouteApi, useNavigate } from "@tanstack/react-router"
import dayjs from "dayjs"
import { zhCN } from "date-fns/locale"
import { RiCalendarLine } from "@remixicon/react"
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

type DateRangeParams = {
  start_date?: string
  end_date?: string
}
const routeApi = getRouteApi("/app/admin/site-stat")

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
  const { user } = useAuth()
  const search = routeApi.useSearch()
  const navigate = useNavigate({ from: "/admin/site-stat" })
  const defaultRange = getDefaultDateRange()
  const startDate = search.start_date || defaultRange.startDate
  const endDate = search.end_date || defaultRange.endDate
  const dateFilter = {
    start_date: startDate || undefined,
    end_date: endDate || undefined,
  }

  function updateDateRange(params: DateRangeParams) {
    void navigate({
      search: (prev) => ({
        ...prev,
        start_date: params.start_date ?? prev.start_date,
        end_date: params.end_date ?? prev.end_date,
      }),
      replace: true,
      resetScroll: false,
    })
  }

  function clearDateRange() {
    void navigate({
      search: (prev) => ({
        ...prev,
        start_date: undefined,
        end_date: undefined,
      }),
      replace: true,
      resetScroll: false,
    })
  }

  const { data: yesterday, isLoading: yLoading } = useYesterdayStats()
  const { data: daily, isLoading: dLoading } = useDailyStats(dateFilter)

  if (!user) return null
  if (!user.is_admin()) {
    return (
      <>
        <PageTitle>
          {copy("copy.admin.pages.admin.site-stats-page.c9cd3687fc73")}
        </PageTitle>
        <PageShell>
          <p className="py-12 text-center text-muted-foreground">
            {copy("copy.admin.pages.admin.site-stats-page.9228f9aa4c65")}
          </p>
        </PageShell>
      </>
    )
  }

  return (
    <>
      <PageTitle>
        {copy("copy.admin.pages.admin.site-stats-page.c9cd3687fc73")}
      </PageTitle>
      <PageShell>
        <div className="space-y-6">
          <h1 className="text-2xl font-semibold">
            {copy("copy.admin.pages.admin.site-stats-page.c9cd3687fc73")}
          </h1>

          {yLoading ? (
            <Skeleton className="h-40 w-full" />
          ) : yesterday ? (
            <StatsCard
              stat={yesterday}
              title={copy(
                "copy.admin.pages.admin.site-stats-page.883ababee8db"
              )}
            />
          ) : (
            <p className="text-sm text-muted-foreground">
              {copy("copy.admin.pages.admin.site-stats-page.addbad62c46d")}
            </p>
          )}

          <div className="flex flex-wrap items-end gap-2">
            <DatePicker
              id="start-date"
              label={copy(
                "copy.admin.pages.admin.site-stats-page.760506491eef"
              )}
              value={startDate}
              onChange={(value) => updateDateRange({ start_date: value })}
            />
            <DatePicker
              id="end-date"
              label={copy(
                "copy.admin.pages.admin.site-stats-page.895cd52fbbb6"
              )}
              value={endDate}
              onChange={(value) => updateDateRange({ end_date: value })}
            />
            <div className="flex flex-wrap items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => updateDateRange(getQuickDateRange(30))}
              >
                {copy("copy.admin.pages.admin.site-stats-page.178393de565a")}
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => updateDateRange(getQuickDateRange(365))}
              >
                {copy("copy.admin.pages.admin.site-stats-page.80c3d880f104")}
              </Button>
              {(search.start_date || search.end_date) && (
                <Button variant="ghost" size="sm" onClick={clearDateRange}>
                  {copy("copy.admin.pages.admin.site-stats-page.cb5d682bac3d")}
                </Button>
              )}
            </div>
          </div>

          <Tabs defaultValue="chart">
            <TabsList>
              <TabsTrigger value="chart">
                {copy("copy.admin.pages.admin.site-stats-page.8cb443ab8379")}
              </TabsTrigger>
              <TabsTrigger value="detail">
                {copy("copy.admin.pages.admin.site-stats-page.206ea896472e")}
              </TabsTrigger>
            </TabsList>
            <TabsContent value="chart">
              {dLoading ? (
                <Skeleton className="h-96 w-full" />
              ) : daily && daily.length > 0 ? (
                <DailyStatsChart stats={daily} />
              ) : (
                <p className="text-sm text-muted-foreground">
                  {copy("copy.admin.pages.admin.site-stats-page.1b8e57ace988")}
                </p>
              )}
            </TabsContent>
            <TabsContent value="detail" className="space-y-4">
              {dLoading ? (
                <Skeleton className="h-96 w-full" />
              ) : daily ? (
                <DailyStatsTable stats={daily} />
              ) : null}
            </TabsContent>
          </Tabs>
        </div>
      </PageShell>
    </>
  )
}
