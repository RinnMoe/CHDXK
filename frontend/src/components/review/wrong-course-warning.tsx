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

  const candidates = findUpdatedSameTeacherCourses(course, currentSemester)
  const updatedCourse = candidates[0]
  if (!updatedCourse) return null

  const extraCount = candidates.length - 1

  return (
    <div className="rounded-md border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-950 dark:border-amber-900/60 dark:bg-amber-950/30 dark:text-amber-100">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex min-w-0 gap-3">
          <RiErrorWarningLine className="mt-0.5 size-4 shrink-0 text-amber-700 dark:text-amber-300" />
          <div className="min-w-0 space-y-1">
            <p className="font-medium">可能选错了课程</p>
            <p className="leading-6">
              当前课程最近开课于 {course.last_semester}。同一教师下有更新的同名课程
              <span className="font-mono"> {updatedCourse.code}</span>，最近开课于{" "}
              {updatedCourse.last_semester}
              {extraCount > 0 ? `，另有 ${extraCount} 门候选课程` : ""}。
            </p>
          </div>
        </div>
        <Button asChild size="sm" className="self-start whitespace-nowrap">
          <Link
            to="/course/$courseID/review/new"
            params={{ courseID: String(updatedCourse.id) }}
          >
            去新课程写点评
            <RiArrowRightLine data-icon="inline-end" />
          </Link>
        </Button>
      </div>
    </div>
  )
}
