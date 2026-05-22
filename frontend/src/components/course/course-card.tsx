import { Link } from "react-router-dom"
import { Badge } from "@/components/ui/badge"
import { RatingDisplay } from "./rating-display"
import type { CourseListItemDTO } from "@/api/course"

interface CourseCardProps {
  course: CourseListItemDTO
}

export function CourseCard({ course }: CourseCardProps) {
  return (
    <Link
      to={`/courses/${course.id}`}
      className="block border-b px-4 py-3 transition-colors hover:bg-muted/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
    >
      <div className="min-w-0 space-y-2">
        <div className="space-y-1">
          <div className="flex min-w-0 items-center justify-between gap-2 text-xs text-muted-foreground">
            <div className="flex min-w-0 items-center gap-2">
              <span className="shrink-0 font-mono">{course.code}</span>
              <span className="min-w-0 truncate text-sm">
                {course.main_teacher.name}
              </span>
            </div>
            <div className="shrink-0">
              <RatingDisplay rating={course.rating} size="sm" />
            </div>
          </div>
          <div className="whitespace-normal break-words font-semibold leading-tight">
            {course.name}
          </div>
          <div className="text-sm text-muted-foreground">
            {course.department}
          </div>
        </div>

        <div className="flex flex-wrap gap-1">
          <Badge
            variant="outline"
            className="border-emerald-200/50 bg-emerald-50/40 text-emerald-700/80 dark:border-emerald-900/40 dark:bg-emerald-950/25 dark:text-emerald-300/80"
          >
            {course.credit} 学分
          </Badge>
          <Badge
            variant="outline"
            className="border-sky-200/50 bg-sky-50/40 text-sky-700/80 dark:border-sky-900/40 dark:bg-sky-950/25 dark:text-sky-300/80"
          >
            {course.language}
          </Badge>
          {course.categories.slice(0, 2).map((cat) => (
            <Badge
              key={cat}
              variant="outline"
              className="border-amber-200/50 bg-amber-50/40 text-amber-800/75 dark:border-amber-900/40 dark:bg-amber-950/25 dark:text-amber-300/80"
            >
              {cat}
            </Badge>
          ))}
        </div>
      </div>
    </Link>
  )
}
