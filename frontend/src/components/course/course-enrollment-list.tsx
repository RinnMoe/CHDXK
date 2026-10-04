import { copy } from "@/lib/copy"
import { Link } from "@tanstack/react-router"
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
        <p className="text-muted-foreground">
          {copy(
            "copy.user.components.course.course-enrollment-list.0cbe09090cfb"
          )}
        </p>
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
            className="flex items-center gap-3 border-b px-4 py-3 transition-colors hover:bg-muted/40"
          >
            <Link
              to="/course/$courseID"
              params={{ courseID: String(course.id) }}
              className="min-w-0 flex-1 space-y-2 rounded-sm focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
            >
              <div className="flex min-w-0 items-center gap-2 text-sm text-muted-foreground">
                <span className="shrink-0 font-mono">{course.code}</span>
                <span className="min-w-0 wrap-break-word">
                  {course.main_teacher.name}
                </span>
              </div>
              <div className="leading-tight font-semibold wrap-break-word whitespace-normal">
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
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    className="hover:bg-destructive/10 hover:text-destructive focus-visible:border-destructive/40 focus-visible:ring-destructive/20 dark:hover:bg-destructive/20"
                    aria-label={copy(
                      "copy.user.components.course.course-enrollment-list.8d7021060db4"
                    )}
                  >
                    <RiDeleteBinLine />
                  </Button>
                </AlertDialogTrigger>
                <AlertDialogContent size="sm">
                  <AlertDialogHeader>
                    <AlertDialogTitle>
                      {copy(
                        "copy.user.components.course.course-enrollment-list.8d7021060db4"
                      )}
                    </AlertDialogTitle>
                    <AlertDialogDescription>
                      {copy(
                        "copy.user.components.course.course-enrollment-list.b5e1bb61b32e"
                      )}
                      {course.name} {enrollment.semester}
                      {copy(
                        "copy.user.components.course.course-enrollment-list.a089fb8b0266"
                      )}
                    </AlertDialogDescription>
                  </AlertDialogHeader>
                  <AlertDialogFooter>
                    <AlertDialogCancel>
                      {copy(
                        "copy.user.components.course.course-enrollment-list.2cd0f3be8738"
                      )}
                    </AlertDialogCancel>
                    <AlertDialogAction
                      variant="destructive"
                      onClick={() => deleteMutation.mutate(enrollment.id)}
                    >
                      {copy(
                        "copy.user.components.course.course-enrollment-list.2f9daa828907"
                      )}
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
