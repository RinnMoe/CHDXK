import { useParams, useNavigate, Link } from "react-router-dom"
import { RiArrowLeftLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { CourseHeaderMeta } from "@/components/course/course-header-meta"
import { ReviewForm } from "@/components/review/review-form"
import { useCourseDetail } from "@/hooks/use-course"
import { useReview, useUpdateReview } from "@/hooks/use-review"
import { useAuth } from "@/contexts/auth-context"
import { getCourseSemesters } from "@/lib/course-semesters"
import type { CreateReviewCommand, UpdateReviewCommand } from "@/api/review"

export function EditReviewPage() {
  const { reviewID } = useParams<{ reviewID: string }>()
  const id = Number(reviewID)
  const navigate = useNavigate()
  const { user } = useAuth()
  const { data: review, isLoading } = useReview(id)
  const { data: course } = useCourseDetail(review?.course_id ?? 0)
  const { mutateAsync, isPending } = useUpdateReview()
  const semesters = course ? getCourseSemesters(course) : undefined

  async function handleSubmit(cmd: CreateReviewCommand | UpdateReviewCommand) {
    await mutateAsync({ reviewID: id, cmd: cmd as UpdateReviewCommand })
    navigate(`/review/${id}`)
  }

  if (isLoading || !review) {
    return (
      <>
        <PageTitle>编辑点评</PageTitle>
        <PageShell>
          <div className="space-y-4">
            <Skeleton className="h-8 w-1/2" />
            <Skeleton className="h-64 w-full" />
          </div>
        </PageShell>
      </>
    )
  }

  const displayCourse = course ?? review.course

  return (
    <>
      <PageTitle>编辑点评</PageTitle>
      <PageShell>
        <Button asChild variant="ghost" size="sm" className="mb-4">
          <Link to={`/review/${id}`}>
            <RiArrowLeftLine data-icon="inline-start" />
            返回点评
          </Link>
        </Button>

        <section className="space-y-4">
          <div className="space-y-4">
            <h1 className="text-lg font-medium">编辑点评</h1>
            {displayCourse && <CourseHeaderMeta course={displayCourse} />}
          </div>

          <ReviewForm
            initialReview={review}
            semesters={semesters}
            onSubmit={handleSubmit}
            onCancel={() => navigate(-1)}
            isSubmitting={isPending}
            draftUserID={user?.id}
          />
        </section>
      </PageShell>
    </>
  )
}
