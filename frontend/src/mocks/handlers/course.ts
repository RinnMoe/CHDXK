import { http, HttpResponse } from "msw"
import {
  mockCourses,
  makeCourseDetail,
  makeCourseFilters,
} from "../fixtures/courses"
import { mockReviews } from "../fixtures/reviews"

const filters = makeCourseFilters()

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
  if (code) list = list.filter((c) => c.code.toLowerCase().includes(code.toLowerCase()) || c.name.includes(code))
  if (department) list = list.filter((c) => c.main_teacher.department === department)
  if (language) list = list.filter((c) => c.language === language)
  if (categories.length > 0) list = list.filter((c) => c.categories.some((cat) => categories.includes(cat)))
  if (targetYears.length > 0) list = list.filter((c) => c.target_years.some((y) => targetYears.includes(y)))

  if (orderBy === "rating_avg") {
    list.sort((a, b) => (ascend ? a.rating.avg - b.rating.avg : b.rating.avg - a.rating.avg))
  } else if (orderBy === "rating_count") {
    list.sort((a, b) => (ascend ? a.rating.count - b.rating.count : b.rating.count - a.rating.count))
  }
  return list
}

export const courseHandlers = [
  http.get("/api/course/filters", () => HttpResponse.json(filters)),

  http.get("/api/course/", ({ request }) => {
    const url = new URL(request.url)
    const list = applyCourseFilter(url)
    const page = Number(url.searchParams.get("page") ?? "1")
    const pageSize = Number(url.searchParams.get("page_size") ?? "20")
    return HttpResponse.json(paginate(list, page, pageSize))
  }),

  http.get("/api/course/followed", ({ request }) => {
    const url = new URL(request.url)
    const list = mockCourses.slice(0, 5)
    const page = Number(url.searchParams.get("page") ?? "1")
    const pageSize = Number(url.searchParams.get("page_size") ?? "20")
    return HttpResponse.json(paginate(list, page, pageSize))
  }),

  http.get("/api/course/ignored", ({ request }) => {
    const url = new URL(request.url)
    const list = mockCourses.slice(5, 8)
    const page = Number(url.searchParams.get("page") ?? "1")
    const pageSize = Number(url.searchParams.get("page_size") ?? "20")
    return HttpResponse.json(paginate(list, page, pageSize))
  }),

  http.get("/api/course/:courseID", ({ params }) => {
    const id = Number(params.courseID)
    const course = mockCourses.find((c) => c.id === id)
    if (!course) {
      return HttpResponse.json({ error: "course not found" }, { status: 404 })
    }
    return HttpResponse.json(makeCourseDetail(course))
  }),

  http.get("/api/course/:courseID/review", ({ params, request }) => {
    const id = Number(params.courseID)
    const url = new URL(request.url)
    const page = Number(url.searchParams.get("page") ?? "1")
    const pageSize = Number(url.searchParams.get("page_size") ?? "20")
    const reviews = mockReviews.filter((r) => r.course_id === id)
    return HttpResponse.json(paginate(reviews, page, pageSize))
  }),

  http.post("/api/course/:courseID/notification", async ({ request }) => {
    await request.json()
    return HttpResponse.json({ message: "ok" })
  }),
]
