export const COURSE_ENROLLMENT_SYNC_CHANNEL = "course-enrollment-sync"

export type CourseEnrollmentSyncMessage = {
  status: "ok" | "error"
  semester: string
  message?: string
  matched?: number
  total?: number
}
