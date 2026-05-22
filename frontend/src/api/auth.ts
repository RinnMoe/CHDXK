import { apiClient } from "./client"

const BASE = "/api"

export interface AuthUserDTO {
  id: number
  username: string
  email: string
  role: string
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
  return apiClient<AuthUserDTO | null>(`${BASE}/auth/me`).catch((err) => {
    if (err && typeof err === "object" && "status" in err && err.status === 401) {
      return null
    }
    throw err
  })
}

export function sendRegisterCode(cmd: SendRegisterCodeCommand): Promise<{ message: string }> {
  return apiClient(`${BASE}/auth/register/code`, {
    method: "POST",
    body: JSON.stringify(cmd),
  })
}

export function register(cmd: RegisterCommand): Promise<AuthUserDTO> {
  return apiClient(`${BASE}/auth/register`, {
    method: "POST",
    body: JSON.stringify(cmd),
  })
}

export function login(cmd: LoginCommand): Promise<AuthUserDTO> {
  return apiClient(`${BASE}/auth/login`, {
    method: "POST",
    body: JSON.stringify(cmd),
  })
}

export function logout(): Promise<{ message: string }> {
  return apiClient(`${BASE}/auth/logout`, {
    method: "POST",
  })
}

export function sendResetCode(cmd: SendResetCodeCommand): Promise<{ message: string }> {
  return apiClient(`${BASE}/auth/password-reset/code`, {
    method: "POST",
    body: JSON.stringify(cmd),
  })
}

export function resetPassword(cmd: ResetPasswordCommand): Promise<{ message: string }> {
  return apiClient(`${BASE}/auth/password-reset`, {
    method: "POST",
    body: JSON.stringify(cmd),
  })
}
