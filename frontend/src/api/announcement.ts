import { apiClient } from "./client"
import { BASE_URL } from "./constants"

export interface AnnouncementDTO {
  id: number
  title: string
  body: string
  priority: number
  show_start: string
  show_end: string
  link_url: string
  link_title: string
  created_at: string
}

export interface SaveAnnouncementCommand {
  title: string
  body: string
  priority: number
  show_start: string
  show_end: string
  link_url?: string
  link_title?: string
}

export function listAnnouncements(): Promise<AnnouncementDTO[]> {
  return apiClient(`${BASE_URL}/announcement/`)
}

export function listAdminAnnouncements(): Promise<AnnouncementDTO[]> {
  return apiClient(`${BASE_URL}/admin/announcement`)
}

export function createAnnouncement(
  cmd: SaveAnnouncementCommand
): Promise<AnnouncementDTO> {
  return apiClient(`${BASE_URL}/admin/announcement`, {
    method: "POST",
    body: JSON.stringify(cmd),
  })
}

export function updateAnnouncement(
  id: number,
  cmd: SaveAnnouncementCommand
): Promise<AnnouncementDTO> {
  return apiClient(`${BASE_URL}/admin/announcement/${id}`, {
    method: "PUT",
    body: JSON.stringify(cmd),
  })
}

export function deleteAnnouncement(id: number): Promise<void> {
  return apiClient(`${BASE_URL}/admin/announcement/${id}`, {
    method: "DELETE",
  })
}
