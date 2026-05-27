import { CourseCompactCard } from "./course-compact-card"
import type { HotCourseListDTO } from "@/api/course"
import { Skeleton } from "@/components/ui/skeleton"
import { useHotCourses } from "@/hooks/use-course"

interface HotCourseListProps {
  period: "week" | "month"
  limit: number
  skeletonCount?: number
  data?: HotCourseListDTO
  isLoading?: boolean
  fetchData?: boolean
}

function HotCourseHeader() {
  return (
    <div className="flex items-center gap-2 border-b px-4 py-2 text-sm font-medium text-muted-foreground">
      <span className="w-8 shrink-0 whitespace-nowrap text-center">排名</span>
      <span className="min-w-0 flex-1 pl-4">课程</span>
      <span className="w-12 shrink-0 text-right">热度</span>
    </div>
  )
}

function HotCourseSkeleton({ count }: { count: number }) {
  return (
    <div>
      <HotCourseHeader />
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="flex items-center gap-2 border-b px-4 py-3">
          <Skeleton className="h-6 w-8 shrink-0" />
          <div className="min-w-0 flex-1 space-y-2">
            <Skeleton className="h-3 w-24" />
            <Skeleton className="h-4 w-3/4" />
            <Skeleton className="h-3 w-1/2" />
          </div>
          <Skeleton className="h-4 w-12" />
          <Skeleton className="h-4 w-8" />
        </div>
      ))}
    </div>
  )
}

export function HotCourseList({
  period,
  limit,
  skeletonCount = 5,
  data: providedData,
  isLoading: providedIsLoading,
  fetchData = true,
}: HotCourseListProps) {
  const query = useHotCourses(period, limit, fetchData)
  const data = providedData ?? query.data
  const isLoading = providedIsLoading ?? query.isLoading

  if (isLoading) return <HotCourseSkeleton count={skeletonCount} />

  const items = data?.items ?? []
  if (items.length === 0) {
    return (
      <div className="py-12 text-center">
        <p className="text-sm text-muted-foreground">暂无热门课程</p>
      </div>
    )
  }

  return (
    <div>
      <HotCourseHeader />
      {items.map((item, i) => (
        <div key={item.course.id} className="flex items-center gap-2 border-b">
          <span
            className={`ml-4 w-8 shrink-0 text-center text-sm font-bold ${
              i < 3 ? "text-amber-500" : "text-muted-foreground"
            }`}
          >
            {i + 1}
          </span>
          <div className="min-w-0 flex-1">
            <CourseCompactCard course={item.course} />
          </div>
          <div className="w-12 shrink-0 pr-4 text-right text-sm font-medium text-muted-foreground">
            {item.score}
          </div>
        </div>
      ))}
    </div>
  )
}
