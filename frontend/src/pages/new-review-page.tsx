import { useParams, useNavigate, Link } from "react-router-dom"
import { RiArrowLeftLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { CourseHeaderMeta } from "@/components/course/course-header-meta"
import { ReviewForm } from "@/components/review/review-form"
import { useCourseDetail } from "@/hooks/use-course"
import { useCreateReview } from "@/hooks/use-review"
import { getCourseSemesters } from "@/lib/course-semesters"
import type { CreateReviewCommand, UpdateReviewCommand } from "@/api/review"

export function NewReviewPage() {
  const { courseID } = useParams<{ courseID: string }>()
  const id = Number(courseID)
  const navigate = useNavigate()
  const { data: course } = useCourseDetail(id)
  const { mutateAsync, isPending } = useCreateReview()
  const semesters = getCourseSemesters(course)

  async function handleSubmit(cmd: CreateReviewCommand | UpdateReviewCommand) {
    await mutateAsync(cmd as CreateReviewCommand)
    navigate(`/course/${id}`)
  }

  return (
    <>
      <PageTitle>写点评</PageTitle>
      <PageShell>
        <Button asChild variant="ghost" size="sm" className="mb-4">
          <Link to={`/course/${id}`}>
            <RiArrowLeftLine data-icon="inline-start" />
            返回课程
          </Link>
        </Button>

        <Card className="shadow-none ring-0">
          <CardHeader className="gap-4">
            <CardTitle>写点评</CardTitle>
            {course && <CourseHeaderMeta course={course} />}
          </CardHeader>
          <CardContent>
            <ReviewForm
              courseID={id}
              semesters={semesters}
              onSubmit={handleSubmit}
              onCancel={() => navigate(`/course/${id}`)}
              isSubmitting={isPending}
            />
          </CardContent>
        </Card>
      </PageShell>
    </>
  )
}
