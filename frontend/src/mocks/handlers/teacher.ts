import { http, HttpResponse } from "msw"
import {
  getMockTeachers,
  makeTeacherFilters,
} from "../fixtures/teachers"
import { mockCourses } from "../fixtures/courses"

function paginate<T>(items: T[], page: number, pageSize: number) {
  const start = (page - 1) * pageSize
  return {
    items: items.slice(start, start + pageSize),
    total: items.length,
    page,
    page_size: pageSize,
  }
}

export const teacherHandlers = [
  http.get("/api/teacher/filters", () => {
    return HttpResponse.json(makeTeacherFilters())
  }),

  http.get("/api/teacher/", ({ request }) => {
    const url = new URL(request.url)
    const department = url.searchParams.get("department") ?? ""
    const title = url.searchParams.get("title") ?? ""
    const pinyin = (url.searchParams.get("pinyin") ?? "").toLowerCase()
    const page = Number(url.searchParams.get("page") ?? "1")
    const pageSize = Number(url.searchParams.get("page_size") ?? "20")

    let list = getMockTeachers()
    if (department) list = list.filter((t) => t.department === department)
    if (title) list = list.filter((t) => t.title === title)
    if (pinyin) list = list.filter((t) => t.name.toLowerCase().includes(pinyin))

    return HttpResponse.json(paginate(list, page, pageSize))
  }),

  http.get("/api/teacher/:teacherID/courses", ({ params, request }) => {
    const teacherID = Number(params.teacherID)
    const url = new URL(request.url)
    const page = Number(url.searchParams.get("page") ?? "1")
    const pageSize = Number(url.searchParams.get("page_size") ?? "20")

    const courses = mockCourses.filter((c) => c.main_teacher.id === teacherID)
    return HttpResponse.json(paginate(courses, page, pageSize))
  }),
]
