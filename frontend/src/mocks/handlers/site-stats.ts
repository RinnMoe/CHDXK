import { http, HttpResponse } from "msw"
import { findUserByID, mockSession } from "../fixtures/auth"
import { getMockYesterdayStats, mockDailyStats } from "../fixtures/site-stats"
import { randomDelay } from "../utils"

function requireAdmin() {
  if (!mockSession.userID) {
    return HttpResponse.json({ error: "unauthorized" }, { status: 401 })
  }
  const u = findUserByID(mockSession.userID)
  if (!u || u.role !== "admin") {
    return HttpResponse.json({ error: "forbidden" }, { status: 403 })
  }
  return null
}

export const siteStatsHandlers = [
  http.get("/api/site-stats/daily/:date", async ({ params }) => {
    await randomDelay()
    const guard = requireAdmin()
    if (guard) return guard
    const stat = getMockYesterdayStats()
    stat.stat_date = params.date as string
    return HttpResponse.json(stat)
  }),

  http.get("/api/site-stats/daily", async ({ request }) => {
    await randomDelay()
    const guard = requireAdmin()
    if (guard) return guard

    const url = new URL(request.url)
    const startDate = url.searchParams.get("start_date") ?? ""
    const endDate = url.searchParams.get("end_date") ?? ""
    const page = Number(url.searchParams.get("page") ?? "1")
    const pageSize = Number(url.searchParams.get("page_size") ?? "20")

    let items = [...mockDailyStats]
    if (startDate) items = items.filter((s) => s.stat_date >= startDate)
    if (endDate) items = items.filter((s) => s.stat_date <= endDate)
    items.sort((a, b) => (a.stat_date < b.stat_date ? 1 : -1))

    const start = (page - 1) * pageSize
    return HttpResponse.json({
      items: items.slice(start, start + pageSize),
      total: items.length,
      page,
      page_size: pageSize,
    })
  }),
]
