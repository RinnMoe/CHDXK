import { http, HttpResponse } from "msw"
import {
  SYSTEM_SETTING_CURRENT_SEMESTER,
  type SystemSettingDTO,
  type UpdateSystemSettingCommand,
} from "@/api/system-settings"
import { MOCK_COURSE_SEMESTERS } from "@/mocks/fixtures/courses"
import { randomDelay } from "@/mocks/utils"
import defaultPages from "@/config/site-pages-defaults.json"
import { defaultTheme, parseTheme, themeSettingKey } from "@/lib/site-theme"
import type { SitePage } from "@/api/site-pages"
import { copyDirectory } from "@/lib/copy"

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
    key: "review.rewards.course_first_review_enabled",
    value: "false",
    default_value: "false",
    type: "bool",
    group: "review_reward",
    label: "启用首评奖励",
    description: "是否启用课程首评积分奖励",
    public: false,
    secret: false,
    requires_restart: false,
  },
  {
    key: "review.rewards.course_first_review_points",
    value: "0",
    default_value: "0",
    type: "int",
    group: "review_reward",
    label: "课程首评奖励积分",
    description: "课程首次点评奖励的积分数量",
    public: false,
    secret: false,
    requires_restart: false,
  },
  {
    key: "review.rewards.review_create_enabled",
    value: "false",
    default_value: "false",
    type: "bool",
    group: "review_reward",
    label: "启用发布点评奖励",
    description: "是否启用每条新点评的积分奖励",
    public: false,
    secret: false,
    requires_restart: false,
  },
  {
    key: "review.rewards.review_create_points",
    value: "1",
    default_value: "1",
    type: "int",
    group: "review_reward",
    label: "发布点评奖励积分",
    description: "每条新点评奖励的积分数量",
    public: false,
    secret: false,
    requires_restart: false,
  },
  {
    key: "review.hot_scores.review_create_score",
    value: "10",
    default_value: "10",
    type: "int",
    group: "review",
    label: "发布点评热度分",
    description: "发布点评时给课程热度增加的分值",
    public: false,
    secret: false,
    requires_restart: false,
  },
  {
    key: "review.hot_scores.review_update_score",
    value: "3",
    default_value: "3",
    type: "int",
    group: "review",
    label: "更新点评热度分",
    description: "更新点评时给课程热度增加的分值",
    public: false,
    secret: false,
    requires_restart: false,
  },
  {
    key: "review.hot_scores.review_vote_score",
    value: "1",
    default_value: "1",
    type: "int",
    group: "review",
    label: "点评投票热度分",
    description: "点评获得投票时给课程热度增加的分值",
    public: false,
    secret: false,
    requires_restart: false,
  },
  {
    key: "review.frequency.violation_admin_emails",
    value: "",
    default_value: "",
    type: "string_list",
    group: "review",
    label: "刷评通知管理员邮箱",
    description:
      "命中刷评策略后接收通知的管理员邮箱，多个值用逗号或换行分隔，留空则不发送",
    public: false,
    secret: false,
    requires_restart: false,
  },
]

settings.push(
  ...[
    {
      key: "content.pages",
      label: "内容页面",
      value: JSON.stringify(defaultPages),
      group: "content",
    },
    {
      key: themeSettingKey,
      label: "主题外观",
      value: JSON.stringify(defaultTheme),
      group: "theme",
    },
  ].map((item) => ({
    ...item,
    default_value: item.value,
    type: "string" as const,
    description: "",
    public: false,
    secret: false,
    requires_restart: false,
  })),
  ...Object.entries(copyDirectory)
    .filter(([, entry]) => entry.editable)
    .map(([key, entry]) => ({
      key,
      value: entry.text,
      default_value: entry.text,
      type: "string" as const,
      group: `copy_${entry.audience}`,
      label: entry.label ?? entry.text,
      description: "保存后刷新页面生效，请保留内容标记。",
      public: true,
      secret: false,
      requires_restart: false,
    }))
)

export function getMockSystemSettingValue(key: string) {
  return settings.find((item) => item.key === key)?.value
}

export const systemSettingsHandlers = [
  http.get("/api/site-pages", () =>
    HttpResponse.json(
      (
        JSON.parse(getMockSystemSettingValue("content.pages")!) as SitePage[]
      ).filter((item) => item.published)
    )
  ),
  http.get("/api/site-theme", () =>
    HttpResponse.json(
      parseTheme(JSON.parse(getMockSystemSettingValue(themeSettingKey)!))
    )
  ),

  http.get("/api/site-copy", () =>
    HttpResponse.json(
      Object.fromEntries(
        settings
          .filter((item) => item.key.startsWith("copy."))
          .map((item) => [item.key, item.value])
      )
    )
  ),
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
