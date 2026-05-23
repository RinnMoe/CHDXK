import { useSearchParams } from "react-router-dom"
import { PageShell } from "@/components/layout/page-shell"
import { ReviewList } from "@/components/review/review-list"
import { PaginationComponent } from "@/components/common/pagination"
import { useLatestReviews } from "@/hooks/use-review"

export function LatestReviewsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Number(searchParams.get("page") ?? "1")
  const { data, isLoading } = useLatestReviews({ page, page_size: 20 })

  function handlePageChange(page: number) {
    const next = new URLSearchParams(searchParams)
    next.set("page", String(page))
    setSearchParams(next)
  }

  return (
    <>
      <title>最新点评 - JCourse</title>
      <PageShell>
        <div className="space-y-6">
          <div>
            <h1 className="text-2xl font-bold">最新点评</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              查看所有最新课程点评
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
