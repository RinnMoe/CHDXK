import type {
  PaginatedResult,
  FilterItem,
  RatingInfoDTO,
  TeacherDTO,
} from "./types"
export type { RatingInfoDTO } from "./types"
import type { ReviewDTO, ReviewListFilter } from "./review"
import { apiClient } from "./client"

const BASE = "/api"

export interface CourseFilters {
  credits: FilterItem[]
  departments: FilterItem[]
  categories: FilterItem[]
  target_years: FilterItem[]
  languages: FilterItem[]
}

export interface CourseReviewFilters {
  semesters: FilterItem[]
  ratings: FilterItem[]
}

export interface CourseReviewTrendItemDTO {
  semester: string
  avg: number
  count: number
}

export type CourseNotificationLevel = 0 | 1 | 2

export interface CourseListItemDTO {
  id: number
  code: string
  name: string
  credit: number
  department: string
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
  notification_level: CourseNotificationLevel
  my_review?: ReviewDTO
}

export interface HotCourseItemDTO {
  course: CourseListItemDTO
  score: number
}

export interface HotCourseListDTO {
  period: "week" | "month"
  items: HotCourseItemDTO[]
}

export interface CourseListFilter {
  q?: string
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
  [key: string]: unknown
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

export function getCourseDetail(courseID: number): Promise<CourseDetailDTO> {
  return apiClient(`${BASE}/course/${courseID}`)
}

export function listCourseReviews(
  courseID: number,
  filter: ReviewListFilter = {}
): Promise<PaginatedResult<ReviewDTO>> {
  return apiClient(`${BASE}/course/${courseID}/review${buildQuery(filter)}`)
}

export function getCourseReviewFilters(
  courseID: number
): Promise<CourseReviewFilters> {
  return apiClient(`${BASE}/course/${courseID}/review/filters`)
}

export function getCourseReviewTrend(
  courseID: number
): Promise<CourseReviewTrendItemDTO[]> {
  return apiClient(`${BASE}/course/${courseID}/review/trend`)
}

export function setNotificationLevel(
  courseID: number,
  level: CourseNotificationLevel
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

export function listHotCourses(
  period: "week" | "month" = "week",
  limit?: number
): Promise<HotCourseListDTO> {
  const params = new URLSearchParams({ period })
  if (limit !== undefined) params.set("limit", String(limit))
  return apiClient(`${BASE}/course/hot?${params}`)
}
