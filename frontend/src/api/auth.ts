import { apiClient } from "./client"
import { BASE_URL } from "./constants"

export interface AuthUserDTO {
  nickname?: string
  local_role?: string
  permissions?: string[]
  must_change_password?: boolean
  has_password?: boolean
  suspended_at?: string | null
  suspend_till?: string | null
  muted_until?: string | null
  moderation_reason?: string
  can: (permission: string) => boolean
  is_bootstrap?: boolean
  bound_nickname?: string
  id: number
  username: string
  email: string
  role: string
  subject?: string
  identity_role?: string
  identity_type?: string
  permission_mode?: string
  violation_count?: number
  initial_password?: boolean
  has_points_account?: boolean
  session_id?: string
  expires_at?: string
  impersonated_by?: string
  is_admin: () => boolean
  is_super_admin: () => boolean
}

export type AuthUserResponse = Omit<
  AuthUserDTO,
  "is_admin" | "is_super_admin" | "can"
>

export function normalizeAuthUser(user: AuthUserResponse): AuthUserDTO {
  return {
    ...user,
    can: (permission) =>
      user.role === "super_admin" ||
      (user.role === "admin" && !!user.permissions?.includes(permission)),
    is_admin: () => user.role === "admin" || user.role === "super_admin",
    is_super_admin: () => user.role === "super_admin",
  }
}

/** Build the browser navigation URL for the xk-side OIDC authorization entry point. */
export function getOIDCStartURL(returnPath?: string) {
  const params = new URLSearchParams()
  if (returnPath) params.set("return_to", returnPath)
  const query = params.toString()
  return `${BASE_URL}/auth/oidc/start${query ? `?${query}` : ""}`
}

export function getCurrentUser(): Promise<AuthUserDTO | null> {
  return apiClient<AuthUserResponse | null>(`${BASE_URL}/auth/me`)
    .then((user) => (user ? normalizeAuthUser(user) : null))
    .catch((err) => {
      if (
        err &&
        typeof err === "object" &&
        "status" in err &&
        err.status === 401
      ) {
        return null
      }
      throw err
    })
}

export function logout(): Promise<{ message: string }> {
  return apiClient(`${BASE_URL}/auth/logout`, {
    method: "POST",
  })
}

export function getLoginAccounts() {
  return apiClient<AuthUserResponse[]>(`${BASE_URL}/auth/accounts`)
}
export function selectLoginAccount(id: number) {
  return apiClient<AuthUserResponse>(`${BASE_URL}/auth/accounts/select`, {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}
export function getAccountBindingURL() {
  return `${BASE_URL}/auth/oidc/start?purpose=bind&return_to=/settings`
}

export function getAuthOptions() {
  return apiClient<{ registration_enabled: boolean }>(
    `${BASE_URL}/auth/options`
  )
}
export function passwordLogin(username: string, password: string) {
  return apiClient<AuthUserResponse>(`${BASE_URL}/auth/login`, {
    method: "POST",
    body: JSON.stringify({ username, password }),
  })
}
export function registerAccount(
  username: string,
  nickname: string,
  password: string
) {
  return apiClient<AuthUserResponse>(`${BASE_URL}/auth/register`, {
    method: "POST",
    body: JSON.stringify({ username, nickname, password }),
  })
}
export function changePassword(old_password: string, password: string) {
  return apiClient(`${BASE_URL}/auth/password`, {
    method: "POST",
    body: JSON.stringify({ old_password, password }),
  })
}
export function updateAccountProfile(username: string, nickname: string) {
  return apiClient(`${BASE_URL}/auth/profile`, {
    method: "PUT",
    body: JSON.stringify({ username, nickname }),
  })
}
export function unbindAccount() {
  return apiClient(`${BASE_URL}/auth/binding`, { method: "DELETE" })
}
