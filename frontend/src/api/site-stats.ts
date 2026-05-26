import { apiClient } from "./client"
import { BASE_URL } from "./constants"

export interface SiteDailyStatDTO {
  stat_date: string
  total_user_count: number
  total_review_count: number
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

export function getDailyStat(date: string): Promise<SiteDailyStatDTO> {
  return apiClient(`${BASE_URL}/site-stat/daily/${date}`)
}

export function listDailyStats(
  filter: SiteDailyStatListFilter = {}
): Promise<SiteDailyStatDTO[]> {
  return apiClient(`${BASE_URL}/site-stat/daily${buildQuery(filter)}`)
}
