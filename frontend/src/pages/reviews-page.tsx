import { useEffect, useState } from "react"
import { useSearchParams } from "react-router-dom"
import { useDebounce } from "use-debounce"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { ReviewList } from "@/components/review/review-list"
import { PaginationComponent } from "@/components/common/pagination"
import { Input } from "@/components/ui/input"
import { useReviews } from "@/hooks/use-review"

const REVIEW_PAGE_SIZE = 20
const REVIEW_SEARCH_DEBOUNCE_MS = 250

type ReviewSearchInputProps = {
  initialValue: string
  onSearchChange: (value: string) => void
}

function ReviewSearchInput({
  initialValue,
  onSearchChange,
}: ReviewSearchInputProps) {
  const [searchValue, setSearchValue] = useState(initialValue)
  const [debouncedSearchValue] = useDebounce(
    searchValue,
    REVIEW_SEARCH_DEBOUNCE_MS
  )

  useEffect(() => {
    onSearchChange(debouncedSearchValue)
  }, [debouncedSearchValue, onSearchChange])

  return (
    <Input
      placeholder="搜索点评内容..."
      value={searchValue}
      onChange={(e) => setSearchValue(e.target.value)}
    />
  )
}

export function ReviewsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Math.max(1, Number(searchParams.get("page") ?? "1") || 1)
  const q = (searchParams.get("q") ?? "").trim()
  const { data, isLoading } = useReviews({
    q: q || undefined,
    page,
    page_size: REVIEW_PAGE_SIZE,
  })

  function handleSearchChange(value: string) {
    const nextQ = value.trim()
    if (nextQ === q) return

    const next = new URLSearchParams(searchParams)
    if (nextQ) {
      next.set("q", nextQ)
    } else {
      next.delete("q")
    }
    next.delete("page")
    setSearchParams(next)
  }

  function handlePageChange(page: number) {
    const next = new URLSearchParams(searchParams)
    next.set("page", String(page))
    setSearchParams(next)
  }

  return (
    <>
      <PageTitle>点评</PageTitle>
      <PageShell>
        <div className="space-y-6">
          <div>
            <h1 className="text-2xl font-bold">点评</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              浏览和搜索所有课程点评
            </p>
          </div>

          <ReviewSearchInput
            initialValue={q}
            onSearchChange={handleSearchChange}
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
