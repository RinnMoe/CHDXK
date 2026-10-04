import { copy } from "@/lib/copy"
import { getRouteApi, Link, useNavigate } from "@tanstack/react-router"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { ReviewList } from "@/components/review/review-list"
import { PaginationComponent } from "@/components/common/pagination"
import { Button } from "@/components/ui/button"
import { useAuth } from "@/contexts/auth-context"
import { useUserReviews } from "@/hooks/use-review"

const PAGE_SIZE = 20
const routeApi = getRouteApi("/app/review/mine")

export function UserReviewsPage() {
  const search = routeApi.useSearch()
  const navigate = useNavigate({ from: "/review/mine" })
  const page = Math.max(1, search.page ?? 1)
  const { user, isLoading: authLoading } = useAuth()
  const { data, isLoading } = useUserReviews(user?.id ?? 0, {
    page,
    page_size: PAGE_SIZE,
  })

  function handlePageChange(page: number) {
    void navigate({
      search: (prev) => ({ ...prev, page }),
      resetScroll: false,
    })
  }

  return (
    <>
      <PageTitle>
        {copy("copy.user.pages.user-reviews-page.477d1ac374fc")}
      </PageTitle>
      <PageShell>
        <div className="space-y-6">
          <div>
            <h1 className="text-2xl font-bold">
              {copy("copy.user.pages.user-reviews-page.477d1ac374fc")}
            </h1>
            <p className="mt-1 text-sm text-muted-foreground">
              {copy("copy.user.pages.user-reviews-page.9608ec435fb4")}
            </p>
          </div>

          {!authLoading && !user && (
            <div className="py-12 text-center">
              <p className="text-muted-foreground">
                {copy("copy.user.pages.user-reviews-page.70507e72f6b8")}
              </p>
              <Button asChild variant="link" className="mt-2">
                <Link to="/login">
                  {copy("copy.user.pages.user-reviews-page.1e2df9c3075a")}
                </Link>
              </Button>
            </div>
          )}

          {user && data && (
            <p className="text-sm text-muted-foreground">
              {copy("copy.user.pages.user-reviews-page.2428420e407f")}
              {data.total}
              {copy("copy.user.pages.user-reviews-page.1c0a8a53fbb5")}
            </p>
          )}

          {user && (
            <ReviewList
              reviews={data?.items ?? []}
              isLoading={authLoading || isLoading}
              showCourse
            />
          )}

          {user && data && data.total > 0 && (
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
