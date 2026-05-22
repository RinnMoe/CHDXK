import { useParams, useNavigate, Link } from "react-router-dom"
import { RiArrowLeftLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { PageShell } from "@/components/layout/page-shell"
import { ReviewForm } from "@/components/review/review-form"
import { useReview, useUpdateReview } from "@/hooks/use-review"
import type { CreateReviewCommand, UpdateReviewCommand } from "@/api/review"

export function EditReviewPage() {
  const { reviewID } = useParams<{ reviewID: string }>()
  const id = Number(reviewID)
  const navigate = useNavigate()
  const { data: review, isLoading } = useReview(id)
  const { mutateAsync, isPending } = useUpdateReview()

  async function handleSubmit(cmd: CreateReviewCommand | UpdateReviewCommand) {
    await mutateAsync({ reviewID: id, cmd: cmd as UpdateReviewCommand })
    navigate(`/reviews/${id}`)
  }

  if (isLoading || !review) {
    return (
      <PageShell>
        <div className="space-y-4">
          <Skeleton className="h-8 w-1/2" />
          <Skeleton className="h-64 w-full" />
        </div>
      </PageShell>
    )
  }

  return (
    <PageShell>
      <Button asChild variant="ghost" size="sm" className="mb-4">
        <Link to={`/reviews/${id}`}>
          <RiArrowLeftLine data-icon="inline-start" />
          返回评价
        </Link>
      </Button>

      <Card>
        <CardHeader>
          <CardTitle>编辑评价</CardTitle>
          {review.course && (
            <div className="text-sm text-muted-foreground space-y-0.5">
              <div>
                <span className="font-mono">{review.course.code}</span>
                <span className="mx-1">·</span>
                <span className="font-medium text-foreground">
                  {review.course.name}
                </span>
              </div>
              <div>
                主讲教师：{review.course.main_teacher.name}
                {review.course.main_teacher.title && (
                  <span className="ml-1">
                    ({review.course.main_teacher.title})
                  </span>
                )}
              </div>
            </div>
          )}
        </CardHeader>
        <CardContent>
          <ReviewForm
            initialReview={review}
            onSubmit={handleSubmit}
            onCancel={() => navigate(`/reviews/${id}`)}
            isSubmitting={isPending}
          />
        </CardContent>
      </Card>
    </PageShell>
  )
}
