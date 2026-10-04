import { copy } from "@/lib/copy"
import type { CourseDetailDTO } from "@/api/course"
import { brand } from "@/config/brand"

export const REVIEW_PAGE_SIZE = 10

export function byRatingDesc(
  a: { rating: { score: number } },
  b: { rating: { score: number } }
) {
  return b.rating.score - a.rating.score
}

export function buildFeedbackMailto(course: CourseDetailDTO) {
  if (!brand.feedbackEmail) return undefined

  const subject = copy(
    "copy.user.components.course.course-detail-utils.d586b094407c",
    { v0: course.code, v1: course.name }
  )
  const courseURL =
    typeof window === "undefined"
      ? ""
      : `${window.location.origin}/course/${course.id}`
  const teacherNames =
    course.teacher_group && course.teacher_group.length > 0
      ? course.teacher_group.map((teacher) => teacher.name).join(" / ")
      : course.main_teacher.name
  const targetYears =
    course.target_years?.join("、") ||
    copy("copy.user.components.course.course-detail-utils.756762e293f2")
  const categories =
    course.categories?.join("、") ||
    copy("copy.user.components.course.course-detail-utils.756762e293f2")
  const body = [
    copy("copy.user.components.course.course-detail-utils.97fba4ee4556"),
    "\n",
    copy("copy.user.components.course.course-detail-utils.6a89b64321c6"),
    copy("copy.user.components.course.course-detail-utils.77a97685633a", {
      v0: course.id,
    }),
    copy("copy.user.components.course.course-detail-utils.d52fc7f53e59", {
      v0: course.code,
    }),
    copy("copy.user.components.course.course-detail-utils.da647969ca34", {
      v0: course.name,
    }),
    copy("copy.user.components.course.course-detail-utils.040e06692483", {
      v0: course.department,
    }),
    copy("copy.user.components.course.course-detail-utils.a7aee2503dce", {
      v0: course.credit,
    }),
    copy("copy.user.components.course.course-detail-utils.f41b1401daa2", {
      v0: course.language,
    }),
    copy("copy.user.components.course.course-detail-utils.d07d1f8b938d", {
      v0: targetYears,
    }),
    copy("copy.user.components.course.course-detail-utils.1f3043d9cb99", {
      v0: categories,
    }),
    copy("copy.user.components.course.course-detail-utils.63296cdae076", {
      v0: course.last_semester,
    }),
    copy("copy.user.components.course.course-detail-utils.a437ba1c813f", {
      v0: course.main_teacher.name,
    }),
    copy("copy.user.components.course.course-detail-utils.01c108d51b41", {
      v0: teacherNames,
    }),
    courseURL
      ? copy("copy.user.components.course.course-detail-utils.60d2204afbb8", {
          v0: courseURL,
        })
      : undefined,
  ]
    .filter((line): line is string => Boolean(line))
    .join("\n")

  return `mailto:${brand.feedbackEmail}?${new URLSearchParams({
    subject,
    body,
  }).toString()}`
}
