export const SCHEMA_VERSION = 1 as const

export interface SemesterOption {
  id: string
  academicYear: string
  term: string
  label: string
  normalized: string
}

export interface LessonSummary {
  lessonNo: string
  courseCode: string
  courseName: string
  courseType: string
  teachClass: string
  teachers: string[]
  actualCount: number | null
  limitCount: number | null
  credit: number | null
  coursePeriod: string
  weekRange: string
  raw: Record<string, string>
}

export interface RestrictionItem {
  name: string
  operator: string
  value: string
  raw: string
}

export interface RestrictionGroup {
  title: string
  items: RestrictionItem[]
  raw: string
}

export interface LessonDetail {
  basic: {
    semester: string
    courseCode: string
    courseName: string
    credit: number | null
    courseType: string
    department: string
    campus: string
    language: string
    assessment: string
    teachers: string[]
  }
  arrangement: {
    totalHours: number | null
    weeklyHours: number | null
    weekRange: string
    startWeek: number | null
    endWeek: number | null
    weekCount: number | null
  }
  restrictions: RestrictionGroup[]
  remark: string
  raw: Record<string, string>
}

export interface LessonRecord {
  id: string
  summary: LessonSummary
  detail: LessonDetail | null
  links: {
    detail: string
    syllabus: string
  }
  warnings: string[]
}

export interface SemesterSource {
  host: string
  path: string
  projectId: string
  projectLabel: string
  semesterId: string
  label: string
}

export interface SemesterDocument {
  schemaVersion: typeof SCHEMA_VERSION
  source: SemesterSource
  semester: string
  fetchedAt: string
  expectedCount: number | null
  lessons: LessonRecord[]
}

export interface FailedLesson {
  lessonId: string
  error: string
  attempts: number
  updatedAt: string
}

export interface Checkpoint {
  schemaVersion: typeof SCHEMA_VERSION
  semester: string
  semesterId: string
  expectedCount: number | null
  listComplete: boolean
  lastPage: number
  pageSize: number
  completedLessonIds: string[]
  retries: Record<string, number>
  failed: FailedLesson[]
  updatedAt: string
}

export interface Report {
  schemaVersion: typeof SCHEMA_VERSION
  semester: string
  semesterId: string
  startedAt: string
  updatedAt: string
  completed: boolean
  listComplete: boolean
  expectedCount: number | null
  lessonCount: number
  detailSuccessCount: number
  missingDetailCount: number
  duplicateLessonIds: string[]
  failedLessonIds: string[]
  warnings: string[]
  sourcePage: {
    host: string
    path: string
    gridSignature: string | null
  }
}

export interface ListingPage {
  pageNo: number
  pageSize: number
  maxPageNo: number
  expectedCount: number
  gridSignature: string
  lessons: LessonRecord[]
  availablePageSizes: number[]
}

export class BkjwError extends Error {
  constructor(message: string, readonly code: string) {
    super(message)
    this.name = "BkjwError"
  }
}

export class AuthenticationRequiredError extends BkjwError {
  constructor(message = "教务系统登录状态已失效，请运行 login 完成登录") {
    super(message, "AUTH_REQUIRED")
    this.name = "AuthenticationRequiredError"
  }
}

export class PageStructureError extends BkjwError {
  constructor(message: string) {
    super(message, "PAGE_STRUCTURE")
    this.name = "PageStructureError"
  }
}

export class InterruptedError extends BkjwError {
  constructor(message = "采集被中断，已保存断点") {
    super(message, "INTERRUPTED")
    this.name = "InterruptedError"
  }
}
