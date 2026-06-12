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
  const subject = `[JCourse课程信息反馈] ${course.code} ${course.name}`
  const courseURL =
    typeof window === "undefined"
      ? ""
      : `${window.location.origin}/course/${course.id}`
  const teacherNames =
    course.teacher_group && course.teacher_group.length > 0
      ? course.teacher_group.map((teacher) => teacher.name).join(" / ")
      : course.main_teacher.name
  const targetYears = course.target_years?.join("、") || "未提供"
  const categories = course.categories?.join("、") || "未提供"
  const body = [
    "请在这里描述需要反馈的问题：",
    "",
    "课程基本信息",
    `课程ID：${course.id}`,
    `课程代码：${course.code}`,
    `课程名称：${course.name}`,
    `院系：${course.department}`,
    `学分：${course.credit}`,
    `授课语言：${course.language}`,
    `面向对象：${targetYears}`,
    `课程分类：${categories}`,
    `最近学期：${course.last_semester}`,
    `主讲教师：${course.main_teacher.name}`,
    `合上教师：${teacherNames}`,
    courseURL ? `课程链接：${courseURL}` : undefined,
  ]
    .filter((line): line is string => Boolean(line))
    .join("\n")

  return `mailto:${brand.feedbackEmail}?${new URLSearchParams({
    subject,
    body,
  }).toString()}`
}
