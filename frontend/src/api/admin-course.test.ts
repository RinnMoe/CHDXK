import { afterEach, expect, it, vi } from "vitest"
import { listManagedCourses } from "./admin-course"

afterEach(() => vi.unstubAllGlobals())

it("sends independent column filters, zero bounds, sorting and pagination", async () => {
  const fetchMock = vi.fn(
    async (_input: string) =>
      new Response(JSON.stringify({ items: [], total: 0 }), {
        headers: { "Content-Type": "application/json" },
      })
  )
  vi.stubGlobal("fetch", fetchMock)
  await listManagedCourses({
    semester: "2026-2027-1",
    code: "A_01",
    name: "100%",
    teacher: "Alice",
    department: "First",
    lesson_no: "A.01",
    min_sections: 0,
    max_sections: 2,
    sort: "sections",
    direction: "desc",
    page: 2,
    page_size: 20,
  })
  const url = new URL(fetchMock.mock.calls[0][0], "http://localhost")
  expect(Object.fromEntries(url.searchParams)).toEqual({
    semester: "2026-2027-1",
    code: "A_01",
    name: "100%",
    teacher: "Alice",
    department: "First",
    lesson_no: "A.01",
    min_sections: "0",
    max_sections: "2",
    sort: "sections",
    direction: "desc",
    page: "2",
    page_size: "20",
  })
  expect(url.pathname).toBe("/api/admin/course")
})
