import { apiClient } from "./client"
import { BASE_URL } from "./constants"

export interface AdminUserDTO {
  id: number
  username: string
  email: string
  role: string
  created_at: string
  last_seen_at: string
  suspended: boolean
  suspended_at?: string
  suspend_till?: string
}

export function getAdminUserByEmail(email: string): Promise<AdminUserDTO> {
  const params = new URLSearchParams({ email })
  return apiClient(`${BASE_URL}/admin/user/by-email?${params.toString()}`)
}

export interface SuspendAdminUserCommand {
  days?: number
}

export function suspendAdminUser(
  userID: number,
  cmd: SuspendAdminUserCommand = {}
): Promise<{ message: string }> {
  return apiClient(`${BASE_URL}/admin/user/${userID}/suspension`, {
    method: "PUT",
    body: JSON.stringify(cmd),
  })
}

export function clearAdminUserSuspension(
  userID: number
): Promise<{ message: string }> {
  return apiClient(`${BASE_URL}/admin/user/${userID}/suspension`, {
    method: "DELETE",
  })
}
