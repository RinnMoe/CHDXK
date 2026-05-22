import { apiClient } from "./client"

const BASE = "/api"

export interface AnnouncementDTO {
  id: number
  title: string
  body: string
  priority: number
  created_at: string
}

export function listAnnouncements(): Promise<AnnouncementDTO[]> {
  return apiClient(`${BASE}/announcement/`)
}
