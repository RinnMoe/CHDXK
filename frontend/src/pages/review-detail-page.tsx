import { useParams, Link } from "react-router-dom"
import { RiArrowLeftLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { ReviewCard } from "@/components/review/review-card"
import { PageShell } from "@/components/layout/page-shell"
import { useReview } from "@/hooks/use-review"

export function ReviewDetailPage() {
  const { reviewID } = useParams<{ reviewID: string }>()
  const id = Number(reviewID)
  const { data: review, isLoading } = useReview(id)

  if (isLoading) {
    return (
      <>
        <title>点评 - JCourse</title>
        <PageShell>
          <Button asChild variant="ghost" size="sm" className="mb-4">
            <Link to="/reviews/latest">
              <RiArrowLeftLine data-icon="inline-start" />
              返回
            </Link>
          </Button>
          <div className="space-y-4">
            <Skeleton className="h-8 w-1/2" />
            <Skeleton className="h-32 w-full" />
          </div>
        </PageShell>
      </>
    )
  }

  if (!review) {
    return (
      <>
        <title>点评 - JCourse</title>
        <PageShell>
          <div className="py-16 text-center">
            <p className="text-muted-foreground">点评不存在</p>
            <Button asChild variant="link" className="mt-4">
              <Link to="/reviews/latest">返回点评列表</Link>
            </Button>
          </div>
        </PageShell>
      </>
    )
  }

  return (
    <>
      <title>点评 - JCourse</title>
      <PageShell>
        <Button asChild variant="ghost" size="sm" className="mb-4">
          <Link to="/reviews/latest">
            <RiArrowLeftLine data-icon="inline-start" />
            返回点评列表
          </Link>
        </Button>

        <div className="space-y-4">
          <ReviewCard review={review} showCourse showVoteButtons />
        </div>
      </PageShell>
    </>
  )
}
