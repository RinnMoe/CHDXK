import { http, HttpResponse } from "msw"
import { getMockTeachers, makeTeacherFilters } from "@/mocks/fixtures/teachers"
import { mockCourses } from "@/mocks/fixtures/courses"
import { randomDelay } from "@/mocks/utils"

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

export const teacherHandlers = [
  http.get("/api/teacher/filter", async () => {
    await randomDelay()
    return HttpResponse.json(makeTeacherFilters())
  }),

  http.get("/api/teacher/", async ({ request }) => {
    await randomDelay()
    const url = new URL(request.url)
    const department = url.searchParams.get("department") ?? ""
    const title = url.searchParams.get("title") ?? ""
    const q = (url.searchParams.get("q") ?? "").trim().toLowerCase()
    const page = Number(url.searchParams.get("page") ?? "1")
    const pageSize = Number(url.searchParams.get("page_size") ?? "20")

    let list = getMockTeachers()
    if (department) list = list.filter((t) => t.department === department)
    if (title) list = list.filter((t) => t.title === title)
    if (q)
      list = list.filter(
        (t) =>
          t.code.toLowerCase().includes(q) || t.name.toLowerCase().includes(q)
      )

    return HttpResponse.json(paginate(list, page, pageSize))
  }),

  http.get("/api/teacher/:teacherID", async ({ params }) => {
    await randomDelay()
    const teacherID = Number(params.teacherID)
    const teacher = getMockTeachers().find((t) => t.id === teacherID)

    if (!teacher) {
      return HttpResponse.json({ error: "teacher not found" }, { status: 404 })
    }

    return HttpResponse.json(teacher)
  }),

  http.get("/api/teacher/:teacherID/course", async ({ params, request }) => {
    await randomDelay()
    const teacherID = Number(params.teacherID)
    const url = new URL(request.url)
    const page = Number(url.searchParams.get("page") ?? "1")
    const pageSize = Number(url.searchParams.get("page_size") ?? "20")

    const courses = mockCourses.filter((c) => c.main_teacher.id === teacherID)
    return HttpResponse.json(paginate(courses, page, pageSize))
  }),
]
