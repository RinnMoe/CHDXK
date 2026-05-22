import { keepPreviousData, useQuery } from "@tanstack/react-query"
import { getYesterdayStats, listDailyStats, type SiteDailyStatListFilter } from "@/api/site-stats"

export function useYesterdayStats() {
  return useQuery({
    queryKey: ["site-stats", "yesterday"],
    queryFn: getYesterdayStats,
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
