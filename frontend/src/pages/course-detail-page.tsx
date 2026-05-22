import { Link, useParams } from "react-router-dom"
import { RiArrowLeftLine, RiAddLine } from "@remixicon/react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { Skeleton } from "@/components/ui/skeleton"
import { CourseCard } from "@/components/course/course-card"
import { RatingDistribution } from "@/components/course/rating-distribution"
import { PageShell } from "@/components/layout/page-shell"
import { useCourseDetail, useCourseReviews } from "@/hooks/use-course"
import { ReviewList } from "@/components/review/review-list"

export function CourseDetailPage() {
  const { courseID } = useParams<{ courseID: string }>()
  const id = Number(courseID)
  const { data: course, isLoading } = useCourseDetail(id)
  const { data: reviews } = useCourseReviews(id, { page: 1, page_size: 10 })

  if (isLoading) {
    return (
      <PageShell>
        <div className="space-y-4">
          <Skeleton className="h-4 w-24" />
          <Skeleton className="h-8 w-2/3" />
          <Skeleton className="h-32 w-full" />
        </div>
      </PageShell>
    )
  }

  if (!course) {
    return (
      <PageShell>
        <div className="py-16 text-center">
          <p className="text-muted-foreground">课程不存在</p>
          <Button asChild variant="link" className="mt-4">
            <Link to="/courses">返回课程列表</Link>
          </Button>
        </div>
      </PageShell>
    )
  }

  return (
    <PageShell>
      <Button asChild variant="ghost" size="sm" className="mb-4">
        <Link to="/courses">
          <RiArrowLeftLine data-icon="inline-start" />
          返回课程列表
        </Link>
      </Button>

      <div className="space-y-6">
        <header className="space-y-3">
          <div className="flex items-center gap-2 text-sm text-muted-foreground font-mono">
            <span>{course.code}</span>
            <span>·</span>
            <span>{course.department}</span>
          </div>
          <div className="flex items-start justify-between gap-4 flex-wrap">
            <h1 className="text-3xl font-bold">{course.name}</h1>
            <Badge variant="secondary">{course.credit} 学分</Badge>
          </div>
          <div className="flex flex-wrap gap-1.5">
            <Badge variant="outline">{course.language}</Badge>
            {course.categories.map((c) => (
              <Badge key={c} variant="outline">
                {c}
              </Badge>
            ))}
            {course.target_years.map((y) => (
              <Badge key={y} variant="outline">
                {y}
              </Badge>
            ))}
          </div>
          <div className="text-sm">
            主讲教师：
            <Link
              to={`/teachers/${course.main_teacher.id}`}
              className="text-primary hover:underline"
            >
              {course.main_teacher.name}
              {course.main_teacher.title && (
                <span className="ml-1 text-muted-foreground">
                  ({course.main_teacher.title})
                </span>
              )}
            </Link>
          </div>
        </header>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg">评分</CardTitle>
          </CardHeader>
          <CardContent>
            <RatingDistribution rating={course.rating} />
          </CardContent>
        </Card>

        {course.offered_courses.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">历史开课</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {course.offered_courses.map((oc) => (
                <div
                  key={oc.semester}
                  className="flex items-center gap-3 text-sm"
                >
                  <Badge variant="outline" className="font-mono">
                    {oc.semester}
                  </Badge>
                  <span className="text-muted-foreground">
                    {oc.teacher_group.map((t) => t.name).join(" / ")}
                  </span>
                </div>
              ))}
            </CardContent>
          </Card>
        )}

        <section>
          <div className="flex items-baseline justify-between mb-3">
            <h2 className="text-lg font-semibold">课程评价</h2>
            <div className="flex items-center gap-2">
              {reviews && (
                <span className="text-sm text-muted-foreground">
                  共 {reviews.total} 条
                </span>
              )}
              <Button asChild size="sm" variant="outline">
                <Link to={`/courses/${course.id}/review/new`}>
                  <RiAddLine data-icon="inline-start" />
                  写评价
                </Link>
              </Button>
            </div>
          </div>
          <ReviewList
            reviews={reviews?.items ?? []}
            isLoading={!reviews}
            emptyText="还没有评价，来抢沙发？"
          />
        </section>

        {course.same_code_courses.length > 0 && (
          <section>
            <h2 className="text-lg font-semibold mb-3">同代码课程</h2>
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {course.same_code_courses.map((c) => (
                <CourseCard key={c.id} course={c} />
              ))}
            </div>
          </section>
        )}

        {course.same_teacher_courses.length > 0 && (
          <section>
            <h2 className="text-lg font-semibold mb-3">同教师其他课程</h2>
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {course.same_teacher_courses.map((c) => (
                <CourseCard key={c.id} course={c} />
              ))}
            </div>
          </section>
        )}

        <Separator />
      </div>
    </PageShell>
  )
}
