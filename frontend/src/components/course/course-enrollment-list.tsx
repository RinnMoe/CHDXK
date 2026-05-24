import { Link } from "react-router-dom"
import { RiDeleteBinLine } from "@remixicon/react"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { CourseBadges, CourseSemesterBadge } from "./course-badges"
import { RatingDisplay } from "./rating-display"
import { useDeleteCourseEnrollment } from "@/hooks/use-course"
import type { CourseEnrollmentDTO } from "@/api/course"

interface CourseEnrollmentListProps {
  enrollments: CourseEnrollmentDTO[]
  isLoading?: boolean
}

function CourseEnrollmentSkeleton() {
  return (
    <div className="space-y-2 border-b px-4 py-3">
      <Skeleton className="h-4 w-32" />
      <Skeleton className="h-5 w-2/3" />
      <Skeleton className="h-4 w-44" />
    </div>
  )
}

export function CourseEnrollmentList({
  enrollments,
  isLoading,
}: CourseEnrollmentListProps) {
  const deleteMutation = useDeleteCourseEnrollment()

  if (isLoading) {
    return (
      <div className="border-t">
        {Array.from({ length: 6 }).map((_, i) => (
          <CourseEnrollmentSkeleton key={i} />
        ))}
      </div>
    )
  }

  if (enrollments.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <p className="text-muted-foreground">暂无选课记录</p>
      </div>
    )
  }

  return (
    <div className="border-t">
      {enrollments.map((enrollment) => {
        const course = enrollment.course
        return (
          <div
            key={enrollment.id}
            className="flex items-center gap-3 border-b px-4 py-3"
          >
            <Link
              to={`/course/${course.id}`}
              className="min-w-0 flex-1 space-y-2 rounded-sm focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
            >
              <div className="flex min-w-0 items-center gap-2 text-sm text-muted-foreground">
                <span className="shrink-0 font-mono">{course.code}</span>
                <span className="min-w-0 break-words">
                  {course.main_teacher.name}
                </span>
              </div>
              <div className="leading-tight font-semibold break-words whitespace-normal">
                {course.name}
              </div>
              <div className="text-sm text-muted-foreground">
                {course.department}
              </div>
              <div className="flex flex-wrap items-center gap-1">
                <CourseSemesterBadge semester={enrollment.semester} />
                <CourseBadges
                  credit={course.credit}
                  language={course.language}
                  categories={course.categories}
                  categoryLimit={2}
                  className="gap-1"
                />
              </div>
            </Link>
            <div className="shrink-0 self-center">
              <RatingDisplay rating={course.rating} size="sm" />
            </div>
            <div className="flex shrink-0 flex-col items-end gap-2 self-center">
              <AlertDialog>
                <AlertDialogTrigger asChild>
                  <Button variant="outline" size="sm">
                    <RiDeleteBinLine />
                    删除
                  </Button>
                </AlertDialogTrigger>
                <AlertDialogContent size="sm">
                  <AlertDialogHeader>
                    <AlertDialogTitle>删除选课记录</AlertDialogTitle>
                    <AlertDialogDescription>
                      删除 {course.name} {enrollment.semester} 的选课记录。
                    </AlertDialogDescription>
                  </AlertDialogHeader>
                  <AlertDialogFooter>
                    <AlertDialogCancel>取消</AlertDialogCancel>
                    <AlertDialogAction
                      variant="destructive"
                      onClick={() => deleteMutation.mutate(enrollment.id)}
                    >
                      删除
                    </AlertDialogAction>
                  </AlertDialogFooter>
                </AlertDialogContent>
              </AlertDialog>
            </div>
          </div>
        )
      })}
    </div>
  )
}
