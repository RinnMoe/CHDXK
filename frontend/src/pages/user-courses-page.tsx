import { getRouteApi, Link, useNavigate } from "@tanstack/react-router"
import { RiMessage3Line } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { CourseList } from "@/components/course/course-list"
import { CourseEnrollmentList } from "@/components/course/course-enrollment-list"
import { CourseNotificationDeleteAction } from "@/components/course/course-notification-delete-action"
import { PaginationComponent } from "@/components/common/pagination"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { useAuth } from "@/contexts/auth-context"
import {
  useCourseEnrollments,
  useFollowedCourses,
  useIgnoredCourses,
  useSetNotificationLevel,
} from "@/hooks/use-course"

const PAGE_SIZE = 20
const ALL = "__all__"
const routeApi = getRouteApi("/app/course/mine")

type View = "enrolled" | "followed" | "ignored"

export function UserCoursesPage() {
  const search = routeApi.useSearch()
  const navigate = useNavigate({ from: "/course/mine" })
  const typeParam = search.type
  const view: View =
    typeParam === "followed" || typeParam === "ignored" ? typeParam : "enrolled"
  const page = Math.max(1, search.page ?? 1)
  const semester = search.semester
  const filter = { page, page_size: PAGE_SIZE }
  const { user, isLoading: authLoading } = useAuth()

  const enrolledCourses = useCourseEnrollments(!!user && view === "enrolled")
  const followedCourses = useFollowedCourses(
    filter,
    !!user && view === "followed"
  )
  const ignoredCourses = useIgnoredCourses(filter, !!user && view === "ignored")
  const notificationMutation = useSetNotificationLevel()
  const enrolledItems = enrolledCourses.data ?? []
  const enrollmentSemesters = [
    ...new Set(enrolledItems.map((item) => item.semester)),
  ].sort((a, b) => b.localeCompare(a))
  const activeSemester =
    semester && enrollmentSemesters.includes(semester) ? semester : undefined
  const filteredEnrollments = activeSemester
    ? enrolledItems.filter((item) => item.semester === activeSemester)
    : enrolledItems
  const activeCourseQuery =
    view === "followed" ? followedCourses : ignoredCourses
  const title =
    view === "enrolled"
      ? "选课记录"
      : view === "followed"
        ? "已关注课程"
        : "已屏蔽课程"

  function handleViewChange(nextView: string) {
    const nextType: View =
      nextView === "followed" || nextView === "ignored" ? nextView : "enrolled"
    void navigate({
      search: (prev) => ({
        ...prev,
        type: nextType,
        page: 1,
      }),
      resetScroll: false,
    })
  }

  function handlePageChange(nextPage: number) {
    void navigate({
      search: (prev) => ({ ...prev, page: nextPage }),
      resetScroll: false,
    })
  }

  function handleSemesterChange(value: string) {
    void navigate({
      search: (prev) => ({
        ...prev,
        semester: value === ALL ? undefined : value,
        type: "enrolled",
        page: 1,
      }),
      resetScroll: false,
    })
  }

  function renderNotificationDeleteAction(
    course: NonNullable<typeof activeCourseQuery.data>["items"][number]
  ) {
    const recordName = view === "followed" ? "关注记录" : "屏蔽记录"

    return (
      <CourseNotificationDeleteAction
        course={course}
        recordName={recordName}
        onDelete={() =>
          notificationMutation.mutate({ courseID: course.id, level: 0 })
        }
      />
    )
  }

  return (
    <>
      <PageTitle>我的课程</PageTitle>
      <PageShell>
        <div className="space-y-6">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <h1 className="text-2xl font-bold">我的课程</h1>
              <p className="mt-1 text-sm text-muted-foreground">
                查看选课记录、已关注和已屏蔽的课程
              </p>
            </div>
            <Button asChild size="sm" variant="outline">
              <Link to="/review/followed">
                <RiMessage3Line data-icon="inline-start" />
                关注动态
              </Link>
            </Button>
          </div>

          <Tabs value={view} onValueChange={handleViewChange}>
            <TabsList>
              <TabsTrigger value="enrolled">选课记录</TabsTrigger>
              <TabsTrigger value="followed">已关注</TabsTrigger>
              <TabsTrigger value="ignored">已屏蔽</TabsTrigger>
            </TabsList>
          </Tabs>

          {view === "enrolled" && user && (
            <div className="flex flex-wrap items-center gap-2">
              <Select
                value={activeSemester ?? ALL}
                onValueChange={handleSemesterChange}
              >
                <SelectTrigger size="sm" className="w-44">
                  <SelectValue placeholder="按学期筛选" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={ALL}>全部学期</SelectItem>
                  {enrollmentSemesters.map((s) => (
                    <SelectItem key={s} value={s}>
                      {s}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}

          {!authLoading && !user && (
            <div className="py-12 text-center">
              <p className="text-muted-foreground">登录后可以查看我的课程</p>
              <Button asChild variant="link" className="mt-2">
                <Link to="/login">登录</Link>
              </Button>
            </div>
          )}

          {user && view === "enrolled" && enrolledCourses.data && (
            <p className="text-sm text-muted-foreground">
              {title}共 {filteredEnrollments.length} 条
            </p>
          )}

          {user && view !== "enrolled" && activeCourseQuery.data && (
            <p className="text-sm text-muted-foreground">
              {title}共 {activeCourseQuery.data.total} 条
            </p>
          )}

          {user && view === "enrolled" && (
            <CourseEnrollmentList
              enrollments={filteredEnrollments}
              isLoading={authLoading || enrolledCourses.isLoading}
            />
          )}

          {user && view !== "enrolled" && (
            <CourseList
              courses={
                view === "followed"
                  ? (followedCourses.data?.items ?? [])
                  : (ignoredCourses.data?.items ?? [])
              }
              isLoading={
                authLoading ||
                (view === "followed"
                  ? followedCourses.isLoading
                  : ignoredCourses.isLoading)
              }
              renderAction={renderNotificationDeleteAction}
            />
          )}

          {user &&
            view !== "enrolled" &&
            activeCourseQuery.data &&
            activeCourseQuery.data.total > 0 && (
              <div className="flex justify-center pt-4">
                <PaginationComponent
                  page={activeCourseQuery.data.page}
                  pageSize={activeCourseQuery.data.page_size}
                  total={activeCourseQuery.data.total}
                  onPageChange={handlePageChange}
                />
              </div>
            )}
        </div>
      </PageShell>
    </>
  )
}
