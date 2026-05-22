import { setupWorker } from "msw/browser"
import { courseHandlers } from "./handlers/course"
import { reviewHandlers } from "./handlers/review"

export const worker = setupWorker(...courseHandlers, ...reviewHandlers)
