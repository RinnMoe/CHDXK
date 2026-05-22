import { setupWorker } from "msw/browser"
import { announcementHandlers } from "./handlers/announcement"
import { authHandlers } from "./handlers/auth"
import { courseHandlers } from "./handlers/course"
import { pointHandlers } from "./handlers/point"
import { reviewHandlers } from "./handlers/review"
import { siteStatsHandlers } from "./handlers/site-stats"
import { teacherHandlers } from "./handlers/teacher"

export const worker = setupWorker(
  ...authHandlers,
  ...courseHandlers,
  ...reviewHandlers,
  ...teacherHandlers,
  ...pointHandlers,
  ...announcementHandlers,
  ...siteStatsHandlers,
)
