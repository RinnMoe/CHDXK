import { Link, useParams, useSearchParams } from "react-router-dom"
import { RiArrowLeftLine, RiAddLine } from "@remixicon/react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { Skeleton } from "@/components/ui/skeleton"
import { CourseCard } from "@/components/course/course-card"
import { RatingDistribution } from "@/components/course/rating-distribution"
import { PageShell } from "@/components/layout/page-shell"
import {
  useCourseDetail,
  useCourseReviewFilters,
  useCourseReviews,
} from "@/hooks/use-course"
import { ReviewList } from "@/components/review/review-list"
import { CourseReviewFilters } from "@/components/course/course-review-filters"
import { PaginationComponent } from "@/components/common/pagination"

const REVIEW_PAGE_SIZE = 10

export function CourseDetailPage() {
  const { courseID } = useParams<{ courseID: string }>()
  const id = Number(courseID)
  const [searchParams, setSearchParams] = useSearchParams()
  const reviewPage = Math.max(1, Number(searchParams.get("page") ?? "1") || 1)
  const semester = searchParams.get("semester") ?? undefined
  const ratingParam = searchParams.get("rating")
  const parsedRating = ratingParam ? Number(ratingParam) : undefined
  const rating =
    parsedRating && parsedRating >= 1 && parsedRating <= 5
      ? parsedRating
      : undefined
  const orderBy =
    searchParams.get("order_by") === "like_count" ? "like_count" : "created_at"

  const { data: course, isLoading } = useCourseDetail(id)
  const { data: reviewFilters } = useCourseReviewFilters(id)
  const reviewFilterTotal =
    reviewFilters?.ratings.reduce((sum, item) => sum + item.count, 0) ??
    course?.rating.count ??
    0
  const { data: reviews, isLoading: reviewsLoading } = useCourseReviews(id, {
    semester,
    rating,
    order_by: orderBy,
    page: reviewPage,
    page_size: REVIEW_PAGE_SIZE,
  })

  function updateReviewParams(next: Record<string, string | undefined>) {
    const params = new URLSearchParams(searchParams)
    for (const [key, value] of Object.entries(next)) {
      if (!value) params.delete(key)
      else params.set(key, value)
    }
    params.set("page", "1")
    setSearchParams(params)
  }

  function handlePageChange(page: number) {
    const params = new URLSearchParams(searchParams)
    params.set("page", String(page))
    setSearchParams(params)
  }

  if (isLoading) {
    return (
      <>
        <title>课程 - JCourse</title>
        <PageShell>
          <div className="space-y-4">
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-8 w-2/3" />
            <Skeleton className="h-32 w-full" />
          </div>
        </PageShell>
      </>
    )
  }

  if (!course) {
    return (
      <>
        <title>课程 - JCourse</title>
        <PageShell>
          <div className="py-16 text-center">
            <p className="text-muted-foreground">课程不存在</p>
            <Button asChild variant="link" className="mt-4">
              <Link to="/courses">返回课程列表</Link>
            </Button>
          </div>
        </PageShell>
      </>
    )
  }

  return (
    <>
      <title>{course.name} - JCourse</title>
      <PageShell>
        <Button asChild variant="ghost" size="sm" className="mb-4">
          <Link to="/courses">
            <RiArrowLeftLine data-icon="inline-start" />
            返回课程列表
          </Link>
        </Button>

        <div className="space-y-6">
          <header className="space-y-3">
            <div className="flex items-center gap-2 font-mono text-sm text-muted-foreground">
              <span>{course.code}</span>
              <span>·</span>
              <span>{course.department}</span>
            </div>
            <div className="flex flex-wrap items-start justify-between gap-4">
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
            <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
              <div className="space-y-1">
                <h2 className="text-lg font-semibold">课程评价</h2>
                {reviews && (
                  <p className="text-sm text-muted-foreground">
                    共 {reviews.total} 条评价
                  </p>
                )}
              </div>
              <Button asChild size="sm" variant="outline">
                <Link to={`/courses/${course.id}/review/new`}>
                  <RiAddLine data-icon="inline-start" />
                  写评价
                </Link>
              </Button>
            </div>

            <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
              <CourseReviewFilters
                semesters={reviewFilters?.semesters ?? []}
                ratings={reviewFilters?.ratings ?? []}
                total={reviewFilterTotal}
                value={{
                  semester,
                  rating,
                  orderBy,
                }}
                onChange={(next) => {
                  const params: Record<string, string | undefined> = {}
                  if ("semester" in next) params.semester = next.semester
                  if ("rating" in next) {
                    params.rating =
                      next.rating === undefined
                        ? undefined
                        : String(next.rating)
                  }
                  if ("orderBy" in next) params.order_by = next.orderBy
                  updateReviewParams(params)
                }}
              />
            </div>
            <ReviewList
              reviews={reviews?.items ?? []}
              isLoading={reviewsLoading}
              emptyText="还没有评价，来抢沙发？"
            />

            {reviews && reviews.total > 0 && (
              <div className="flex justify-center pt-4">
                <PaginationComponent
                  page={reviews.page}
                  pageSize={reviews.page_size}
                  total={reviews.total}
                  onPageChange={handlePageChange}
                />
              </div>
            )}
          </section>

          {course.same_code_courses.length > 0 && (
            <section>
              <h2 className="mb-3 text-lg font-semibold">同代码课程</h2>
              <div className="border-t">
                {course.same_code_courses.map((c) => (
                  <CourseCard key={c.id} course={c} />
                ))}
              </div>
            </section>
          )}

          {course.same_teacher_courses.length > 0 && (
            <section>
              <h2 className="mb-3 text-lg font-semibold">同教师其他课程</h2>
              <div className="border-t">
                {course.same_teacher_courses.map((c) => (
                  <CourseCard key={c.id} course={c} />
                ))}
              </div>
            </section>
          )}

          <Separator />
        </div>
      </PageShell>
    </>
  )
}
