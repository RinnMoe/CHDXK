import { copy } from "@/lib/copy"
import { Link } from "@tanstack/react-router"
import { RiAddLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { ReviewCard } from "@/components/review/review-card"
import { ReviewList } from "@/components/review/review-list"
import { PaginationComponent } from "@/components/common/pagination"
import { CourseReviewFilters } from "./course-review-filters"
import { CourseReviewTrendDialog } from "./course-review-trend-dialog"
import type {
  CourseDetailDTO,
  CourseReviewFilters as CourseReviewFiltersDTO,
} from "@/api/course"
import type { ReviewDTO } from "@/api/review"

type ReviewsPageData = {
  items: ReviewDTO[]
  total: number
  page: number
  page_size: number
}

interface CourseDetailReviewsSectionProps {
  course: CourseDetailDTO
  reviews?: ReviewsPageData
  listedReviews: ReviewDTO[]
  reviewsLoading: boolean
  reviewFilters?: CourseReviewFiltersDTO
  semester?: string
  rating?: number
  orderBy: "created_at" | "updated_at" | "like_count"
  onFilterChange: (next: Record<string, string | undefined>) => void
  onPageChange: (page: number) => void
}

export function CourseDetailReviewsSection({
  course,
  reviews,
  listedReviews,
  reviewsLoading,
  reviewFilters,
  semester,
  rating,
  orderBy,
  onFilterChange,
  onPageChange,
}: CourseDetailReviewsSectionProps) {
  return (
    <section>
      {course.my_review && (
        <div className="mb-6">
          <h2 className="mb-3 text-lg font-semibold">
            {copy(
              "copy.user.components.course.course-detail-reviews-section.477d1ac374fc"
            )}
          </h2>
          <div className="border-t">
            <ReviewCard review={course.my_review} />
          </div>
        </div>
      )}

      <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
        <div className="space-y-1">
          <h2 className="text-lg font-semibold">
            {copy(
              "copy.user.components.course.course-detail-reviews-section.f48b8c7146d4"
            )}
          </h2>
          {reviews && (
            <p className="text-sm text-muted-foreground">
              {copy(
                "copy.user.components.course.course-detail-reviews-section.2428420e407f"
              )}
              {reviews.total}
              {copy(
                "copy.user.components.course.course-detail-reviews-section.1c0a8a53fbb5"
              )}
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
              <Link
                to="/course/$courseID/review/new"
                params={{ courseID: String(course.id) }}
              >
                <RiAddLine data-icon="inline-start" />
                {copy(
                  "copy.user.components.course.course-detail-reviews-section.74cdbe78b5b2"
                )}
              </Link>
            </Button>
          )}
        </div>
      </div>

      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <CourseReviewFilters
          semesters={reviewFilters?.semesters}
          ratings={reviewFilters?.ratings}
          value={{ semester, rating, orderBy }}
          onChange={(next) => {
            const params: Record<string, string | undefined> = {}
            if ("semester" in next) params.semester = next.semester
            if ("rating" in next) {
              params.rating =
                next.rating === undefined ? undefined : String(next.rating)
            }
            if ("orderBy" in next) params.order_by = next.orderBy
            onFilterChange(params)
          }}
        />
      </div>
      <ReviewList
        reviews={listedReviews}
        isLoading={reviewsLoading}
        emptyText={
          course.my_review
            ? copy(
                "copy.user.components.course.course-detail-reviews-section.b81bd1d0511d"
              )
            : copy(
                "copy.user.components.course.course-detail-reviews-section.667c47ebc44b"
              )
        }
      />

      {reviews && reviews.total > 0 && (
        <div className="flex justify-center pt-4">
          <PaginationComponent
            page={reviews.page}
            pageSize={reviews.page_size}
            total={reviews.total}
            onPageChange={onPageChange}
          />
        </div>
      )}
    </section>
  )
}
