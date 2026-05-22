import { useSearchParams } from "react-router-dom"
import { PageShell } from "@/components/layout/page-shell"
import { ReviewList } from "@/components/review/review-list"
import { PaginationComponent } from "@/components/common/pagination"
import { useUserReviews } from "@/hooks/use-review"

export function UserReviewsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Number(searchParams.get("page") ?? "1")
  // Mock userID=1 for now — will be replaced by auth context
  const userID = 1
  const { data, isLoading } = useUserReviews(userID, { page, page_size: 20 })

  function handlePageChange(page: number) {
    const next = new URLSearchParams(searchParams)
    next.set("page", String(page))
    setSearchParams(next)
  }

  return (
    <PageShell>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold">我的评价</h1>
          <p className="text-sm text-muted-foreground mt-1">
            管理你发布的所有课程评价
          </p>
        </div>

        {data && (
          <p className="text-sm text-muted-foreground">
            共 {data.total} 条评价
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
  )
}
