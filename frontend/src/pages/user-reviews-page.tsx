import { Link, useSearchParams } from "react-router-dom"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { ReviewList } from "@/components/review/review-list"
import { PaginationComponent } from "@/components/common/pagination"
import { Button } from "@/components/ui/button"
import { useAuth } from "@/contexts/auth-context"
import { useUserReviews } from "@/hooks/use-review"

const PAGE_SIZE = 20

export function UserReviewsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Math.max(1, Number(searchParams.get("page") ?? "1") || 1)
  const { user, isLoading: authLoading } = useAuth()
  const { data, isLoading } = useUserReviews(user?.id ?? 0, {
    page,
    page_size: PAGE_SIZE,
  })

  function handlePageChange(page: number) {
    const next = new URLSearchParams(searchParams)
    next.set("page", String(page))
    setSearchParams(next)
  }

  return (
    <>
      <PageTitle>我的点评</PageTitle>
      <PageShell>
        <div className="space-y-6">
          <div>
            <h1 className="text-2xl font-bold">我的点评</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              管理你发布的所有课程点评
            </p>
          </div>

          {!authLoading && !user && (
            <div className="py-12 text-center">
              <p className="text-muted-foreground">登录后可以查看我的点评</p>
              <Button asChild variant="link" className="mt-2">
                <Link to="/login">登录</Link>
              </Button>
            </div>
          )}

          {user && data && (
            <p className="text-sm text-muted-foreground">
              共 {data.total} 条点评
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
