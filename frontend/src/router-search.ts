import {
  enumParam,
  numberParam,
  optionalPositiveIntParam,
  rawStringParam,
  stringArrayParam,
  stringParam,
  type SearchRecord,
} from "@/lib/router/search"
import { getSafeRedirectPath } from "@/lib/auth-redirect"

export type LoginSearch = { redirect?: string }
export type CourseListSearch = {
  q?: string
  department?: string
  language?: string
  categories?: string[]
  target_years?: string[]
  credit?: number
  order_by?: "rating_score" | "rating_count"
  page?: number
}
export type CourseDetailSearch = {
  page?: number
  semester?: string
  rating?: number
  order_by?: "created_at" | "updated_at" | "like_count"
}
export type PageSearch = { page?: number }
export type ReviewsSearch = PageSearch & { q?: string }
export type UserCoursesSearch = PageSearch & {
  type?: "enrolled" | "followed" | "ignored"
  semester?: string
}
export type TeachersSearch = PageSearch & {
  department?: string
  title?: string
  q?: string
}
export type TeacherDetailSearch = PageSearch & {
  order_by?: "rating_score" | "rating_count"
}
export type UserAdminSearch = {
  tab?:
    | "user"
    | "admin"
    | "system-api-key"
    | "system-settings"
    | "announcement"
    | "audit-log"
  email?: string
  username?: string
  review_id?: number
  page?: number
  audit_page?: number
  audit_start_time?: string
  audit_end_time?: string
  audit_action?: string
  audit_actor_user_id?: number
}
export type SiteStatsSearch = { start_date?: string; end_date?: string }
export type CourseEnrollmentSyncCallbackSearch = {
  status?: "ok" | "error"
  semester?: string
  message?: string
  matched?: number
  total?: number
}

export function validateLoginSearch(search: SearchRecord): LoginSearch {
  return {
    redirect: rawStringParam(search.redirect)
      ? getSafeRedirectPath(rawStringParam(search.redirect))
      : undefined,
  }
}

export function validateCourseEnrollmentSyncCallbackSearch(
  search: SearchRecord
): CourseEnrollmentSyncCallbackSearch {
  return {
    status: enumParam(search.status, ["ok", "error"] as const) ?? "error",
    semester: rawStringParam(search.semester) ?? "",
    message: rawStringParam(search.message),
    matched: numberParam(search.matched) ?? 0,
    total: numberParam(search.total) ?? 0,
  }
}

export function validateCourseListSearch(
  search: SearchRecord
): CourseListSearch {
  return {
    q: stringParam(search.q),
    department: rawStringParam(search.department),
    language: rawStringParam(search.language),
    categories: stringArrayParam(search.categories),
    target_years: stringArrayParam(search.target_years),
    credit: numberParam(search.credit),
    order_by: enumParam(search.order_by, [
      "rating_score",
      "rating_count",
    ] as const),
    page: numberParam(search.page),
  }
}

export function validateCourseDetailSearch(
  search: SearchRecord
): CourseDetailSearch {
  return {
    page: numberParam(search.page),
    semester: rawStringParam(search.semester),
    rating: numberParam(search.rating),
    order_by: enumParam(search.order_by, [
      "created_at",
      "updated_at",
      "like_count",
    ] as const),
  }
}

export function validateReviewsSearch(search: SearchRecord): ReviewsSearch {
  return {
    q: stringParam(search.q),
    page: numberParam(search.page),
  }
}

export function validatePageSearch(search: SearchRecord): PageSearch {
  return { page: numberParam(search.page) }
}

export function validateUserCoursesSearch(
  search: SearchRecord
): UserCoursesSearch {
  return {
    type: enumParam(search.type, ["enrolled", "followed", "ignored"] as const),
    page: numberParam(search.page),
    semester: rawStringParam(search.semester),
  }
}

export function validateTeachersSearch(search: SearchRecord): TeachersSearch {
  return {
    department: rawStringParam(search.department),
    title: rawStringParam(search.title),
    q: stringParam(search.q),
    page: numberParam(search.page),
  }
}

export function validateTeacherDetailSearch(
  search: SearchRecord
): TeacherDetailSearch {
  return {
    page: numberParam(search.page),
    order_by: enumParam(search.order_by, [
      "rating_score",
      "rating_count",
    ] as const),
  }
}

export function validateUserAdminSearch(search: SearchRecord): UserAdminSearch {
  return {
    tab: enumParam(search.tab, [
      "user",
      "admin",
      "system-api-key",
      "system-settings",
      "announcement",
      "audit-log",
    ] as const),
    email: rawStringParam(search.email),
    username: rawStringParam(search.username),
    review_id: optionalPositiveIntParam(search.review_id),
    page: numberParam(search.page),
    audit_page: numberParam(search.audit_page),
    audit_start_time: rawStringParam(search.audit_start_time),
    audit_end_time: rawStringParam(search.audit_end_time),
    audit_action: rawStringParam(search.audit_action),
    audit_actor_user_id: optionalPositiveIntParam(search.audit_actor_user_id),
  }
}

export function validateSiteStatsSearch(search: SearchRecord): SiteStatsSearch {
  return {
    start_date: rawStringParam(search.start_date),
    end_date: rawStringParam(search.end_date),
  }
}
