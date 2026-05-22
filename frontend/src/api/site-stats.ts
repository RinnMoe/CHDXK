import type { PaginatedResult } from "./types"
import { apiClient } from "./client"

const BASE = "/api"

export interface SiteDailyStatDTO {
  stat_date: string
  active_user_count: number
  new_user_count: number
  new_review_count: number
  review_author_count: number
  reviewed_course_total: number
  new_like_count: number
  new_dislike_count: number
  generated_at: string
  updated_at: string
}

export interface SiteDailyStatListFilter {
  start_date?: string
  end_date?: string
  page?: number
  page_size?: number
  [key: string]: unknown
}

function buildQuery(filter: Record<string, unknown>): string {
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(filter)) {
    if (value === undefined || value === null || value === "") continue
    params.append(key, String(value))
  }
  const q = params.toString()
  return q ? `?${q}` : ""
}

export function getYesterdayStats(): Promise<SiteDailyStatDTO> {
  return apiClient(`${BASE}/site-stats/daily/yesterday`)
}

export function listDailyStats(
  filter: SiteDailyStatListFilter = {}
): Promise<PaginatedResult<SiteDailyStatDTO>> {
  return apiClient(`${BASE}/site-stats/daily${buildQuery(filter)}`)
}
