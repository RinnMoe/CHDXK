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
  is_admin: () => boolean
  is_super_admin: () => boolean
}

type AdminUserResponse = Omit<AdminUserDTO, "is_admin" | "is_super_admin">

function normalizeAdminUser(user: AdminUserResponse): AdminUserDTO {
  return {
    ...user,
    is_admin: () => user.role === "admin" || user.role === "super_admin",
    is_super_admin: () => user.role === "super_admin",
  }
}

export function listAdminUsers(): Promise<AdminUserDTO[]> {
  return apiClient<AdminUserResponse[]>(`${BASE_URL}/admin/user/admin`).then(
    (users) => users.map(normalizeAdminUser)
  )
}

export function getAdminUserByEmail(email: string): Promise<AdminUserDTO> {
  const params = new URLSearchParams({ email })
  return apiClient<AdminUserResponse>(
    `${BASE_URL}/admin/user/by-email?${params.toString()}`
  ).then(normalizeAdminUser)
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

export function grantAdminUser(userID: number): Promise<{ message: string }> {
  return apiClient(`${BASE_URL}/admin/user/${userID}/admin`, {
    method: "PUT",
  })
}

export function revokeAdminUser(userID: number): Promise<{ message: string }> {
  return apiClient(`${BASE_URL}/admin/user/${userID}/admin`, {
    method: "DELETE",
  })
}
