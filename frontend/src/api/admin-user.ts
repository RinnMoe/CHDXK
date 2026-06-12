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
  password_hash?: string
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

export interface AdminUserLookup {
  email?: string
  username?: string
  review_id?: number
}

export function getAdminUser(lookup: AdminUserLookup): Promise<AdminUserDTO> {
  const params = new URLSearchParams()
  if (lookup.email) params.set("email", lookup.email)
  if (lookup.username) params.set("username", lookup.username)
  if (lookup.review_id) params.set("review_id", String(lookup.review_id))

  return apiClient<AdminUserResponse>(
    `${BASE_URL}/admin/user?${params.toString()}`
  ).then(normalizeAdminUser)
}

export interface SuspendAdminUserCommand {
  days: number
}

export function suspendAdminUser(
  userID: number,
  cmd: SuspendAdminUserCommand
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

export interface ResetAdminUserPasswordCommand {
  password: string
}

export function resetAdminUserPassword(
  userID: number,
  cmd: ResetAdminUserPasswordCommand
): Promise<{ message: string }> {
  return apiClient(`${BASE_URL}/admin/user/${userID}/password`, {
    method: "PUT",
    body: JSON.stringify(cmd),
  })
}
