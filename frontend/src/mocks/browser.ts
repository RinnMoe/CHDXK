import { setupWorker } from "msw/browser"
import { courseHandlers } from "./handlers/course"

export const worker = setupWorker(...courseHandlers)
