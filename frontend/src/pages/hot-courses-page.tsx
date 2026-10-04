import { copy } from "@/lib/copy"
import { useState } from "react"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { HotCourseList } from "@/components/course/hot-course-list"
import { HotTeacherList } from "@/components/teacher/hot-teacher-list"
import { useHotCourses } from "@/hooks/use-course"
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs"

function formatHotPeriodDescription(
  period: "week" | "month",
  periodKey?: string
) {
  if (!periodKey) return ""
  const [year, value] = periodKey.split("-")
  const number = Number(value)
  if (!year || !Number.isFinite(number)) return ""
  if (period === "week")
    return copy("copy.user.pages.hot-courses-page.7f15ad783d80", {
      v0: year,
      v1: number,
    })
  return copy("copy.user.pages.hot-courses-page.dd3caeeb9921", {
    v0: year,
    v1: value,
  })
}

export function HotCoursesPage() {
  const [period, setPeriod] = useState<"week" | "month">("week")
  const [board, setBoard] = useState("courses")
  const { data, isLoading } = useHotCourses(period, 50)
  const periodDescription = formatHotPeriodDescription(period, data?.period_key)

  return (
    <>
      <PageTitle>
        {copy("copy.user.pages.hot-courses-page.cdc394e2e036")}
      </PageTitle>
      <PageShell>
        <div className="space-y-6">
          <div>
            <h1 className="text-2xl font-bold">
              {copy("copy.user.pages.hot-courses-page.cdc394e2e036")}
            </h1>
            <p className="mt-1 text-sm text-muted-foreground">
              {copy("copy.user.pages.hot-courses-page.eb63349bd137")}
            </p>
            <p className="mt-2 text-xs text-muted-foreground">
              {copy("copy.user.pages.hot-courses-page.07588028c5ba")}
            </p>
          </div>
          <Tabs value={board} onValueChange={setBoard}>
            <TabsList>
              <TabsTrigger value="courses">
                {copy("copy.user.pages.hot-courses-page.0f84a59a7d38")}
              </TabsTrigger>
              <TabsTrigger value="teachers">
                {copy("copy.user.pages.hot-courses-page.eb9eeb865a8a")}
              </TabsTrigger>
            </TabsList>
          </Tabs>

          <Tabs
            value={period}
            onValueChange={(v) => setPeriod(v as "week" | "month")}
          >
            <TabsList>
              <TabsTrigger value="week">
                {copy("copy.user.pages.hot-courses-page.b4c6c3eb0bce")}
              </TabsTrigger>
              <TabsTrigger value="month">
                {copy("copy.user.pages.hot-courses-page.0eeecd26f2ba")}
              </TabsTrigger>
            </TabsList>
            {periodDescription ? (
              <p className="mt-3 text-sm text-muted-foreground">
                {periodDescription}
              </p>
            ) : null}
            <TabsContent value={period}>
              {board === "teachers" ? (
                <HotTeacherList period={period} />
              ) : (
                <HotCourseList
                  period={period}
                  limit={50}
                  skeletonCount={10}
                  data={data}
                  isLoading={isLoading}
                  fetchData={false}
                />
              )}
            </TabsContent>
          </Tabs>
        </div>
      </PageShell>
    </>
  )
}
