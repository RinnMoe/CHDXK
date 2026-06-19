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
            <p className="font-medium">可能选错了课程</p>
            <p className="leading-6">
              当前课程最近开课于 {course.last_semester}。以下课程最近开课时间更新，可能是你想点评的课程。
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
                  主讲教师：{candidate.main_teacher.name}，最近开课：
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
                  去写点评
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
