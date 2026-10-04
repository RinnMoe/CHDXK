import { copy } from "@/lib/copy"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { ReviewList } from "@/components/review/review-list"
import { PaginationComponent } from "@/components/common/pagination"
import { useFollowedReviews } from "@/hooks/use-review"
import { getRouteApi, useNavigate } from "@tanstack/react-router"

const routeApi = getRouteApi("/app/review/followed")

export function FollowedReviewsPage() {
  const search = routeApi.useSearch()
  const navigate = useNavigate({ from: "/review/followed" })
  const page = search.page ?? 1
  const { data, isLoading } = useFollowedReviews({ page, page_size: 20 })

  function handlePageChange(page: number) {
    void navigate({
      search: (prev) => ({ ...prev, page }),
      resetScroll: false,
    })
  }

  return (
    <>
      <PageTitle>
        {copy("copy.user.pages.followed-reviews-page.497d4c0de4ae")}
      </PageTitle>
      <PageShell>
        <div className="space-y-6">
          <div>
            <h1 className="text-2xl font-bold">
              {copy("copy.user.pages.followed-reviews-page.497d4c0de4ae")}
            </h1>
            <p className="mt-1 text-sm text-muted-foreground">
              {copy("copy.user.pages.followed-reviews-page.e798d1fe24d6")}
            </p>
          </div>

          {data && (
            <p className="text-sm text-muted-foreground">
              {copy("copy.user.pages.followed-reviews-page.2428420e407f")}
              {data.total}
              {copy("copy.user.pages.followed-reviews-page.1c0a8a53fbb5")}
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
