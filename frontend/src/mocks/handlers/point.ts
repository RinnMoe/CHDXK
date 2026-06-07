import { http, HttpResponse } from "msw"
import { findUserByID, mockSession } from "../fixtures/auth"
import { getMockPointRecords, getMockPointTotal } from "../fixtures/points"
import { randomDelay } from "../utils"

const DEFAULT_PAGE = 1
const DEFAULT_PAGE_SIZE = 20
const MAX_PAGE_SIZE = 100

function normalizePage(value: number) {
  return value > 0 ? value : DEFAULT_PAGE
}

function normalizePageSize(value: number) {
  if (value <= 0) return DEFAULT_PAGE_SIZE
  return Math.min(value, MAX_PAGE_SIZE)
}

export const pointHandlers = [
  http.get("/api/user/:userID/point", async ({ params, request }) => {
    await randomDelay()
    if (!mockSession.userID) {
      return HttpResponse.json({ error: "unauthorized" }, { status: 401 })
    }
    const userID = Number(params.userID)
    if (mockSession.userID !== userID) {
      const me = findUserByID(mockSession.userID)
      if (!me || !me.is_admin()) {
        return HttpResponse.json({ error: "forbidden" }, { status: 403 })
      }
    }
    const url = new URL(request.url)
    const page = normalizePage(Number(url.searchParams.get("page") ?? "1"))
    const pageSize = normalizePageSize(
      Number(url.searchParams.get("page_size") ?? "20")
    )

    const records = getMockPointRecords(userID)
    const start = (page - 1) * pageSize
    const items = records.slice(start, start + pageSize)
    return HttpResponse.json({
      total: getMockPointTotal(userID),
      records: {
        items,
        total: records.length,
        page,
        page_size: pageSize,
      },
    })
  }),
]
