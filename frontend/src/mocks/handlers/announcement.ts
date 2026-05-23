import { http, HttpResponse } from "msw"
import { findUserByID, mockSession, setMockSessionUserID } from "../fixtures/auth"
import { mockAnnouncements } from "../fixtures/announcements"
import { randomDelay } from "../utils"

export const announcementHandlers = [
  http.get("/api/announcement/", async () => {
    await randomDelay()
    if (!mockSession.userID) {
      return HttpResponse.json({ error: "unauthorized" }, { status: 401 })
    }
    if (!findUserByID(mockSession.userID)) {
      setMockSessionUserID(null)
      return HttpResponse.json({ error: "unauthorized" }, { status: 401 })
    }

    const sorted = [...mockAnnouncements].sort(
      (a, b) => b.priority - a.priority
    )
    return HttpResponse.json(sorted)
  }),
]
