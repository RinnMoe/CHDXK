import { http, HttpResponse } from "msw"
import { mockAnnouncements } from "../fixtures/announcements"

export const announcementHandlers = [
  http.get("/api/announcement/", () => {
    const sorted = [...mockAnnouncements].sort((a, b) => b.priority - a.priority)
    return HttpResponse.json(sorted)
  }),
]
