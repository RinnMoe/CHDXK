import { copy } from "@/lib/copy"
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
import { EnrollmentImportDialog } from "@/components/course/enrollment-import-dialog"
import { EnrollmentEditorDialog } from "@/components/course/enrollment-editor-dialog"
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
      ? copy("copy.user.pages.user-courses-page.c461004afbf9")
      : view === "followed"
        ? copy("copy.user.pages.user-courses-page.7be923dbf8a8")
        : copy("copy.user.pages.user-courses-page.4c435ec140fe")

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
    const recordName =
      view === "followed"
        ? copy("copy.user.pages.user-courses-page.7eccc1d93c64")
        : copy("copy.user.pages.user-courses-page.af7d90f61e5d")

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
      <PageTitle>
        {copy("copy.user.pages.user-courses-page.e80149e50579")}
      </PageTitle>
      <PageShell>
        <div className="space-y-6">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <h1 className="text-2xl font-bold">
                {copy("copy.user.pages.user-courses-page.e80149e50579")}
              </h1>
              <p className="mt-1 text-sm text-muted-foreground">
                {copy("copy.user.pages.user-courses-page.ac571e162283")}
              </p>
            </div>
            <Button asChild size="sm" variant="outline">
              <Link to="/review/followed">
                <RiMessage3Line data-icon="inline-start" />
                {copy("copy.user.pages.user-courses-page.497d4c0de4ae")}
              </Link>
            </Button>
          </div>

          {view === "enrolled" && (
            <div className="flex flex-wrap gap-2">
              <EnrollmentImportDialog />
              <EnrollmentEditorDialog />
            </div>
          )}

          <Tabs value={view} onValueChange={handleViewChange}>
            <TabsList>
              <TabsTrigger value="enrolled">
                {copy("copy.user.pages.user-courses-page.c461004afbf9")}
              </TabsTrigger>
              <TabsTrigger value="followed">
                {copy("copy.user.pages.user-courses-page.9fe1d0c9cca3")}
              </TabsTrigger>
              <TabsTrigger value="ignored">
                {copy("copy.user.pages.user-courses-page.b66705ba0b8e")}
              </TabsTrigger>
            </TabsList>
          </Tabs>

          {view === "enrolled" && user && (
            <div className="flex flex-wrap items-center gap-2">
              <Select
                value={activeSemester ?? ALL}
                onValueChange={handleSemesterChange}
              >
                <SelectTrigger size="sm" className="w-44">
                  <SelectValue
                    placeholder={copy(
                      "copy.user.pages.user-courses-page.42647ab29d2b"
                    )}
                  />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={ALL}>
                    {copy("copy.user.pages.user-courses-page.a0670d57e5bf")}
                  </SelectItem>
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
              <p className="text-muted-foreground">
                {copy("copy.user.pages.user-courses-page.739f42e69719")}
              </p>
              <Button asChild variant="link" className="mt-2">
                <Link to="/login">
                  {copy("copy.user.pages.user-courses-page.1e2df9c3075a")}
                </Link>
              </Button>
            </div>
          )}

          {user && view === "enrolled" && enrolledCourses.data && (
            <p className="text-sm text-muted-foreground">
              {title}
              {copy("copy.user.pages.user-courses-page.2428420e407f")}
              {filteredEnrollments.length}
              {copy("copy.user.pages.user-courses-page.6fb2840d568a")}
            </p>
          )}

          {user && view !== "enrolled" && activeCourseQuery.data && (
            <p className="text-sm text-muted-foreground">
              {title}
              {copy("copy.user.pages.user-courses-page.2428420e407f")}
              {activeCourseQuery.data.total}
              {copy("copy.user.pages.user-courses-page.6fb2840d568a")}
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
