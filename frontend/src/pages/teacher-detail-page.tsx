import { useParams, Link } from "react-router-dom"
import { RiArrowLeftLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { TitleBadge } from "@/components/ui/title-badge"
import { Skeleton } from "@/components/ui/skeleton"
import { CourseCard } from "@/components/course/course-card"
import { PaginationComponent } from "@/components/common/pagination"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { useSearchParams } from "react-router-dom"
import { useTeacher, useTeacherCourses } from "@/hooks/use-teacher"

export function TeacherDetailPage() {
  const { teacherID } = useParams<{ teacherID: string }>()
  const id = Number(teacherID)
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Number(searchParams.get("page") ?? "1")
  const { data: teacher, isLoading: isTeacherLoading } = useTeacher(id)
  const { data, isLoading } = useTeacherCourses(id, { page, page_size: 20 })

  function handlePageChange(p: number) {
    const next = new URLSearchParams(searchParams)
    next.set("page", String(p))
    setSearchParams(next)
  }

  return (
    <>
      <PageTitle>{teacher?.name ?? "教师"}</PageTitle>
      <PageShell>
        <Button asChild variant="ghost" size="sm" className="mb-4">
          <Link to="/teachers">
            <RiArrowLeftLine data-icon="inline-start" />
            返回教师列表
          </Link>
        </Button>

        <div className="space-y-6">
          <header className="space-y-3">
            {isTeacherLoading ? (
              <>
                <Skeleton className="h-4 w-24" />
                <Skeleton className="h-9 w-40" />
                <div className="flex gap-2">
                  <Skeleton className="h-6 w-16" />
                  <Skeleton className="h-6 w-24" />
                </div>
              </>
            ) : teacher ? (
              <>
                <div className="font-mono text-sm text-muted-foreground">
                  {teacher.code}
                </div>
                <h1 className="text-3xl font-bold">{teacher.name}</h1>
                <div className="flex flex-wrap gap-2">
                  {teacher.title && <TitleBadge>{teacher.title}</TitleBadge>}
                  {teacher.department && (
                    <Badge variant="outline">{teacher.department}</Badge>
                  )}
                </div>
              </>
            ) : (
              <>
                <div className="font-mono text-sm text-muted-foreground">
                  #{id}
                </div>
                <h1 className="text-3xl font-bold">教师不存在</h1>
              </>
            )}
          </header>

          <section>
            <div className="mb-3 flex items-baseline justify-between">
              <h2 className="text-lg font-semibold">开设课程</h2>
              {data && (
                <span className="text-sm text-muted-foreground">
                  共 {data.total} 门
                </span>
              )}
            </div>

            {isLoading ? (
              <div className="border-t">
                {Array.from({ length: 3 }).map((_, i) => (
                  <div key={i} className="space-y-2 border-b px-4 py-3">
                    <div className="flex items-center justify-between gap-2">
                      <div className="flex items-center gap-2">
                        <Skeleton className="h-3 w-16" />
                        <Skeleton className="h-4 w-20" />
                      </div>
                      <Skeleton className="h-4 w-28" />
                    </div>
                    <Skeleton className="h-4 w-3/4" />
                    <Skeleton className="h-4 w-40" />
                  </div>
                ))}
              </div>
            ) : data && data.items.length > 0 ? (
              <>
                <div className="border-t">
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
              <p className="py-8 text-center text-sm text-muted-foreground">
                暂无开设课程
              </p>
            )}
          </section>
        </div>
      </PageShell>
    </>
  )
}
