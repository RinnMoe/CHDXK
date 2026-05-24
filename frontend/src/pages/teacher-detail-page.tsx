import { useParams, Link } from "react-router-dom"
import { RiArrowLeftLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { TitleBadge } from "@/components/ui/title-badge"
import { Skeleton } from "@/components/ui/skeleton"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { CourseCard } from "@/components/course/course-card"
import { PaginationComponent } from "@/components/common/pagination"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { displayTeacherTitle } from "@/lib/utils"
import { useSearchParams } from "react-router-dom"
import { useTeacher, useTeacherCourses } from "@/hooks/use-teacher"

const ALL = "__all__"
type CourseSort = "rating_count" | "rating_avg"

export function TeacherDetailPage() {
  const { teacherID } = useParams<{ teacherID: string }>()
  const id = Number(teacherID)
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Number(searchParams.get("page") ?? "1")
  const orderByParam = searchParams.get("order_by")
  const orderBy: CourseSort | undefined =
    orderByParam === "rating_count" || orderByParam === "rating_avg"
      ? orderByParam
      : undefined
  const { data: teacher, isLoading: isTeacherLoading } = useTeacher(id)
  const { data, isLoading } = useTeacherCourses(id, {
    order_by: orderBy,
    page,
    page_size: 20,
  })
  const teacherTitle = displayTeacherTitle(teacher?.title)

  function updateCourseSort(value: string) {
    const next = new URLSearchParams(searchParams)
    if (value === ALL) {
      next.delete("order_by")
    } else {
      next.set("order_by", value)
    }
    next.delete("page")
    setSearchParams(next)
  }

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
          <Link to="/teacher">
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
                <div className="flex flex-wrap items-center gap-1.5 font-mono text-sm text-muted-foreground">
                  <span>{teacher.code}</span>
                  {teacherTitle && <TitleBadge>{teacherTitle}</TitleBadge>}
                </div>
                <h1 className="text-3xl font-bold">{teacher.name}</h1>
                {teacher.department && (
                  <div className="text-sm text-muted-foreground">
                    {teacher.department}
                  </div>
                )}
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
            <div className="mb-3 space-y-3">
              <div className="flex items-baseline justify-between">
                <h2 className="text-lg font-semibold">开设课程</h2>
                {data && (
                  <span className="text-sm text-muted-foreground">
                    共 {data.total} 门
                  </span>
                )}
              </div>
              <div>
                <Tabs
                  value={searchParams.get("order_by") ?? ALL}
                  onValueChange={updateCourseSort}
                >
                  <TabsList>
                    <TabsTrigger value={ALL}>默认</TabsTrigger>
                    <TabsTrigger value="rating_count">点评数量</TabsTrigger>
                    <TabsTrigger value="rating_avg">平均评分</TabsTrigger>
                  </TabsList>
                </Tabs>
              </div>
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
