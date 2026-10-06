import { describe, expect, it } from "vitest"
import { router } from "./router"

describe("review dialog routes", () => {
  it.each([
    [
      "/course/4/review/new",
      "/app/course/$courseID",
      "/app/course/$courseID/review/new",
    ],
    ["/review/2/edit", "/app/review/$reviewID", "/app/review/$reviewID/edit"],
  ])("keeps the detail background for %s", (path, detail, dialog) => {
    const ids = router.matchRoutes(path).map((match) => match.routeId)
    expect(ids.slice(-2)).toEqual([detail, dialog])
  })
})
