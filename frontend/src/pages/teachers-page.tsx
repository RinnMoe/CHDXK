import { useEffect, useState } from "react"
import { useSearchParams } from "react-router-dom"
import { useDebounce } from "use-debounce"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { TeacherFilters } from "@/components/teacher/teacher-filters"
import { TeacherList } from "@/components/teacher/teacher-list"
import { PaginationComponent } from "@/components/common/pagination"
import { Input } from "@/components/ui/input"
import { useTeacherFilters, useTeachers } from "@/hooks/use-teacher"
import { cn } from "@/lib/utils"

const TEACHER_SEARCH_DEBOUNCE_MS = 250

type TeacherSearchInputProps = {
  initialValue: string
  onSearchChange: (value: string) => void
}

function TeacherSearchInput({
  initialValue,
  onSearchChange,
}: TeacherSearchInputProps) {
  const [searchValue, setSearchValue] = useState(initialValue)
  const [debouncedSearchValue] = useDebounce(
    searchValue,
    TEACHER_SEARCH_DEBOUNCE_MS
  )

  useEffect(() => {
    onSearchChange(debouncedSearchValue)
  }, [debouncedSearchValue, onSearchChange])

  return (
    <Input
      placeholder="搜索教师姓名、拼音或工号..."
      value={searchValue}
      onChange={(e) => setSearchValue(e.target.value)}
    />
  )
}

export function TeachersPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Number(searchParams.get("page") ?? "1")
  const department = searchParams.get("department") ?? ""
  const title = searchParams.get("title") ?? ""
  const q = (searchParams.get("q") ?? "").trim()

  const filter = {
    department: department || undefined,
    title: title || undefined,
    q: q || undefined,
    page,
    page_size: 20,
  }
  const { data: filters, isLoading: filtersLoading } = useTeacherFilters()
  const { data, isLoading } = useTeachers(filter)

  function update(key: string, value: string | null) {
    const next = new URLSearchParams(searchParams)
    if (!value) next.delete(key)
    else next.set(key, value)
    next.delete("page")
    setSearchParams(next)
  }

  function handleSearchChange(value: string) {
    const nextQ = value.trim()
    if (nextQ === q) return
    update("q", nextQ)
  }

  function handlePageChange(p: number) {
    const next = new URLSearchParams(searchParams)
    next.set("page", String(p))
    setSearchParams(next)
  }

  return (
    <>
      <PageTitle>教师</PageTitle>
      <PageShell>
        <div className="space-y-6 lg:flex lg:flex-col lg:gap-6 lg:space-y-0">
          <div>
            <h1 className="text-2xl font-bold">教师</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              浏览和搜索所有教师
            </p>
          </div>

          <TeacherSearchInput
            initialValue={q}
            onSearchChange={handleSearchChange}
          />

          <div className="flex flex-col gap-6 lg:flex-row">
            {filters && (
              <div className="w-full shrink-0 lg:w-1/4">
                <TeacherFilters filters={filters} />
              </div>
            )}

            <div
              className={cn(
                "min-w-0 flex-1 space-y-4",
                filters && "lg:border-l lg:pl-6"
              )}
            >
              {data && (
                <p className="text-sm text-muted-foreground">
                  共 {data.total} 位教师
                </p>
              )}

              <TeacherList
                teachers={data?.items ?? []}
                isLoading={isLoading || filtersLoading}
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
