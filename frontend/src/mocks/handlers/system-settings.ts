import { http, HttpResponse } from "msw"
import {
  SYSTEM_SETTING_CURRENT_SEMESTER,
  type SystemSettingDTO,
  type UpdateSystemSettingCommand,
} from "@/api/system-settings"
import { MOCK_COURSE_SEMESTERS } from "../fixtures/courses"
import { randomDelay } from "../utils"

let settings: SystemSettingDTO[] = [
  {
    key: SYSTEM_SETTING_CURRENT_SEMESTER,
    value: MOCK_COURSE_SEMESTERS[0],
    default_value: "",
    type: "string",
    group: "course",
    label: "当前学期",
    description: "前台课程和写评默认使用的当前学期",
    public: true,
    secret: false,
    requires_restart: false,
  },
  {
    key: "review.vote.max_daily_votes",
    value: "50",
    default_value: "50",
    type: "int",
    group: "review",
    label: "每日投票上限",
    description: "单个用户每天可投票的最大次数",
    public: false,
    secret: false,
    requires_restart: false,
  },
  {
    key: "review.rewards.enabled",
    value: "false",
    default_value: "false",
    type: "bool",
    group: "review",
    label: "启用首评奖励",
    description: "是否启用课程首评积分奖励",
    public: false,
    secret: false,
    requires_restart: false,
  },
]

export const systemSettingsHandlers = [
  http.get("/api/system-settings", async () => {
    await randomDelay()
    return HttpResponse.json(settings)
  }),

  http.get("/api/admin/system-settings", async () => {
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

    const previous = settings.find((item) => item.key === key)
    const updated = previous
      ? { ...previous, value: body.value }
      : {
          key,
          value: body.value,
          default_value: "",
          type: "string" as const,
          group: "other",
          label: key,
          description: "",
          public: false,
          secret: false,
          requires_restart: false,
        }
    const index = settings.findIndex((item) => item.key === key)
    settings =
      index < 0
        ? [...settings, updated]
        : settings.map((item, i) => (i === index ? updated : item))
    return HttpResponse.json(updated)
  }),
]
