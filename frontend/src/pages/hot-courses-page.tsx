import { useState } from "react"
import { PageShell } from "@/components/layout/page-shell"
import { HotCourseList } from "@/components/course/hot-course-list"
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs"

export function HotCoursesPage() {
  const [period, setPeriod] = useState<"week" | "month">("week")

  return (
    <>
      <title>热门 - JCourse</title>
      <PageShell>
        <div className="space-y-6">
          <div>
            <h1 className="text-2xl font-bold">热门</h1>
            <p className="mt-1 text-sm text-muted-foreground">课程热度排行</p>
          </div>

          <Tabs
            value={period}
            onValueChange={(v) => setPeriod(v as "week" | "month")}
          >
            <TabsList>
              <TabsTrigger value="week">本周</TabsTrigger>
              <TabsTrigger value="month">本月</TabsTrigger>
            </TabsList>
            <TabsContent value={period}>
              <HotCourseList period={period} limit={20} skeletonCount={6} />
            </TabsContent>
          </Tabs>
        </div>
      </PageShell>
    </>
  )
}
