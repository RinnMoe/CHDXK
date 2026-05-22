import { Link, useSearchParams } from "react-router-dom"
import { RiMessage3Line } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { CourseList } from "@/components/course/course-list"
import { PaginationComponent } from "@/components/common/pagination"
import { PageShell } from "@/components/layout/page-shell"
import { useAuth } from "@/contexts/auth-context"
import { useFollowedCourses, useIgnoredCourses } from "@/hooks/use-course"

const PAGE_SIZE = 20

type View = "followed" | "ignored"

export function UserCoursesPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const view: View = searchParams.get("type") === "ignored" ? "ignored" : "followed"
  const page = Math.max(1, Number(searchParams.get("page") ?? "1") || 1)
  const filter = { page, page_size: PAGE_SIZE }
  const { user, isLoading: authLoading } = useAuth()

  const followedCourses = useFollowedCourses(filter, !!user && view === "followed")
  const ignoredCourses = useIgnoredCourses(filter, !!user && view === "ignored")
  const activeQuery = view === "followed" ? followedCourses : ignoredCourses
  const title = view === "followed" ? "已关注课程" : "已屏蔽课程"

  function handleViewChange(nextView: string) {
    const next = new URLSearchParams(searchParams)
    next.set("type", nextView)
    next.set("page", "1")
    setSearchParams(next)
  }

  function handlePageChange(nextPage: number) {
    const next = new URLSearchParams(searchParams)
    next.set("page", String(nextPage))
    setSearchParams(next)
  }

  return (
    <>
      <title>我的课程 - JCourse</title>
      <PageShell>
        <div className="space-y-6">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <h1 className="text-2xl font-bold">我的课程</h1>
              <p className="mt-1 text-sm text-muted-foreground">
                查看已关注和已屏蔽的课程
              </p>
            </div>
            <Button asChild size="sm" variant="outline">
              <Link to="/reviews/followed">
                <RiMessage3Line data-icon="inline-start" />
                关注动态
              </Link>
            </Button>
          </div>

          <Tabs value={view} onValueChange={handleViewChange}>
            <TabsList>
              <TabsTrigger value="followed">已关注</TabsTrigger>
              <TabsTrigger value="ignored">已屏蔽</TabsTrigger>
            </TabsList>
          </Tabs>

          {!authLoading && !user && (
            <div className="py-12 text-center">
              <p className="text-muted-foreground">登录后可以查看我的课程</p>
              <Button asChild variant="link" className="mt-2">
                <Link to="/login">登录</Link>
              </Button>
            </div>
          )}

          {user && activeQuery.data && (
            <p className="text-sm text-muted-foreground">
              {title}共 {activeQuery.data.total} 门
            </p>
          )}

          {user && (
            <CourseList
              courses={activeQuery.data?.items ?? []}
              isLoading={authLoading || activeQuery.isLoading}
            />
          )}

          {user && activeQuery.data && activeQuery.data.total > 0 && (
            <div className="flex justify-center pt-4">
              <PaginationComponent
                page={activeQuery.data.page}
                pageSize={activeQuery.data.page_size}
                total={activeQuery.data.total}
                onPageChange={handlePageChange}
              />
            </div>
          )}
        </div>
      </PageShell>
    </>
  )
}
