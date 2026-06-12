import { Link } from "@tanstack/react-router"
import type { ReactNode } from "react"
import { RatingDisplay } from "./rating-display"
import { CourseBadges } from "./course-badges"
import type { CourseListItemDTO } from "@/api/course"

interface CourseCardProps {
  course: CourseListItemDTO
  action?: ReactNode
}

function CourseCardContent({ course }: { course: CourseListItemDTO }) {
  return (
    <div className="flex min-w-0 items-center gap-2">
      <div className="min-w-0 flex-1 space-y-2">
        <div className="flex min-w-0 items-center gap-2 text-sm text-muted-foreground">
          <span className="shrink-0 font-mono">{course.code}</span>
          <span className="min-w-0 text-sm wrap-break-word">
            {course.main_teacher.name}
          </span>
        </div>
        <div className="leading-tight font-semibold wrap-break-word whitespace-normal">
          {course.name}
        </div>
        <div className="text-sm text-muted-foreground">{course.department}</div>

        <CourseBadges
          credit={course.credit}
          language={course.language}
          categories={course.categories}
          categoryLimit={2}
          className="gap-1"
        />
      </div>
      <div className="shrink-0">
        <RatingDisplay rating={course.rating} size="sm" />
      </div>
    </div>
  )
}

export function CourseCard({ course, action }: CourseCardProps) {
  if (action) {
    return (
      <div className="flex items-center gap-3 border-b p-2">
        <Link
          to="/course/$courseID"
          params={{ courseID: String(course.id) }}
          className="min-w-0 flex-1 rounded-sm transition-colors hover:bg-muted/40 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
        >
          <CourseCardContent course={course} />
        </Link>
        <div className="flex shrink-0 flex-col items-end gap-2 self-center">
          {action}
        </div>
      </div>
    )
  }

  return (
    <Link
      to="/course/$courseID"
      params={{ courseID: String(course.id) }}
      className="block border-b p-2 transition-colors hover:bg-muted/40 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
    >
      <CourseCardContent course={course} />
    </Link>
  )
}
