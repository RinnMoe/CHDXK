import type { PaginatedResult } from "./types"
import type { CourseListItemDTO } from "./course"
import { apiClient } from "./client"
import { BASE_URL } from "./constants"

export interface FilterItem {
  name: string
  count: number
}

export interface TeacherFilters {
  departments?: FilterItem[]
  titles?: FilterItem[]
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
  q?: string
  page?: number
  page_size?: number
  [key: string]: unknown
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
  return apiClient(`${BASE_URL}/teacher/filter`)
}

export function listTeachers(
  filter: TeacherListFilter = {}
): Promise<PaginatedResult<TeacherDTO>> {
  return apiClient(`${BASE_URL}/teacher/${buildQuery(filter)}`)
}

export function getTeacher(teacherID: number): Promise<TeacherDTO> {
  return apiClient(`${BASE_URL}/teacher/${teacherID}`)
}

export function listTeacherCourses(
  teacherID: number,
  filter: Record<string, unknown> = {}
): Promise<PaginatedResult<CourseListItemDTO>> {
  return apiClient(`${BASE_URL}/teacher/${teacherID}/course${buildQuery(filter)}`)
}
