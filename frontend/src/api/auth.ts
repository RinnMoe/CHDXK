import { apiClient } from "./client"
import { BASE_URL } from "./constants"

export interface AuthUserDTO {
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

export type AuthUserResponse = Omit<AuthUserDTO, "is_admin" | "is_super_admin">

export function normalizeAuthUser(user: AuthUserResponse): AuthUserDTO {
  return {
    ...user,
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
