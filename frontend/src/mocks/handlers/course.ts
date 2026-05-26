import { http, HttpResponse } from "msw"
import {
  mockCourses,
  makeCourseDetail,
  makeCourseFilters,
} from "../fixtures/courses"
import { findUserByID, mockSession } from "../fixtures/auth"
import { mockReviews } from "../fixtures/reviews"
import { randomDelay } from "../utils"
import type {
  CourseNotificationLevel,
  CourseReviewTrendItemDTO,
} from "@/api/course"
import type { ReviewDTO } from "@/api/review"

const filters = makeCourseFilters()
const notificationLevels = new Map<number, CourseNotificationLevel>([
  [1, 1],
  [2, 1],
  [3, 1],
  [4, 1],
  [5, 1],
  [6, 2],
  [7, 2],
  [8, 2],
])
const moderatorRemarks = new Map<number, string>([
  [1, "课程信息已由管理员核对。"],
])

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

function getNotificationLevel(courseID: number): CourseNotificationLevel {
  return notificationLevels.get(courseID) ?? 0
}

function mockHotPeriodKey(period: "week" | "month") {
  const now = new Date()
  if (period === "month") {
    return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}`
  }

  const date = new Date(
    Date.UTC(now.getFullYear(), now.getMonth(), now.getDate())
  )
  const day = date.getUTCDay() || 7
  date.setUTCDate(date.getUTCDate() + 4 - day)
  const yearStart = new Date(Date.UTC(date.getUTCFullYear(), 0, 1))
  const week = Math.ceil(
    ((date.getTime() - yearStart.getTime()) / 86400000 + 1) / 7
  )
  return `${date.getUTCFullYear()}-${String(week).padStart(2, "0")}`
}

function shouldShowMockMyReview(
  courseID: number,
  user: { id: number; username: string }
) {
  if (user.username === "demo") return true
  return (courseID + user.id) % 2 === 0
}

function makeMockMyReview(
  courseID: number,
  user: { id: number; username: string }
) {
  if (!shouldShowMockMyReview(courseID, user)) return undefined
  const review = mockReviews.find((r) => r.course_id === courseID)
  if (!review) return undefined
  return { ...review, user_id: user.id }
}

function paginate<T>(items: T[], page: number, pageSize: number) {
  page = normalizePage(page)
  pageSize = normalizePageSize(pageSize)
  const start = (page - 1) * pageSize
  return {
    items: items.slice(start, start + pageSize),
    total: items.length,
    page,
    page_size: pageSize,
  }
}

function applyCourseFilter(url: URL) {
  const q = (url.searchParams.get("q") ?? "").trim().toLowerCase()
  const department = url.searchParams.get("department") ?? ""
  const language = url.searchParams.get("language") ?? ""
  const creditParam = url.searchParams.get("credit")
  const credit = creditParam === null ? undefined : Number(creditParam)
  const categories = url.searchParams.getAll("categories")
  const targetYears = url.searchParams.getAll("target_years")
  const orderBy = url.searchParams.get("order_by") ?? ""
  const ascend = url.searchParams.get("ascend") === "1"

  let list = [...mockCourses]
  if (q)
    list = list.filter(
      (c) =>
        c.code.toLowerCase().includes(q) ||
        c.name.toLowerCase().includes(q) ||
        c.main_teacher.name.toLowerCase().includes(q)
    )
  if (department) list = list.filter((c) => c.department === department)
  if (language) list = list.filter((c) => c.language === language)
  if (credit !== undefined && Number.isFinite(credit)) {
    list = list.filter((c) => c.credit === credit)
  }
  if (categories.length > 0)
    list = list.filter((c) =>
      c.categories?.some((cat) => categories.includes(cat))
    )
  if (targetYears.length > 0)
    list = list.filter((c) =>
      c.target_years?.some((y) => targetYears.includes(y))
    )

  if (orderBy === "rating_avg") {
    list.sort(
      (a, b) =>
        b.rating.avg - a.rating.avg ||
        b.rating.count - a.rating.count ||
        a.code.localeCompare(b.code)
    )
  } else if (orderBy === "rating_score" || !orderBy) {
    list.sort(
      (a, b) =>
        b.rating.score - a.rating.score ||
        b.rating.count - a.rating.count ||
        b.rating.avg - a.rating.avg ||
        a.code.localeCompare(b.code)
    )
  } else if (orderBy === "rating_count") {
    list.sort(
      (a, b) =>
        (ascend
          ? a.rating.count - b.rating.count
          : b.rating.count - a.rating.count) ||
        b.rating.avg - a.rating.avg ||
        a.code.localeCompare(b.code)
    )
  }
  return list
}

function applyReviewFilter(url: URL, list: ReviewDTO[]): ReviewDTO[] {
  const semester = url.searchParams.get("semester") ?? ""
  const rating = Number(url.searchParams.get("rating") ?? "0")
  const orderBy = url.searchParams.get("order_by") ?? "created_at"
  const ascend = url.searchParams.get("ascend") === "1"

  let filtered = [...list]
  if (semester) filtered = filtered.filter((r) => r.semester === semester)
  if (rating > 0) filtered = filtered.filter((r) => r.rating === rating)

  filtered.sort((a, b) => {
    if (orderBy === "like_count") {
      return ascend
        ? a.vote.like_count - b.vote.like_count
        : b.vote.like_count - a.vote.like_count
    }
    const aT = new Date(a.created_at).getTime()
    const bT = new Date(b.created_at).getTime()
    return ascend ? aT - bT : bT - aT
  })
  return filtered
}

function makeReviewFilters(list: ReviewDTO[]) {
  const semesters = Array.from(
    list.reduce((counts, review) => {
      if (review.semester) {
        counts.set(review.semester, (counts.get(review.semester) ?? 0) + 1)
      }
      return counts
    }, new Map<string, number>())
  )
    .sort(([a], [b]) => b.localeCompare(a))
    .map(([name, count]) => ({ name, count }))

  const ratings = [5, 4, 3, 2, 1].map((rating) => ({
    name: String(rating),
    count: list.filter((review) => review.rating === rating).length,
  }))

  return { semesters, ratings }
}

function makeReviewTrend(list: ReviewDTO[]): CourseReviewTrendItemDTO[] {
  const groups = list.reduce((map, review) => {
    if (!review.semester) return map
    const item = map.get(review.semester) ?? { total: 0, count: 0 }
    item.total += review.rating
    item.count += 1
    map.set(review.semester, item)
    return map
  }, new Map<string, { total: number; count: number }>())

  return Array.from(groups)
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([semester, item]) => ({
      semester,
      avg: item.count > 0 ? item.total / item.count : 0,
      count: item.count,
    }))
}

export const courseHandlers = [
  http.get("/api/course/filter", async () => {
    await randomDelay()
    return HttpResponse.json(filters)
  }),

  http.get("/api/course/hot", async ({ request }) => {
    await randomDelay()
    const url = new URL(request.url)
    const period = (url.searchParams.get("period") ?? "week") as
      | "week"
      | "month"
    const limit = Math.min(
      Number(url.searchParams.get("limit") ?? "5"),
      mockCourses.length
    )
    const items = mockCourses.slice(0, limit).map((course, i) => ({
      course,
      score: (limit - i) * 10 + Math.floor(Math.random() * 6),
    }))
    return HttpResponse.json({
      period,
      period_key: mockHotPeriodKey(period),
      items,
    })
  }),

  http.get("/api/course/", async ({ request }) => {
    await randomDelay()
    const url = new URL(request.url)
    const list = applyCourseFilter(url)
    const page = Number(url.searchParams.get("page") ?? "1")
    const pageSize = Number(url.searchParams.get("page_size") ?? "20")
    return HttpResponse.json(paginate(list, page, pageSize))
  }),

  http.get("/api/course/followed", async ({ request }) => {
    await randomDelay()
    const url = new URL(request.url)
    const list = mockCourses.filter(
      (course) => getNotificationLevel(course.id) === 1
    )
    const page = Number(url.searchParams.get("page") ?? "1")
    const pageSize = Number(url.searchParams.get("page_size") ?? "20")
    return HttpResponse.json(paginate(list, page, pageSize))
  }),

  http.get("/api/course/ignored", async ({ request }) => {
    await randomDelay()
    const url = new URL(request.url)
    const list = mockCourses.filter(
      (course) => getNotificationLevel(course.id) === 2
    )
    const page = Number(url.searchParams.get("page") ?? "1")
    const pageSize = Number(url.searchParams.get("page_size") ?? "20")
    return HttpResponse.json(paginate(list, page, pageSize))
  }),

  http.get("/api/course/:courseID", async ({ params }) => {
    await randomDelay()
    const id = Number(params.courseID)
    const course = mockCourses.find((c) => c.id === id)
    if (!course) {
      return HttpResponse.json({ error: "course not found" }, { status: 404 })
    }
    const user = mockSession.userID
      ? findUserByID(mockSession.userID)
      : undefined
    const myReview = user ? makeMockMyReview(id, user) : undefined
    return HttpResponse.json({
      ...makeCourseDetail(course),
      moderator_remark: moderatorRemarks.get(id) ?? "",
      notification_level: getNotificationLevel(id),
      my_review: myReview,
    })
  }),

  http.put(
    "/api/course/:courseID/moderator-remark",
    async ({ params, request }) => {
      await randomDelay()
      const user = mockSession.userID
        ? findUserByID(mockSession.userID)
        : undefined
      if (!user || user.role !== "admin") {
        return HttpResponse.json({ error: "forbidden" }, { status: 403 })
      }
      const id = Number(params.courseID)
      const course = mockCourses.find((c) => c.id === id)
      if (!course) {
        return HttpResponse.json({ error: "course not found" }, { status: 404 })
      }
      const body = (await request.json()) as { moderator_remark?: string }
      moderatorRemarks.set(id, body.moderator_remark ?? "")
      return HttpResponse.json({ message: "ok" })
    }
  ),

  http.get("/api/course/:courseID/review", async ({ params, request }) => {
    await randomDelay()
    const id = Number(params.courseID)
    const url = new URL(request.url)
    const page = Number(url.searchParams.get("page") ?? "1")
    const pageSize = Number(url.searchParams.get("page_size") ?? "20")
    const reviews = applyReviewFilter(
      url,
      mockReviews.filter((r) => r.course_id === id)
    )
    return HttpResponse.json(paginate(reviews, page, pageSize))
  }),

  http.get("/api/course/:courseID/review/filter", async ({ params }) => {
    await randomDelay()
    const id = Number(params.courseID)
    const reviews = mockReviews.filter((r) => r.course_id === id)
    return HttpResponse.json(makeReviewFilters(reviews))
  }),

  http.get("/api/course/:courseID/review/trend", async ({ params }) => {
    await randomDelay()
    const id = Number(params.courseID)
    const reviews = mockReviews.filter((r) => r.course_id === id)
    return HttpResponse.json(makeReviewTrend(reviews))
  }),

  http.post(
    "/api/course/:courseID/notification",
    async ({ params, request }) => {
      await randomDelay()
      const id = Number(params.courseID)
      const body = (await request.json()) as { level: CourseNotificationLevel }
      if (![0, 1, 2].includes(body.level)) {
        return HttpResponse.json(
          { error: "level must be 0, 1, or 2" },
          { status: 400 }
        )
      }
      if (body.level === 0) notificationLevels.delete(id)
      else notificationLevels.set(id, body.level)
      return HttpResponse.json({ message: "ok" })
    }
  ),
]
