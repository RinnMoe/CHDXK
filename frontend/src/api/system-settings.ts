import { apiClient } from "./client"
import { BASE_URL } from "./constants"

export const SYSTEM_SETTING_CURRENT_SEMESTER = "current_semester"

export interface SystemSettingDTO {
  key: string
  value: string
}

export interface UpdateSystemSettingCommand {
  value: string
}

export function listSystemSettings(): Promise<SystemSettingDTO[]> {
  return apiClient(`${BASE_URL}/system-settings`)
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
