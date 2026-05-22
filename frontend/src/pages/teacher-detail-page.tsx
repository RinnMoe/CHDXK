import { useParams, Link } from "react-router-dom"
import { RiArrowLeftLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { CourseCard } from "@/components/course/course-card"
import { PaginationComponent } from "@/components/common/pagination"
import { PageShell } from "@/components/layout/page-shell"
import { useSearchParams } from "react-router-dom"
import { getMockTeacher } from "@/mocks/fixtures/teachers"
import { useTeacherCourses } from "@/hooks/use-teacher"
import type { TeacherDTO } from "@/api/teacher"

export function TeacherDetailPage() {
  const { teacherID } = useParams<{ teacherID: string }>()
  const id = Number(teacherID)
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Number(searchParams.get("page") ?? "1")
  const { data, isLoading } = useTeacherCourses(id, { page, page_size: 20 })

  const mockTeacher = getMockTeacher(id) as TeacherDTO | undefined

  function handlePageChange(p: number) {
    const next = new URLSearchParams(searchParams)
    next.set("page", String(p))
    setSearchParams(next)
  }

  return (
    <PageShell>
      <Button asChild variant="ghost" size="sm" className="mb-4">
        <Link to="/teachers">
          <RiArrowLeftLine data-icon="inline-start" />
          返回教师列表
        </Link>
      </Button>

      <div className="space-y-6">
        <header className="space-y-3">
          <div className="text-sm text-muted-foreground font-mono">
            {mockTeacher?.code ?? `T${String(id).padStart(5, "0")}`}
          </div>
          <h1 className="text-3xl font-bold">
            {mockTeacher?.name ?? "教师"}
          </h1>
          <div className="flex flex-wrap gap-2">
            {mockTeacher?.title && (
              <Badge variant="secondary">{mockTeacher.title}</Badge>
            )}
            {mockTeacher?.department && (
              <Badge variant="outline">{mockTeacher.department}</Badge>
            )}
          </div>
        </header>

        <section>
          <div className="flex items-baseline justify-between mb-3">
            <h2 className="text-lg font-semibold">开设课程</h2>
            {data && (
              <span className="text-sm text-muted-foreground">
                共 {data.total} 门
              </span>
            )}
          </div>

          {isLoading ? (
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {Array.from({ length: 3 }).map((_, i) => (
                <div key={i} className="rounded-lg border p-4 space-y-2">
                  <Skeleton className="h-3 w-16" />
                  <Skeleton className="h-4 w-3/4" />
                  <Skeleton className="h-3 w-1/2" />
                </div>
              ))}
            </div>
          ) : data && data.items.length > 0 ? (
            <>
              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {data.items.map((course) => (
                  <CourseCard key={course.id} course={course} />
                ))}
              </div>
              <div className="flex justify-center pt-4">
                <PaginationComponent
                  page={data.page}
                  pageSize={data.page_size}
                  total={data.total}
                  onPageChange={handlePageChange}
                />
              </div>
            </>
          ) : (
            <p className="text-sm text-muted-foreground py-8 text-center">
              暂无开设课程
            </p>
          )}
        </section>
      </div>
    </PageShell>
  )
}
