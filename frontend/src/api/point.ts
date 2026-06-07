import type { PaginatedResult } from "./types"
import { apiClient } from "./client"
import { BASE_URL } from "./constants"

export type RecordReason = string

export interface PointRecordDTO {
  reason: RecordReason
  amount: number
  description: string
  created_at: string
}

export interface PointSummaryDTO {
  total: number
  records: PaginatedResult<PointRecordDTO>
}

export interface PointRecordListFilter {
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

export function getUserPoints(
  userID: number,
  filter: PointRecordListFilter = {}
): Promise<PointSummaryDTO> {
  return apiClient(`${BASE_URL}/user/${userID}/point${buildQuery(filter)}`)
}
