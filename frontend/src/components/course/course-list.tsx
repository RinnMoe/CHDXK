import { CourseCard } from "./course-card"
import { Skeleton } from "@/components/ui/skeleton"
import type { CourseListItemDTO } from "@/api/course"

interface CourseListProps {
  courses: CourseListItemDTO[]
  isLoading?: boolean
}

function CourseCardSkeleton() {
  return (
    <div className="space-y-2 border-b px-4 py-3">
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <Skeleton className="h-3 w-16" />
          <Skeleton className="h-4 w-20" />
        </div>
        <Skeleton className="h-4 w-28" />
      </div>
      <Skeleton className="h-4 w-3/4" />
      <Skeleton className="h-4 w-40" />
      <div className="flex gap-1">
        <Skeleton className="h-5 w-12" />
        <Skeleton className="h-5 w-16" />
        <Skeleton className="h-5 w-12" />
      </div>
      <Skeleton className="h-3 w-24" />
    </div>
  )
}

export function CourseList({ courses, isLoading }: CourseListProps) {
  if (isLoading) {
    return (
      <div className="border-t">
        {Array.from({ length: 6 }).map((_, i) => (
          <CourseCardSkeleton key={i} />
        ))}
      </div>
    )
  }

  if (courses.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <div className="text-4xl mb-4">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-12 w-12 text-muted-foreground/50 mx-auto"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"
            />
          </svg>
        </div>
        <p className="text-muted-foreground">暂无课程</p>
        <p className="text-sm text-muted-foreground/70 mt-1">
          尝试调整筛选条件
        </p>
      </div>
    )
  }

  return (
    <div className="border-t">
      {courses.map((course) => (
        <CourseCard key={course.id} course={course} />
      ))}
    </div>
  )
}
