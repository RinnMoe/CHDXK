import type { PaginatedResult } from "./types"
import { apiClient } from "./client"
import { BASE_URL } from "./constants"

export interface AuditLogDTO {
  id: string
  occurred_at: string
  actor_user_id: number
  action: string
  target_type: string
  target_id: string
  details: Record<string, unknown>
  created_at: string
}

export interface AuditLogListFilter {
  start_time?: string
  end_time?: string
  action?: string
  actor_user_id?: number
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

export function listAuditLogs(
  filter: AuditLogListFilter = {}
): Promise<PaginatedResult<AuditLogDTO>> {
  return apiClient(`${BASE_URL}/admin/audit-log${buildQuery(filter)}`)
}
