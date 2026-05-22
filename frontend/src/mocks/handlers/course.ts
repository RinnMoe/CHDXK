import { http, HttpResponse } from "msw"
import {
  mockCourses,
  makeCourseDetail,
  makeCourseFilters,
} from "../fixtures/courses"
import { mockReviews } from "../fixtures/reviews"
import { randomDelay } from "../utils"
import type { CourseNotificationLevel } from "@/api/course"
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

function getNotificationLevel(courseID: number): CourseNotificationLevel {
  return notificationLevels.get(courseID) ?? 0
}

function paginate<T>(items: T[], page: number, pageSize: number) {
  const start = (page - 1) * pageSize
  return {
    items: items.slice(start, start + pageSize),
    total: items.length,
    page,
    page_size: pageSize,
  }
}

function applyCourseFilter(url: URL) {
  const code = url.searchParams.get("code") ?? ""
  const department = url.searchParams.get("department") ?? ""
  const language = url.searchParams.get("language") ?? ""
  const categories = url.searchParams.getAll("categories")
  const targetYears = url.searchParams.getAll("target_years")
  const orderBy = url.searchParams.get("order_by") ?? ""
  const ascend = url.searchParams.get("ascend") === "1"

  let list = [...mockCourses]
  if (code)
    list = list.filter(
      (c) =>
        c.code.toLowerCase().includes(code.toLowerCase()) ||
        c.name.includes(code)
    )
  if (department) list = list.filter((c) => c.department === department)
  if (language) list = list.filter((c) => c.language === language)
  if (categories.length > 0)
    list = list.filter((c) =>
      c.categories.some((cat) => categories.includes(cat))
    )
  if (targetYears.length > 0)
    list = list.filter((c) =>
      c.target_years.some((y) => targetYears.includes(y))
    )

  if (orderBy === "rating_avg") {
    list.sort((a, b) =>
      ascend ? a.rating.avg - b.rating.avg : b.rating.avg - a.rating.avg
    )
  } else if (orderBy === "rating_count") {
    list.sort((a, b) =>
      ascend ? a.rating.count - b.rating.count : b.rating.count - a.rating.count
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

export const courseHandlers = [
  http.get("/api/course/filters", async () => {
    await randomDelay()
    return HttpResponse.json(filters)
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
    const list = mockCourses.filter((course) => getNotificationLevel(course.id) === 1)
    const page = Number(url.searchParams.get("page") ?? "1")
    const pageSize = Number(url.searchParams.get("page_size") ?? "20")
    return HttpResponse.json(paginate(list, page, pageSize))
  }),

  http.get("/api/course/ignored", async ({ request }) => {
    await randomDelay()
    const url = new URL(request.url)
    const list = mockCourses.filter((course) => getNotificationLevel(course.id) === 2)
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
    return HttpResponse.json({
      ...makeCourseDetail(course),
      notification_level: getNotificationLevel(id),
    })
  }),

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

  http.get("/api/course/:courseID/review/filters", async ({ params }) => {
    await randomDelay()
    const id = Number(params.courseID)
    const reviews = mockReviews.filter((r) => r.course_id === id)
    return HttpResponse.json(makeReviewFilters(reviews))
  }),

  http.post("/api/course/:courseID/notification", async ({ params, request }) => {
    await randomDelay()
    const id = Number(params.courseID)
    const body = (await request.json()) as { level: CourseNotificationLevel }
    if (![0, 1, 2].includes(body.level)) {
      return HttpResponse.json({ error: "level must be 0, 1, or 2" }, { status: 400 })
    }
    if (body.level === 0) notificationLevels.delete(id)
    else notificationLevels.set(id, body.level)
    return HttpResponse.json({ message: "ok" })
  }),
]
