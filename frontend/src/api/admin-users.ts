import { apiClient } from "./client"
import { BASE_URL } from "./constants"
import type { AuthUserResponse } from "./auth"

export function getManagedUsers(q: string, role: string, page: number) {
  return apiClient<{ items: AuthUserResponse[]; total: number }>(
    `${BASE_URL}/admin/users?${new URLSearchParams({ q, role, page: String(page) })}`
  )
}
export function manageUser(
  id: number | undefined,
  action: string,
  body?: unknown
) {
  const method =
    action === "create"
      ? "POST"
      : action === "delete"
        ? "DELETE"
        : action === "profile" || action === "role"
          ? "PUT"
          : "POST"
  const suffix =
    id === undefined
      ? ""
      : `/${id}${action === "profile" || action === "delete" ? "" : `/${action}`}`
  return apiClient<{ balance?: number }>(`${BASE_URL}/admin/users${suffix}`, {
    method,
    body: body === undefined ? undefined : JSON.stringify(body),
  })
}
