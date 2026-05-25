import { useEffect, useMemo, useState } from "react"
import { Link, useSearchParams } from "react-router-dom"
import {
  RiDeleteBinLine,
  RiMessage3Line,
  RiRefreshLine,
} from "@remixicon/react"
import { courseEnrollmentSyncStartURL } from "@/api/course"
import {
  COURSE_ENROLLMENT_SYNC_CHANNEL,
  type CourseEnrollmentSyncMessage,
} from "@/lib/course-enrollment-sync"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
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
import { PaginationComponent } from "@/components/common/pagination"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { useAuth } from "@/contexts/auth-context"
import {
  useCourseEnrollments,
  useCourseFilters,
  useFollowedCourses,
  useIgnoredCourses,
  useSetNotificationLevel,
} from "@/hooks/use-course"
import { useUserSettings } from "@/hooks/use-user-settings"
import { getDefaultSemester } from "@/lib/course-semesters"

const PAGE_SIZE = 20
const ALL = "__all__"

type View = "enrolled" | "followed" | "ignored"

export function UserCoursesPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const typeParam = searchParams.get("type")
  const view: View =
    typeParam === "followed" || typeParam === "ignored" ? typeParam : "enrolled"
  const page = Math.max(1, Number(searchParams.get("page") ?? "1") || 1)
  const semester = searchParams.get("semester") ?? undefined
  const filter = { page, page_size: PAGE_SIZE }
  const { user, isLoading: authLoading } = useAuth()
  const [syncOpen, setSyncOpen] = useState(false)
  const [syncSemester, setSyncSemester] = useState("")
  const [syncMessage, setSyncMessage] = useState("")

  const enrolledCourses = useCourseEnrollments(!!user && view === "enrolled")
  const filtersQuery = useCourseFilters()
  const settingsQuery = useUserSettings(!!user)
  const followedCourses = useFollowedCourses(
    filter,
    !!user && view === "followed"
  )
  const ignoredCourses = useIgnoredCourses(filter, !!user && view === "ignored")
  const notificationMutation = useSetNotificationLevel()
  const syncSemesters = useMemo(
    () => filtersQuery.data?.semesters?.filter((item) => item.name) ?? [],
    [filtersQuery.data?.semesters]
  )
  const syncSemesterNames = useMemo(
    () => syncSemesters.map((item) => item.name),
    [syncSemesters]
  )
  const defaultSyncSemester = getDefaultSemester(
    syncSemesterNames,
    settingsQuery.data?.current_semester
  )
  const enrolledItems = enrolledCourses.data ?? []
  const enrollmentSemesters = [
    ...new Set(enrolledItems.map((item) => item.semester)),
  ].sort((a, b) => b.localeCompare(a))
  const filteredEnrollments = semester
    ? enrolledItems.filter((item) => item.semester === semester)
    : enrolledItems
  const activeCourseQuery =
    view === "followed" ? followedCourses : ignoredCourses
  const title =
    view === "enrolled"
      ? "选课记录"
      : view === "followed"
        ? "已关注课程"
        : "已屏蔽课程"

  useEffect(() => {
    const handleMessage = (payload: CourseEnrollmentSyncMessage) => {
      if (payload.status === "ok") {
        void enrolledCourses.refetch()
        setSyncMessage(
          `已同步 ${payload.semester}，匹配 ${payload.matched ?? 0} 条记录。`
        )
        const next = new URLSearchParams(searchParams)
        next.set("type", "enrolled")
        if (payload.semester) next.set("semester", payload.semester)
        setSearchParams(next, { replace: true })
        return
      }
      setSyncMessage(payload.message ?? "同步失败，请重试。")
    }

    const channel = new BroadcastChannel(COURSE_ENROLLMENT_SYNC_CHANNEL)
    channel.onmessage = (event: MessageEvent<CourseEnrollmentSyncMessage>) => {
      handleMessage(event.data)
    }
    const onWindowMessage = (event: MessageEvent) => {
      if (event.origin !== window.location.origin) return
      if (event.data?.type !== COURSE_ENROLLMENT_SYNC_CHANNEL) return
      handleMessage(event.data.payload as CourseEnrollmentSyncMessage)
    }
    window.addEventListener("message", onWindowMessage)
    return () => {
      channel.close()
      window.removeEventListener("message", onWindowMessage)
    }
  }, [enrolledCourses, searchParams, setSearchParams])

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

  function handleSemesterChange(value: string) {
    const next = new URLSearchParams(searchParams)
    if (value !== ALL) next.set("semester", value)
    else next.delete("semester")
    next.set("type", "enrolled")
    next.set("page", "1")
    setSearchParams(next)
  }

  function handleOpenSyncDialog() {
    setSyncSemester(defaultSyncSemester)
    setSyncOpen(true)
  }

  function handleStartSync() {
    if (!syncSemester) return
    setSyncMessage("")
    const width = 560
    const height = 720
    const left = Math.max(0, window.screenX + (window.outerWidth - width) / 2)
    const top = Math.max(0, window.screenY + (window.outerHeight - height) / 2)
    const popup = window.open(
      courseEnrollmentSyncStartURL(syncSemester),
      "jaccount-course-sync",
      `popup=yes,width=${width},height=${height},left=${left},top=${top},menubar=no,toolbar=no,status=no,resizable=yes,scrollbars=yes`
    )
    if (!popup) {
      setSyncMessage("浏览器阻止了登录窗口，请允许弹出窗口后重试。")
      return
    }
    popup.focus()
    setSyncOpen(false)
  }

  function renderNotificationDeleteAction(
    course: NonNullable<typeof activeCourseQuery.data>["items"][number]
  ) {
    const recordName = view === "followed" ? "关注记录" : "屏蔽记录"

    return (
      <AlertDialog>
        <AlertDialogTrigger asChild>
          <Button variant="outline" size="sm">
            <RiDeleteBinLine />
            删除
          </Button>
        </AlertDialogTrigger>
        <AlertDialogContent size="sm">
          <AlertDialogHeader>
            <AlertDialogTitle>删除{recordName}</AlertDialogTitle>
            <AlertDialogDescription>
              删除 {course.name} 的{recordName}。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={() =>
                notificationMutation.mutate({ courseID: course.id, level: 0 })
              }
            >
              删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
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
                value={semester ?? ALL}
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

          {syncMessage && view === "enrolled" && user && (
            <p className="text-sm text-muted-foreground">{syncMessage}</p>
          )}

          <Dialog open={syncOpen} onOpenChange={setSyncOpen}>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>同步课表</DialogTitle>
                <DialogDescription>
                  选择学期后会打开 jAccount
                  登录窗口，同步完成后自动刷新选课记录。
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-2">
                <Select value={syncSemester} onValueChange={setSyncSemester}>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="选择学期" />
                  </SelectTrigger>
                  <SelectContent>
                    {syncSemesters.map((s) => (
                      <SelectItem key={s.name} value={s.name}>
                        {s.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setSyncOpen(false)}>
                  取消
                </Button>
                <Button onClick={handleStartSync} disabled={!syncSemester}>
                  <RiRefreshLine data-icon="inline-start" />
                  同步
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {!authLoading && !user && (
            <div className="py-12 text-center">
              <p className="text-muted-foreground">登录后可以查看我的课程</p>
              <Button asChild variant="link" className="mt-2">
                <Link to="/login">登录</Link>
              </Button>
            </div>
          )}

          {user && view === "enrolled" && enrolledCourses.data && (
            <div className="flex flex-wrap items-center justify-between gap-2">
              <p className="text-sm text-muted-foreground">
                {title}共 {filteredEnrollments.length} 条
              </p>
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={handleOpenSyncDialog}
                disabled={syncSemesters.length === 0}
              >
                <RiRefreshLine data-icon="inline-start" />
                同步课表
              </Button>
            </div>
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
