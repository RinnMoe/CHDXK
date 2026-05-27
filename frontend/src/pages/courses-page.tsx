import { useSearchParams } from "react-router-dom"
import { CourseSearchBar } from "@/components/course/course-search-bar"
import { CourseList } from "@/components/course/course-list"
import { CourseFilters } from "@/components/course/course-filters"
import { PaginationComponent } from "@/components/common/pagination"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { useCourseFilters, useCourses } from "@/hooks/use-course"
import { cn } from "@/lib/utils"

export function CoursesPage() {
  const [searchParams, setSearchParams] = useSearchParams()

  function toFilter() {
    const categories = searchParams.getAll("categories")
    const target_years = searchParams.getAll("target_years")
    const departments = searchParams.getAll("department")
    const creditParam = searchParams.get("credit")
    const credit = creditParam === null ? undefined : Number(creditParam)
    const orderByParam = searchParams.get("order_by")
    const orderBy =
      orderByParam === "rating_score" || orderByParam === "rating_count"
        ? orderByParam
        : undefined
    return {
      q: searchParams.get("q")?.trim() || undefined,
      department: departments[0] ?? undefined,
      language: searchParams.get("language") ?? undefined,
      categories: categories.length > 0 ? categories : undefined,
      target_years: target_years.length > 0 ? target_years : undefined,
      credit: credit !== undefined && Number.isFinite(credit) ? credit : undefined,
      order_by: orderBy,
      page: Number(searchParams.get("page") ?? "1"),
      page_size: 20,
    }
  }

  const filter = toFilter()
  const { data: filters } = useCourseFilters()
  const { data, isLoading } = useCourses(filter)

  function handlePageChange(page: number) {
    const next = new URLSearchParams(searchParams)
    next.set("page", String(page))
    setSearchParams(next)
  }

  return (
    <>
      <PageTitle>课程</PageTitle>
      <PageShell>
        <div className="space-y-6 lg:flex lg:flex-col lg:gap-6 lg:space-y-0">
          <div>
            <h1 className="text-2xl font-bold">课程</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              浏览和搜索所有课程
            </p>
          </div>

          <CourseSearchBar />

          <div className="flex flex-col gap-6 lg:flex-row">
            {filters && (
              <div className="w-full shrink-0 lg:w-1/4">
                <CourseFilters filters={filters} />
              </div>
            )}

            <div
              className={cn(
                "min-w-0 flex-1 space-y-4",
                filters && "lg:border-l lg:pl-6"
              )}
            >
              {data && (
                <div className="text-sm text-muted-foreground">
                  共 {data.total} 门课程
                </div>
              )}

              <CourseList courses={data?.items ?? []} isLoading={isLoading} />

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
          </div>
        </div>
      </PageShell>
    </>
  )
}
