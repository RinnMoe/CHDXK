import { setupWorker } from "msw/browser"
import { courseHandlers } from "./handlers/course"
import { reviewHandlers } from "./handlers/review"
import { teacherHandlers } from "./handlers/teacher"

export const worker = setupWorker(...courseHandlers, ...reviewHandlers, ...teacherHandlers)
