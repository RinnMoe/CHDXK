import { http, HttpResponse } from "msw"
import { mockAnnouncements } from "../fixtures/announcements"
import { randomDelay } from "../utils"

export const announcementHandlers = [
  http.get("/api/announcement/", async () => {
    await randomDelay()
    const sorted = [...mockAnnouncements].sort((a, b) => b.priority - a.priority)
    return HttpResponse.json(sorted)
  }),
]
