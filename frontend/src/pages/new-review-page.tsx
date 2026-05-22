import { useParams, useNavigate, Link } from "react-router-dom"
import { RiArrowLeftLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { PageShell } from "@/components/layout/page-shell"
import { ReviewForm } from "@/components/review/review-form"
import { useCourseDetail } from "@/hooks/use-course"
import { useCreateReview } from "@/hooks/use-review"
import type { CreateReviewCommand, UpdateReviewCommand } from "@/api/review"

export function NewReviewPage() {
  const { courseID } = useParams<{ courseID: string }>()
  const id = Number(courseID)
  const navigate = useNavigate()
  const { data: course } = useCourseDetail(id)
  const { mutateAsync, isPending } = useCreateReview()
  const semesters = course?.offered_courses?.map((oc) => oc.semester) ?? []

  async function handleSubmit(cmd: CreateReviewCommand | UpdateReviewCommand) {
    await mutateAsync(cmd as CreateReviewCommand)
    navigate(`/courses/${id}`)
  }

  return (
    <>
      <title>写点评 - JCourse</title>
      <PageShell>
      <Button asChild variant="ghost" size="sm" className="mb-4">
        <Link to={`/courses/${id}`}>
          <RiArrowLeftLine data-icon="inline-start" />
          返回课程
        </Link>
      </Button>

      <Card>
        <CardHeader>
          <CardTitle>写点评</CardTitle>
          {course && (
            <div className="text-sm text-muted-foreground space-y-0.5">
              <div>
                <span className="font-mono">{course.code}</span>
                <span className="mx-1">·</span>
                <span className="font-medium text-foreground">
                  {course.name}
                </span>
              </div>
              <div>
                主讲教师：{course.main_teacher.name}
                {course.main_teacher.title && (
                  <span className="ml-1">({course.main_teacher.title})</span>
                )}
              </div>
            </div>
          )}
        </CardHeader>
        <CardContent>
          <ReviewForm
            courseID={id}
            semesters={semesters}
            onSubmit={handleSubmit}
            onCancel={() => navigate(`/courses/${id}`)}
            isSubmitting={isPending}
          />
        </CardContent>
      </Card>
    </PageShell>
    </>
  )
}
