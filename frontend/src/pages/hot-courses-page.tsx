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
  if (period === "week") return `${year} 年第 ${number} 周`
  return `${year} 年第 ${value} 月`
}

export function HotCoursesPage() {
  const [period, setPeriod] = useState<"week" | "month">("week")
  const [board, setBoard] = useState("courses")
  const { data, isLoading } = useHotCourses(period, 50)
  const periodDescription = formatHotPeriodDescription(period, data?.period_key)

  return (
    <>
      <PageTitle>热门</PageTitle>
      <PageShell>
        <div className="space-y-6">
          <div>
            <h1 className="text-2xl font-bold">热门</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              发现本周和本月大家正在讨论的课程与教师
            </p>
            <p className="mt-2 text-xs text-muted-foreground">
              热度来自点评与有用投票，反映讨论活跃度，不代表教学质量。教师热度汇总其当前任教课程的热度。
            </p>
          </div>
          <Tabs value={board} onValueChange={setBoard}>
            <TabsList>
              <TabsTrigger value="courses">课程热度榜</TabsTrigger>
              <TabsTrigger value="teachers">教师热度榜</TabsTrigger>
            </TabsList>
          </Tabs>

          <Tabs
            value={period}
            onValueChange={(v) => setPeriod(v as "week" | "month")}
          >
            <TabsList>
              <TabsTrigger value="week">本周</TabsTrigger>
              <TabsTrigger value="month">本月</TabsTrigger>
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
