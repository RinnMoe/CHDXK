import { getRouteApi, useNavigate, useRouter } from "@tanstack/react-router"
import { RiArrowLeftLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { ReviewCard } from "@/components/review/review-card"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { useReview } from "@/hooks/use-review"

const routeApi = getRouteApi("/app/review/$reviewID")

export function ReviewDetailPage() {
  const navigate = useNavigate()
  const router = useRouter()
  const { reviewID } = routeApi.useParams()
  const id = Number(reviewID)
  const { data: review, isLoading } = useReview(id)

  function handleBack() {
    if (router.history.canGoBack()) {
      router.history.back()
      return
    }

    void navigate({ to: "/review" })
  }

  if (isLoading) {
    return (
      <>
        <PageTitle>点评</PageTitle>
        <PageShell>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="mb-4"
            onClick={handleBack}
          >
            <RiArrowLeftLine data-icon="inline-start" />
            返回
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
        <PageTitle>点评</PageTitle>
        <PageShell>
          <div className="py-16 text-center">
            <p className="text-muted-foreground">点评不存在</p>
            <Button
              type="button"
              variant="link"
              className="mt-4"
              onClick={handleBack}
            >
              返回
            </Button>
          </div>
        </PageShell>
      </>
    )
  }

  return (
    <>
      <PageTitle>点评</PageTitle>
      <PageShell>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="mb-4"
          onClick={handleBack}
        >
          <RiArrowLeftLine data-icon="inline-start" />
          返回
        </Button>

        <div className="space-y-4">
          <ReviewCard review={review} showCourse showVoteButtons />
        </div>
      </PageShell>
    </>
  )
}
