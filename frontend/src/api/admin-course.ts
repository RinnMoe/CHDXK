import { apiClient } from "./client"
import { BASE_URL } from "./constants"
import type { CourseSectionDTO } from "./course"
import type { PaginatedResult } from "./types"

export interface ManagedCourseOfferingDTO {
  id: number
  course_id: number
  code: string
  semester: string
  sections: CourseSectionDTO[]
}

export interface SaveManagedCourseOffering {
  code: string
  semester: string
  sections: Omit<CourseSectionDTO, "id">[]
}

export interface AdminCourseFilter {
  semester: string
  code?: string
  name?: string
  teacher?: string
  department?: string
  lesson_no?: string
  min_sections?: number
  max_sections?: number
  sort?: "code" | "name" | "teacher" | "department" | "sections"
  direction?: "asc" | "desc"
  page?: number
  page_size?: number
}

export function listManagedCourseSemesters(): Promise<string[]> {
  return apiClient(`${BASE_URL}/admin/course/semester`)
}

export function listManagedCourses(
  filter: AdminCourseFilter
): Promise<PaginatedResult<ManagedCourseOfferingDTO>> {
  const params = new URLSearchParams()
  params.set("semester", filter.semester)
  for (const key of [
    "code",
    "name",
    "teacher",
    "department",
    "lesson_no",
    "min_sections",
    "max_sections",
    "sort",
    "direction",
  ] as const) {
    const value = filter[key]
    if (value !== undefined && value !== "") params.set(key, String(value))
  }
  if (filter.page) params.set("page", String(filter.page))
  if (filter.page_size) params.set("page_size", String(filter.page_size))
  return apiClient(`${BASE_URL}/admin/course?${params.toString()}`)
}

export function createManagedCourse(
  command: SaveManagedCourseOffering
): Promise<ManagedCourseOfferingDTO> {
  return apiClient(`${BASE_URL}/admin/course`, {
    method: "POST",
    body: JSON.stringify(command),
  })
}

export function updateManagedCourse(
  id: number,
  command: SaveManagedCourseOffering
): Promise<ManagedCourseOfferingDTO> {
  return apiClient(`${BASE_URL}/admin/course/${id}`, {
    method: "PUT",
    body: JSON.stringify(command),
  })
}

export function deleteManagedCourse(id: number): Promise<void> {
  return apiClient(`${BASE_URL}/admin/course/${id}`, { method: "DELETE" })
}
