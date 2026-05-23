import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { ReviewList } from "@/components/review/review-list"
import { PaginationComponent } from "@/components/common/pagination"
import { useFollowedReviews } from "@/hooks/use-review"
import { useSearchParams } from "react-router-dom"

export function FollowedReviewsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Number(searchParams.get("page") ?? "1")
  const { data, isLoading } = useFollowedReviews({ page, page_size: 20 })

  function handlePageChange(page: number) {
    const next = new URLSearchParams(searchParams)
    next.set("page", String(page))
    setSearchParams(next)
  }

  return (
    <>
      <PageTitle>关注动态</PageTitle>
      <PageShell>
        <div className="space-y-6">
          <div>
            <h1 className="text-2xl font-bold">关注动态</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              你关注的课程的点评
            </p>
          </div>

          {data && (
            <p className="text-sm text-muted-foreground">
              共 {data.total} 条点评
            </p>
          )}

          <ReviewList
            reviews={data?.items ?? []}
            isLoading={isLoading}
            showCourse
          />

          {data && data.total > 0 && (
            <div className="flex justify-center pt-4">
              <PaginationComponent
                page={data.page}
                pageSize={data.page_size}
                total={data.total}
                onPageChange={handlePageChange}
              />
            </div>
          )}
        </div>
      </PageShell>
    </>
  )
}
