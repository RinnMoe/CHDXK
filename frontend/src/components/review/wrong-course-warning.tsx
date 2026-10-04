import { copy } from "@/lib/copy"
import { Link } from "@tanstack/react-router"
import { RiArrowRightLine, RiErrorWarningLine } from "@remixicon/react"

import { Button } from "@/components/ui/button"
import type { CourseDetailDTO, CourseListItemDTO } from "@/api/course"

interface WrongCourseWarningProps {
  course?: CourseDetailDTO | null
  currentSemester?: string | null
}

function findUpdatedSameTeacherCourses(
  course: CourseDetailDTO,
  currentSemester: string
) {
  if (!course.last_semester || course.last_semester >= currentSemester) {
    return []
  }

  return course.same_teacher_courses
    .filter(
      (candidate) =>
        candidate.name === course.name &&
        candidate.code !== course.code &&
        candidate.last_semester > course.last_semester
    )
    .sort(byLastSemesterDesc)
}

function findUpdatedSameCodeCourses(
  course: CourseDetailDTO,
  currentSemester: string
) {
  if (!course.last_semester || course.last_semester >= currentSemester) {
    return []
  }

  return course.same_code_courses
    .filter(
      (candidate) =>
        candidate.main_teacher.id !== course.main_teacher.id &&
        candidate.last_semester > course.last_semester
    )
    .sort(byLastSemesterDesc)
}

function findUpdatedCourseCandidates(
  course: CourseDetailDTO,
  currentSemester: string
) {
  return [
    ...findUpdatedSameTeacherCourses(course, currentSemester),
    ...findUpdatedSameCodeCourses(course, currentSemester),
  ].sort(byLastSemesterDesc)
}

function byLastSemesterDesc(a: CourseListItemDTO, b: CourseListItemDTO) {
  if (a.last_semester < b.last_semester) return 1
  if (a.last_semester > b.last_semester) return -1
  return a.code.localeCompare(b.code)
}

export function WrongCourseWarning({
  course,
  currentSemester,
}: WrongCourseWarningProps) {
  if (!course || !currentSemester) return null

  const candidates = findUpdatedCourseCandidates(course, currentSemester)
  if (candidates.length === 0) return null

  return (
    <div className="rounded-md border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-950 dark:border-amber-900/60 dark:bg-amber-950/30 dark:text-amber-100">
      <div className="space-y-3">
        <div className="flex min-w-0 gap-3">
          <RiErrorWarningLine className="mt-0.5 size-4 shrink-0 text-amber-700 dark:text-amber-300" />
          <div className="min-w-0 space-y-1">
            <p className="font-medium">
              {copy(
                "copy.user.components.review.wrong-course-warning.13bd2ee75a52"
              )}
            </p>
            <p className="leading-6">
              {copy(
                "copy.user.components.review.wrong-course-warning.fae705e0a006"
              )}
              {course.last_semester}
              {copy(
                "copy.user.components.review.wrong-course-warning.fc27f265f0ea"
              )}
            </p>
          </div>
        </div>
        <div className="divide-y divide-amber-200/80 border-t border-amber-200/80 pl-7 dark:divide-amber-900/60 dark:border-amber-900/60">
          {candidates.map((candidate) => (
            <div
              key={candidate.id}
              className="flex flex-col gap-2 py-2.5 sm:flex-row sm:items-center sm:justify-between"
            >
              <div className="min-w-0 space-y-1">
                <div className="flex min-w-0 flex-wrap items-baseline gap-x-2 gap-y-1">
                  <span className="font-mono text-xs text-amber-800 dark:text-amber-200">
                    {candidate.code}
                  </span>
                  <span className="font-medium wrap-break-word">
                    {candidate.name}
                  </span>
                </div>
                <div className="text-xs text-amber-900/80 dark:text-amber-100/80">
                  {copy(
                    "copy.user.components.review.wrong-course-warning.9ae8dcbe2d0d"
                  )}
                  {candidate.main_teacher.name}
                  {copy(
                    "copy.user.components.review.wrong-course-warning.1bc395f9109b"
                  )}
                  {candidate.last_semester}
                </div>
              </div>
              <Button
                asChild
                size="sm"
                variant="outline"
                className="self-start whitespace-nowrap"
              >
                <Link
                  to="/course/$courseID/review/new"
                  params={{ courseID: String(candidate.id) }}
                >
                  {copy(
                    "copy.user.components.review.wrong-course-warning.a295de460aba"
                  )}
                  <RiArrowRightLine data-icon="inline-end" />
                </Link>
              </Button>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
