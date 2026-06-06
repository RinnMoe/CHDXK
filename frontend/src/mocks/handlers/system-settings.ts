import { http, HttpResponse } from "msw"
import {
  SYSTEM_SETTING_CURRENT_SEMESTER,
  type SystemSettingDTO,
  type UpdateSystemSettingCommand,
} from "@/api/system-settings"
import { MOCK_COURSE_SEMESTERS } from "../fixtures/courses"
import { randomDelay } from "../utils"

let settings: SystemSettingDTO[] = [
  { key: SYSTEM_SETTING_CURRENT_SEMESTER, value: MOCK_COURSE_SEMESTERS[0] },
]

export const systemSettingsHandlers = [
  http.get("/api/system-settings", async () => {
    await randomDelay()
    return HttpResponse.json(settings)
  }),

  http.put("/api/admin/system-settings/:key", async ({ params, request }) => {
    await randomDelay()
    const key = String(params.key)
    const body = (await request.json()) as UpdateSystemSettingCommand
    if (
      key === SYSTEM_SETTING_CURRENT_SEMESTER &&
      !MOCK_COURSE_SEMESTERS.includes(body.value)
    ) {
      return HttpResponse.json(
        { error: "invalid current semester" },
        { status: 400 }
      )
    }

    const updated = { key, value: body.value }
    const index = settings.findIndex((item) => item.key === key)
    settings =
      index < 0
        ? [...settings, updated]
        : settings.map((item, i) => (i === index ? updated : item))
    return HttpResponse.json(updated)
  }),
]
