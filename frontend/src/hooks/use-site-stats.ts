import { keepPreviousData, useQuery } from "@tanstack/react-query"
import {
  getDailyStat,
  listDailyStats,
  type SiteDailyStatListFilter,
} from "@/api/site-stats"

function yesterdayDateStr(): string {
  const d = new Date()
  d.setDate(d.getDate() - 1)
  return d.toISOString().split("T")[0]
}

export function useYesterdayStats() {
  return useQuery({
    queryKey: ["site-stats", "daily", "yesterday"],
    queryFn: () => getDailyStat(yesterdayDateStr()),
    retry: false,
  })
}

export function useDailyStats(filter: SiteDailyStatListFilter = {}) {
  return useQuery({
    queryKey: ["site-stats", "daily", filter],
    queryFn: () => listDailyStats(filter),
    placeholderData: keepPreviousData,
  })
}
