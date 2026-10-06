import { copy } from "@/lib/copy"
import { getRouteApi, useNavigate } from "@tanstack/react-router"
import { ReviewDialog } from "@/components/review/review-dialog"
import { ReviewForm } from "@/components/review/review-form"
import { WrongCourseWarning } from "@/components/review/wrong-course-warning"
import { useCourseDetail } from "@/hooks/use-course"
import { useCreateReview } from "@/hooks/use-review"
import {
  getCurrentSemesterSetting,
  useSystemSettings,
} from "@/hooks/use-system-settings"
import { useAuth } from "@/contexts/auth-context"
import {
  filterSemestersUpTo,
  getCourseSemesters,
  getDefaultSemester,
} from "@/lib/course-semesters"
import type { CreateReviewCommand, UpdateReviewCommand } from "@/api/review"

const routeApi = getRouteApi("/app/course/$courseID/review/new")

export function NewReviewPage() {
  const { courseID } = routeApi.useParams()
  const id = Number(courseID)
  const navigate = useNavigate()
  const { user } = useAuth()
  const { data: course } = useCourseDetail(id)
  const systemSettingsQuery = useSystemSettings(!!user)
  const { mutateAsync, isPending } = useCreateReview()
  const currentSemester = getCurrentSemesterSetting(systemSettingsQuery.data)
  const semesters = systemSettingsQuery.isLoading
    ? []
    : filterSemestersUpTo(getCourseSemesters(course), currentSemester)
  const defaultSemester = getDefaultSemester(semesters, currentSemester)

  function closeDialog() {
    if (!isPending)
      void navigate({
        to: "/course/$courseID",
        params: { courseID: String(id) },
        replace: true,
      })
  }

  async function handleSubmit(cmd: CreateReviewCommand | UpdateReviewCommand) {
    await mutateAsync(cmd as CreateReviewCommand)
    await navigate({
      to: "/course/$courseID",
      params: { courseID: String(id) },
    })
  }

  return (
    <ReviewDialog
      title={copy("copy.user.pages.new-review-page.74cdbe78b5b2")}
      course={course}
      onClose={closeDialog}
      busy={isPending}
    >
      <WrongCourseWarning course={course} currentSemester={currentSemester} />
      <ReviewForm
        courseID={id}
        semesters={semesters}
        defaultSemester={defaultSemester}
        onSubmit={handleSubmit}
        onCancel={closeDialog}
        isSubmitting={isPending}
        draftUserID={user?.id}
      />
    </ReviewDialog>
  )
}
