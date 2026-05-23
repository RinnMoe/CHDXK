import { useEffect, useState } from "react"
import { useSearchParams } from "react-router-dom"
import { PageShell } from "@/components/layout/page-shell"
import { ReviewList } from "@/components/review/review-list"
import { PaginationComponent } from "@/components/common/pagination"
import { Input } from "@/components/ui/input"
import { useReviews } from "@/hooks/use-review"

const REVIEW_PAGE_SIZE = 20

export function ReviewsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Math.max(1, Number(searchParams.get("page") ?? "1") || 1)
  const q = searchParams.get("q") ?? ""
  const [searchValue, setSearchValue] = useState(q)
  const { data, isLoading } = useReviews({
    q: q || undefined,
    page,
    page_size: REVIEW_PAGE_SIZE,
  })

  useEffect(() => {
    setSearchValue(q)
  }, [q])

  useEffect(() => {
    const handler = setTimeout(() => {
      const current = searchParams.get("q") ?? ""
      if (searchValue === current) return

      const next = new URLSearchParams(searchParams)
      if (searchValue) {
        next.set("q", searchValue)
      } else {
        next.delete("q")
      }
      next.delete("page")
      setSearchParams(next)
    }, 500)

    return () => clearTimeout(handler)
  }, [searchParams, searchValue, setSearchParams])

  function handlePageChange(page: number) {
    const next = new URLSearchParams(searchParams)
    next.set("page", String(page))
    setSearchParams(next)
  }

  return (
    <>
      <title>点评 - JCourse</title>
      <PageShell>
        <div className="space-y-6">
          <div>
            <h1 className="text-2xl font-bold">点评</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              浏览和搜索所有课程点评
            </p>
          </div>

          <Input
            placeholder="搜索点评内容..."
            value={searchValue}
            onChange={(e) => setSearchValue(e.target.value)}
            className="max-w-md"
          />

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
