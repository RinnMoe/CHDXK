import { keepPreviousData, useQuery } from "@tanstack/react-query"
import {
  getDailyStat,
  listDailyStats,
  type SiteDailyStatListFilter,
} from "@/api/site-stats"
import { formatRelativeDateInputValue } from "@/lib/date"

export function useYesterdayStats() {
  return useQuery({
    queryKey: ["site-stats", "daily", "yesterday"],
    queryFn: () => getDailyStat(formatRelativeDateInputValue(-1)),
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
