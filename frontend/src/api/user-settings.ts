import { apiClient } from "./client"
import { BASE_URL } from "./constants"

export interface UserSettingsDTO {
  current_semester: string
}

export interface UpdateUserSettingsCommand {
  current_semester: string
}

export function getUserSettings(): Promise<UserSettingsDTO> {
  return apiClient(`${BASE_URL}/user/settings`)
}

export function updateUserSettings(
  cmd: UpdateUserSettingsCommand
): Promise<UserSettingsDTO> {
  return apiClient(`${BASE_URL}/user/settings`, {
    method: "PUT",
    body: JSON.stringify(cmd),
  })
}
