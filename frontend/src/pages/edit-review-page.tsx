import { copy } from "@/lib/copy"
import { getRouteApi, useNavigate } from "@tanstack/react-router"
import { ReviewDialog } from "@/components/review/review-dialog"
import { ReviewForm } from "@/components/review/review-form"
import { useCourseDetail } from "@/hooks/use-course"
import { useReview, useUpdateReview } from "@/hooks/use-review"
import { useAuth } from "@/contexts/auth-context"
import { getCourseSemesters } from "@/lib/course-semesters"
import type { CreateReviewCommand, UpdateReviewCommand } from "@/api/review"

const routeApi = getRouteApi("/app/review/$reviewID/edit")

export function EditReviewPage() {
  const { reviewID } = routeApi.useParams()
  const id = Number(reviewID)
  const navigate = useNavigate()
  const { user } = useAuth()
  const { data: review, isLoading } = useReview(id)
  const { data: course } = useCourseDetail(review?.course_id ?? 0)
  const { mutateAsync, isPending } = useUpdateReview()
  const semesters = course ? getCourseSemesters(course) : undefined

  function closeDialog() {
    if (!isPending)
      void navigate({
        to: "/review/$reviewID",
        params: { reviewID: String(id) },
        replace: true,
      })
  }

  async function handleSubmit(cmd: CreateReviewCommand | UpdateReviewCommand) {
    await mutateAsync({ reviewID: id, cmd: cmd as UpdateReviewCommand })
    await navigate({
      to: "/review/$reviewID",
      params: { reviewID: String(id) },
    })
  }

  const displayCourse = course ?? review?.course

  return (
    <ReviewDialog
      title={copy("copy.user.pages.edit-review-page.bacd75e6ddb7")}
      course={displayCourse}
      onClose={closeDialog}
      busy={isPending}
    >
      {isLoading ? (
        <p role="status">正在加载点评…</p>
      ) : review ? (
        <ReviewForm
          initialReview={review}
          semesters={semesters}
          onSubmit={handleSubmit}
          onCancel={closeDialog}
          isSubmitting={isPending}
          draftUserID={user?.id}
        />
      ) : (
        <p role="alert">未找到这条点评。</p>
      )}
    </ReviewDialog>
  )
}
