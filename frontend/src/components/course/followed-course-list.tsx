import { CourseCompactCard } from "./course-compact-card"
import { Skeleton } from "@/components/ui/skeleton"
import { useAuth } from "@/contexts/auth-context"
import { useFollowedCourses } from "@/hooks/use-course"

interface FollowedCourseListProps {
  limit?: number
  skeletonCount?: number
}

function FollowedCourseSkeleton({ count }: { count: number }) {
  return (
    <div className="border-t">
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="flex items-center gap-4 border-b px-4 py-3">
          <div className="min-w-0 flex-1 space-y-2">
            <Skeleton className="h-3 w-24" />
            <Skeleton className="h-4 w-3/4" />
            <Skeleton className="h-3 w-1/2" />
          </div>
          <Skeleton className="h-4 w-12" />
        </div>
      ))}
    </div>
  )
}

export function FollowedCourseList({
  limit = 10,
  skeletonCount = 5,
}: FollowedCourseListProps) {
  const { user, isLoading: authLoading } = useAuth()
  const { data, isLoading } = useFollowedCourses(
    { page: 1, page_size: limit },
    !!user
  )

  if (authLoading || (!!user && isLoading)) {
    return <FollowedCourseSkeleton count={skeletonCount} />
  }

  if (!user) {
    return (
      <div className="py-12 text-center">
        <p className="text-sm text-muted-foreground">登录后查看关注课程</p>
      </div>
    )
  }

  const courses = data?.items ?? []
  if (courses.length === 0) {
    return (
      <div className="py-12 text-center">
        <p className="text-sm text-muted-foreground">暂无关注课程</p>
      </div>
    )
  }

  return (
    <div className="border-t">
      {courses.map((course) => (
        <CourseCompactCard key={course.id} course={course} />
      ))}
    </div>
  )
}
