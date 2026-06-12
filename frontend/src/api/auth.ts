import { apiClient } from "./client"
import { BASE_URL } from "./constants"

export interface AuthUserDTO {
  id: number
  username: string
  email: string
  role: string
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

export interface SendRegisterCodeCommand {
  email: string
}

export interface RegisterCommand {
  email: string
  code: string
  password: string
}

export interface LoginCommand {
  email: string
  password: string
}

export interface SendResetCodeCommand {
  email: string
}

export interface ResetPasswordCommand {
  email: string
  code: string
  new_password: string
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

export function sendRegisterCode(
  cmd: SendRegisterCodeCommand
): Promise<{ message: string }> {
  return apiClient(`${BASE_URL}/auth/register/code`, {
    method: "POST",
    body: JSON.stringify(cmd),
  })
}

export function register(cmd: RegisterCommand): Promise<AuthUserDTO> {
  return apiClient<AuthUserResponse>(`${BASE_URL}/auth/register`, {
    method: "POST",
    body: JSON.stringify(cmd),
  }).then(normalizeAuthUser)
}

export function login(cmd: LoginCommand): Promise<AuthUserDTO> {
  return apiClient<AuthUserResponse>(`${BASE_URL}/auth/login`, {
    method: "POST",
    body: JSON.stringify(cmd),
  }).then(normalizeAuthUser)
}

export function logout(): Promise<{ message: string }> {
  return apiClient(`${BASE_URL}/auth/logout`, {
    method: "POST",
  })
}

export function sendResetCode(
  cmd: SendResetCodeCommand
): Promise<{ message: string }> {
  return apiClient(`${BASE_URL}/auth/password-reset/code`, {
    method: "POST",
    body: JSON.stringify(cmd),
  })
}

export function resetPassword(
  cmd: ResetPasswordCommand
): Promise<{ message: string }> {
  return apiClient(`${BASE_URL}/auth/password-reset`, {
    method: "POST",
    body: JSON.stringify(cmd),
  })
}
