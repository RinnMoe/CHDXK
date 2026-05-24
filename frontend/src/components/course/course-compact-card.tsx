import { Link } from "react-router-dom"
import { RatingDisplay } from "./rating-display"
import { TitleBadge } from "@/components/ui/title-badge"
import type { CourseListItemDTO } from "@/api/course"

interface CourseCompactCardProps {
  course: CourseListItemDTO
}

export function CourseCompactCard({ course }: CourseCompactCardProps) {
  return (
    <Link
      to={`/course/${course.id}`}
      className="flex items-center gap-4 border-b px-4 py-3 transition-colors hover:bg-muted/40 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
    >
      <div className="min-w-0 flex-1 space-y-2">
        <div className="flex min-w-0 items-center gap-2 text-sm text-muted-foreground">
          <span className="shrink-0 font-mono">{course.code}</span>
          <span className="min-w-0 text-sm break-words">
            {course.main_teacher.name}
          </span>
        </div>
        <div className="leading-tight font-semibold break-words whitespace-normal">
          {course.name}
        </div>
        <div className="text-sm text-muted-foreground">{course.department}</div>
      </div>
      <div className="shrink-0">
        <RatingDisplay rating={course.rating} size="sm" />
      </div>
    </Link>
  )
}

interface SameCodeCourseCardProps {
  course: CourseListItemDTO
}

export function SameCodeCourseCard({ course }: SameCodeCourseCardProps) {
  return (
    <Link
      to={`/course/${course.id}`}
      className="block border-b px-4 py-3 transition-colors hover:bg-muted/40 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
    >
      <div className="flex min-w-0 items-center gap-3">
        <div className="min-w-0 flex-1 space-y-1">
          <div className="flex items-center gap-1.5">
            <span className="shrink-0 font-mono text-sm text-muted-foreground">
              {course.main_teacher.code}
            </span>
            {course.main_teacher.title && (
              <TitleBadge>{course.main_teacher.title}</TitleBadge>
            )}
          </div>
          <div className="leading-tight font-semibold">
            {course.main_teacher.name}
          </div>
          <div className="text-sm text-muted-foreground">
            {course.department}
          </div>
        </div>
        <div className="shrink-0">
          <RatingDisplay rating={course.rating} size="sm" />
        </div>
      </div>
    </Link>
  )
}
