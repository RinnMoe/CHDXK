import { apiClient } from "./client"
import { BASE_URL } from "./constants"

export interface AnnouncementDTO {
  id: number
  title: string
  body: string
  priority: number
  created_at: string
}

export function listAnnouncements(): Promise<AnnouncementDTO[]> {
  return apiClient(`${BASE_URL}/announcement/`)
}
