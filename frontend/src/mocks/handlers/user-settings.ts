import { http, HttpResponse } from "msw"
import type { UpdateUserSettingsCommand } from "@/api/user-settings"
import { MOCK_COURSE_SEMESTERS } from "../fixtures/courses"
import { randomDelay } from "../utils"

let currentSemester = MOCK_COURSE_SEMESTERS[0]

export const userSettingsHandlers = [
  http.get("/api/user/settings", async () => {
    await randomDelay()
    return HttpResponse.json({ current_semester: currentSemester })
  }),

  http.put("/api/user/settings", async ({ request }) => {
    await randomDelay()
    const body = (await request.json()) as UpdateUserSettingsCommand
    if (!MOCK_COURSE_SEMESTERS.includes(body.current_semester)) {
      return HttpResponse.json(
        { error: "invalid current semester" },
        { status: 400 }
      )
    }
    currentSemester = body.current_semester
    return HttpResponse.json({ current_semester: currentSemester })
  }),
]
