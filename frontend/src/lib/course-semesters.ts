import type { CourseDetailDTO } from "@/api/course"

type CourseSemesterSource = Pick<
  CourseDetailDTO,
  "last_semester" | "offered_courses"
>

function sortDesc(a: string, b: string) {
  if (a < b) return 1
  if (a > b) return -1
  return 0
}

export function getCourseSemesters(course?: CourseSemesterSource | null) {
  const semesters = new Set<string>()
  const offeredCourses = course?.offered_courses ?? []

  if (offeredCourses.length > 0) {
    for (const offeredCourse of offeredCourses) {
      if (offeredCourse.semester) semesters.add(offeredCourse.semester)
    }
  } else if (course?.last_semester) {
    semesters.add(course.last_semester)
  }

  return [...semesters].sort(sortDesc)
}

export function getDefaultSemester(
  semesters: string[],
  currentSemester?: string | null
) {
  if (currentSemester && semesters.includes(currentSemester)) {
    return currentSemester
  }
  return semesters[0] ?? ""
}
