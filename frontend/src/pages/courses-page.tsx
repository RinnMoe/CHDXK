import { useSearchParams } from "react-router-dom"
import { CourseSearchBar } from "@/components/course/course-search-bar"
import { CourseList } from "@/components/course/course-list"
import { CourseFilters } from "@/components/course/course-filters"
import { PaginationComponent } from "@/components/common/pagination"
import { PageShell } from "@/components/layout/page-shell"
import { useCourseFilters, useCourses } from "@/hooks/use-course"

export function CoursesPage() {
  const [searchParams, setSearchParams] = useSearchParams()

  function toFilter() {
    const categories = searchParams.getAll("categories")
    const target_years = searchParams.getAll("target_years")
    const departments = searchParams.getAll("department")
    const credits = searchParams
      .getAll("credit")
      .map(Number)
      .filter(Boolean)
    return {
      q: searchParams.get("q") ?? undefined,
      department: departments[0] ?? undefined,
      language: searchParams.get("language") ?? undefined,
      categories: categories.length > 0 ? categories : undefined,
      target_years: target_years.length > 0 ? target_years : undefined,
      credit: credits[0],
      order_by: (searchParams.get("order_by") as "rating_count" | "rating_avg") ?? undefined,
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
      <title>课程 - JCourse</title>
      <PageShell>
      <div className="space-y-6 lg:flex lg:flex-col lg:space-y-0 lg:gap-6">
        <div>
          <h1 className="text-2xl font-bold">课程</h1>
          <p className="text-sm text-muted-foreground mt-1">
            浏览和搜索所有课程
          </p>
        </div>

        <CourseSearchBar />

        <div className="flex flex-col gap-6 lg:flex-row">
          {filters && <div className="w-full lg:w-1/4 shrink-0"><CourseFilters filters={filters} /></div>}

          <div className="min-w-0 flex-1 space-y-4">
            {data && (
              <div className="text-sm text-muted-foreground">
                共 {data.total} 门课程
              </div>
            )}

            <CourseList
              courses={data?.items ?? []}
              isLoading={isLoading}
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
        </div>
      </div>
      </PageShell>
    </>
  )
}
