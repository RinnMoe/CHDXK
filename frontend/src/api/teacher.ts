import type { PaginatedResult } from "./types"
import type { CourseListItemDTO } from "./course"
import { apiClient } from "./client"

const BASE = "/api"

export interface FilterItem {
  name: string
  count: number
}

export interface TeacherFilters {
  departments: FilterItem[]
  titles: FilterItem[]
}

export interface TeacherDTO {
  id: number
  code: string
  name: string
  department: string
  title?: string
}

export interface TeacherListFilter {
  department?: string
  title?: string
  pinyin?: string
  page?: number
  page_size?: number
}

function buildQuery(filter: Record<string, unknown>): string {
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(filter)) {
    if (value === undefined || value === null || value === "") continue
    if (Array.isArray(value)) {
      for (const v of value) params.append(key, String(v))
    } else {
      params.append(key, String(value))
    }
  }
  const q = params.toString()
  return q ? `?${q}` : ""
}

export function getTeacherFilters(): Promise<TeacherFilters> {
  return apiClient(`${BASE}/teacher/filters`)
}

export function listTeachers(
  filter: TeacherListFilter = {}
): Promise<PaginatedResult<TeacherDTO>> {
  return apiClient(`${BASE}/teacher/${buildQuery(filter)}`)
}

export function listTeacherCourses(
  teacherID: number,
  filter: Record<string, unknown> = {}
): Promise<PaginatedResult<CourseListItemDTO>> {
  return apiClient(`${BASE}/teacher/${teacherID}/courses${buildQuery(filter)}`)
}
