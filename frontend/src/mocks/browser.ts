import { setupWorker } from "msw/browser"
import { announcementHandlers } from "./handlers/announcement"
import { apiKeyHandlers } from "./handlers/api-key"
import { authHandlers } from "./handlers/auth"
import { courseHandlers } from "./handlers/course"
import { pointHandlers } from "./handlers/point"
import { reviewHandlers } from "./handlers/review"
import { siteStatsHandlers } from "./handlers/site-stats"
import { systemSettingsHandlers } from "./handlers/system-settings"
import { teacherHandlers } from "./handlers/teacher"
import { userSettingsHandlers } from "./handlers/user-settings"

export const worker = setupWorker(
  ...authHandlers,
  ...apiKeyHandlers,
  ...courseHandlers,
  ...reviewHandlers,
  ...teacherHandlers,
  ...pointHandlers,
  ...userSettingsHandlers,
  ...systemSettingsHandlers,
  ...announcementHandlers,
  ...siteStatsHandlers
)
