import { http, HttpResponse } from "msw"
import type { SaveAnnouncementCommand } from "@/api/announcement"
import {
  findUserByID,
  mockSession,
  setMockSessionUserID,
} from "@/mocks/fixtures/auth"
import { mockAnnouncements } from "@/mocks/fixtures/announcements"
import { randomDelay } from "@/mocks/utils"

function requireMockUser() {
  if (!mockSession.userID) return null
  const user = findUserByID(mockSession.userID)
  if (!user) {
    setMockSessionUserID(null)
    return null
  }
  return user
}

function sortAnnouncements() {
  return [...mockAnnouncements].sort((a, b) => b.priority - a.priority)
}

export const announcementHandlers = [
  http.get("/api/announcement/", async () => {
    await randomDelay()
    if (!requireMockUser()) {
      return HttpResponse.json({ error: "unauthorized" }, { status: 401 })
    }

    return HttpResponse.json(sortAnnouncements())
  }),
  http.get("/api/admin/announcement", async () => {
    await randomDelay()
    const user = requireMockUser()
    if (!user) {
      return HttpResponse.json({ error: "unauthorized" }, { status: 401 })
    }
    if (!user.is_admin()) {
      return HttpResponse.json({ error: "forbidden" }, { status: 403 })
    }

    return HttpResponse.json(sortAnnouncements())
  }),
  http.post("/api/admin/announcement", async ({ request }) => {
    await randomDelay()
    const user = requireMockUser()
    if (!user) {
      return HttpResponse.json({ error: "unauthorized" }, { status: 401 })
    }
    if (!user.is_admin()) {
      return HttpResponse.json({ error: "forbidden" }, { status: 403 })
    }
    const cmd = (await request.json()) as SaveAnnouncementCommand
    const nextID = Math.max(0, ...mockAnnouncements.map((item) => item.id)) + 1
    const item = {
      id: nextID,
      title: cmd.title,
      body: cmd.body,
      priority: cmd.priority,
      show_start: cmd.show_start,
      show_end: cmd.show_end,
      link_url: cmd.link_url ?? "",
      link_title: cmd.link_title ?? "",
      created_at: new Date().toISOString(),
    }
    mockAnnouncements.unshift(item)
    return HttpResponse.json(item, { status: 201 })
  }),
  http.put(
    "/api/admin/announcement/:announcementID",
    async ({ params, request }) => {
      await randomDelay()
      const user = requireMockUser()
      if (!user) {
        return HttpResponse.json({ error: "unauthorized" }, { status: 401 })
      }
      if (!user.is_admin()) {
        return HttpResponse.json({ error: "forbidden" }, { status: 403 })
      }
      const id = Number(params.announcementID)
      const index = mockAnnouncements.findIndex((item) => item.id === id)
      if (index < 0) {
        return HttpResponse.json({ error: "not found" }, { status: 404 })
      }
      const cmd = (await request.json()) as SaveAnnouncementCommand
      const item = {
        ...mockAnnouncements[index],
        title: cmd.title,
        body: cmd.body,
        priority: cmd.priority,
        show_start: cmd.show_start,
        show_end: cmd.show_end,
        link_url: cmd.link_url ?? "",
        link_title: cmd.link_title ?? "",
      }
      mockAnnouncements[index] = item
      return HttpResponse.json(item)
    }
  ),
  http.delete("/api/admin/announcement/:announcementID", async ({ params }) => {
    await randomDelay()
    const user = requireMockUser()
    if (!user) {
      return HttpResponse.json({ error: "unauthorized" }, { status: 401 })
    }
    if (!user.is_admin()) {
      return HttpResponse.json({ error: "forbidden" }, { status: 403 })
    }
    const id = Number(params.announcementID)
    const index = mockAnnouncements.findIndex((item) => item.id === id)
    if (index < 0) {
      return HttpResponse.json({ error: "not found" }, { status: 404 })
    }
    mockAnnouncements.splice(index, 1)
    return new HttpResponse(null, { status: 204 })
  }),
]
