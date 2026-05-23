import { Link, useParams, useSearchParams } from "react-router-dom"
import { RiArrowLeftLine, RiAddLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { TitleBadge } from "@/components/ui/title-badge"
import {
  CourseCompactCard,
  SameCodeCourseCard,
} from "@/components/course/course-compact-card"
import { CourseBadge, CourseBadges } from "@/components/course/course-badges"
import { CourseNotificationControl } from "@/components/course/course-notification-control"
import { CourseReviewTrendDialog } from "@/components/course/course-review-trend-dialog"
import { RatingDistribution } from "@/components/course/rating-distribution"
import { PageShell } from "@/components/layout/page-shell"
import {
  useCourseDetail,
  useCourseReviewFilters,
  useCourseReviews,
} from "@/hooks/use-course"
import { ReviewList } from "@/components/review/review-list"
import { ReviewCard } from "@/components/review/review-card"
import { CourseReviewFilters } from "@/components/course/course-review-filters"
import { PaginationComponent } from "@/components/common/pagination"

const REVIEW_PAGE_SIZE = 10

function byRatingDesc(
  a: { rating: { avg: number } },
  b: { rating: { avg: number } }
) {
  return b.rating.avg - a.rating.avg
}

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

  const listedReviews =
    reviews?.items.filter((review) => review.id !== course.my_review?.id) ?? []

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

        <div className="space-y-6 lg:grid lg:grid-cols-3 lg:gap-6 lg:space-y-0">
          <div className="space-y-6 lg:col-span-2">
            <header className="space-y-3">
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <span className="font-mono">{course.code}</span>
                <span>·</span>
                <span>{course.department}</span>
              </div>
              <div className="space-y-2">
                <h1 className="text-3xl font-bold">{course.name}</h1>
                <div className="text-base">
                  主讲教师：
                  <Link
                    to={`/teachers/${course.main_teacher.id}`}
                    className="font-medium text-primary hover:underline"
                  >
                    {course.main_teacher.name}
                  </Link>
                  {course.main_teacher.title && (
                    <TitleBadge className="ml-1">
                      {course.main_teacher.title}
                    </TitleBadge>
                  )}
                </div>
              </div>
              <CourseBadges
                credit={course.credit}
                language={course.language}
                categories={course.categories}
                targetYears={course.target_years}
              />
            </header>

            <CourseNotificationControl
              courseID={course.id}
              level={course.notification_level}
            />

            {course.offered_courses.length > 0 && (
              <section className="space-y-3">
                <h2 className="text-lg font-semibold">历史开课</h2>
                <div className="space-y-3">
                  {course.offered_courses.map((oc) => (
                    <div
                      key={oc.semester}
                      className="flex items-center gap-3 text-sm"
                    >
                      <CourseBadge kind="targetYear" className="font-mono">
                        {oc.semester}
                      </CourseBadge>
                      <span className="text-muted-foreground">
                        {oc.teacher_group.map((t, index) => (
                          <span key={t.id}>
                            {index > 0 && <span className="mx-1">/</span>}
                            <Link
                              to={`/teachers/${t.id}`}
                              className="hover:text-primary hover:underline"
                            >
                              {t.name}
                            </Link>
                          </span>
                        ))}
                      </span>
                    </div>
                  ))}
                </div>
              </section>
            )}

            <Card>
              <CardContent className="py-6">
                <RatingDistribution rating={course.rating} />
              </CardContent>
            </Card>

            <section>
              {course.my_review && (
                <div className="mb-6">
                  <h2 className="mb-3 text-lg font-semibold">我的点评</h2>
                  <div className="border-t">
                    <ReviewCard
                      review={course.my_review}
                      showVoteButtons={false}
                    />
                  </div>
                </div>
              )}

              <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
                <div className="space-y-1">
                  <h2 className="text-lg font-semibold">课程点评</h2>
                  {reviews && (
                    <p className="text-sm text-muted-foreground">
                      共 {reviews.total} 条点评
                    </p>
                  )}
                </div>
                <div className="flex flex-wrap gap-2">
                  <CourseReviewTrendDialog
                    courseID={course.id}
                    courseName={course.name}
                  />
                  {!course.my_review && (
                    <Button asChild size="sm">
                      <Link to={`/courses/${course.id}/review/new`}>
                        <RiAddLine data-icon="inline-start" />
                        写点评
                      </Link>
                    </Button>
                  )}
                </div>
              </div>

              <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
                <CourseReviewFilters
                  semesters={reviewFilters?.semesters ?? []}
                  ratings={reviewFilters?.ratings ?? []}
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
                reviews={listedReviews}
                isLoading={reviewsLoading}
                emptyText={
                  course.my_review ? "还没有其他点评" : "还没有点评，来抢沙发？"
                }
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
          </div>

          {(course.same_code_courses.length > 0 ||
            course.same_teacher_courses.length > 0) && (
            <aside className="space-y-6">
              {course.same_code_courses.length > 0 && (
                <section>
                  <h2 className="mb-3 text-lg font-semibold">
                    其他老师的{course.name}
                  </h2>
                  <div className="border-t">
                    {[...course.same_code_courses]
                      .sort(byRatingDesc)
                      .map((c) => (
                        <SameCodeCourseCard key={c.id} course={c} />
                      ))}
                  </div>
                </section>
              )}

              {course.same_teacher_courses.length > 0 && (
                <section>
                  <h2 className="mb-3 text-lg font-semibold">
                    {course.main_teacher.name}的其他课
                  </h2>
                  <div className="border-t">
                    {[...course.same_teacher_courses]
                      .sort(byRatingDesc)
                      .map((c) => (
                        <CourseCompactCard key={c.id} course={c} />
                      ))}
                  </div>
                </section>
              )}
            </aside>
          )}
        </div>
      </PageShell>
    </>
  )
}
