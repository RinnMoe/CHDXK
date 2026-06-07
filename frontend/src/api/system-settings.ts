import { apiClient } from "./client"
import { BASE_URL } from "./constants"

export const SYSTEM_SETTING_CURRENT_SEMESTER = "current_semester"
export const SYSTEM_SETTING_AUTH_EMAIL_DOMAIN = "auth.email_domain"

export type SystemSettingType =
  | "string"
  | "int"
  | "bool"
  | "float"
  | "duration"
  | "string_list"

export interface SystemSettingDTO {
  key: string
  value: string
  default_value: string
  type: SystemSettingType
  group: string
  label: string
  description: string
  public: boolean
  secret: boolean
  requires_restart: boolean
}

export interface UpdateSystemSettingCommand {
  value: string
}

export function listSystemSettings(): Promise<SystemSettingDTO[]> {
  return apiClient(`${BASE_URL}/system-settings`)
}

export function listAdminSystemSettings(): Promise<SystemSettingDTO[]> {
  return apiClient(`${BASE_URL}/admin/system-settings`)
}

export function updateSystemSetting(
  key: string,
  cmd: UpdateSystemSettingCommand
): Promise<SystemSettingDTO> {
  return apiClient(
    `${BASE_URL}/admin/system-settings/${encodeURIComponent(key)}`,
    {
      method: "PUT",
      body: JSON.stringify(cmd),
    }
  )
}
