import type {
  PaginatedResult,
  FilterItem,
  RatingInfoDTO,
  TeacherDTO,
} from "./types"
import { apiClient } from "./client"

const BASE = "/api"

export interface CourseFilters {
  credits: FilterItem[]
  departments: FilterItem[]
  categories: FilterItem[]
  target_years: FilterItem[]
}

export interface CourseListItemDTO {
  id: number
  code: string
  name: string
  credit: number
  language: string
  target_years: string[]
  categories: string[]
  main_teacher: TeacherDTO
  rating: RatingInfoDTO
}

export interface OfferedCourseDTO {
  semester: string
  language: string
  target_years: string[]
  categories: string[]
  teacher_group: TeacherDTO[]
}

export interface CourseDetailDTO {
  id: number
  code: string
  name: string
  credit: number
  department: string
  language: string
  target_years: string[]
  categories: string[]
  main_teacher: TeacherDTO
  offered_courses: OfferedCourseDTO[]
  rating: RatingInfoDTO
  same_code_courses: CourseListItemDTO[]
  same_teacher_courses: CourseListItemDTO[]
  notification_level: number
}

export interface ReviewDTO {
  id: number
  course?: CourseListItemDTO
  course_id: number
  score: string
  rating: number
  content: string
  vote: {
    like_count: number
    dislike_count: number
    my_vote?: number
  }
  created_at: string
  updated_at: string
}

export interface CourseListFilter {
  code?: string
  department?: string
  language?: string
  categories?: string[]
  target_years?: string[]
  credit?: number
  has_review?: boolean
  order_by?: "rating_count" | "rating_avg"
  ascend?: boolean
  page?: number
  page_size?: number
}

export interface ReviewListFilter {
  page?: number
  page_size?: number
  order_by?: string
  ascend?: boolean
}

function buildQuery(filter: Record<string, unknown>): string {
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(filter)) {
    if (value === undefined || value === null) continue
    if (Array.isArray(value)) {
      for (const v of value) {
        params.append(key, String(v))
      }
    } else if (typeof value === "boolean") {
      params.append(key, String(Number(value)))
    } else {
      params.append(key, String(value))
    }
  }
  const q = params.toString()
  return q ? `?${q}` : ""
}

export function getCourseFilters(): Promise<CourseFilters> {
  return apiClient(`${BASE}/course/filters`)
}

export function listCourses(
  filter: CourseListFilter = {}
): Promise<PaginatedResult<CourseListItemDTO>> {
  return apiClient(`${BASE}/course/${buildQuery(filter)}`)
}

export function getCourseDetail(
  courseID: number
): Promise<CourseDetailDTO> {
  return apiClient(`${BASE}/course/${courseID}`)
}

export function listCourseReviews(
  courseID: number,
  filter: ReviewListFilter = {}
): Promise<PaginatedResult<ReviewDTO>> {
  return apiClient(`${BASE}/course/${courseID}/review${buildQuery(filter)}`)
}

export function setNotificationLevel(
  courseID: number,
  level: number
): Promise<{ message: string }> {
  return apiClient(`${BASE}/course/${courseID}/notification`, {
    method: "POST",
    body: JSON.stringify({ level }),
  })
}

export function listFollowedCourses(
  filter: CourseListFilter = {}
): Promise<PaginatedResult<CourseListItemDTO>> {
  return apiClient(`${BASE}/course/followed${buildQuery(filter)}`)
}

export function listIgnoredCourses(
  filter: CourseListFilter = {}
): Promise<PaginatedResult<CourseListItemDTO>> {
  return apiClient(`${BASE}/course/ignored${buildQuery(filter)}`)
}
