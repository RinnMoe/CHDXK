import { Link } from "@tanstack/react-router"
import { useHotTeachers } from "@/hooks/use-teacher"
import { Skeleton } from "@/components/ui/skeleton"

export function HotTeacherList({ period }: { period: "week" | "month" }) {
  const { data, isLoading, isError } = useHotTeachers(period)
  if (isLoading)
    return (
      <div className="space-y-3">
        {Array.from({ length: 5 }, (_, i) => (
          <Skeleton key={i} className="h-16 w-full" />
        ))}
      </div>
    )
  if (isError)
    return (
      <p className="py-12 text-center text-muted-foreground">
        暂时无法加载教师榜，请稍后重试
      </p>
    )
  if (!data?.items.length)
    return (
      <p className="py-12 text-center text-muted-foreground">
        本期还没有教师上榜，课程有新的点评互动后会更新
      </p>
    )
  return (
    <div>
      <div className="flex gap-3 border-b py-2 text-sm text-muted-foreground">
        <span>排名</span>
        <span className="flex-1">教师</span>
        <span>热度</span>
      </div>
      {data.items.map((item, i) => (
        <Link
          key={item.teacher.id}
          to="/teacher/$teacherID"
          params={{ teacherID: String(item.teacher.id) }}
          className="flex items-center gap-3 border-b py-4 hover:bg-muted/50"
        >
          <span
            className={`w-7 text-center font-bold ${i < 3 ? "text-amber-500" : "text-muted-foreground"}`}
          >
            {i + 1}
          </span>
          <div className="min-w-0 flex-1">
            <p className="font-medium">{item.teacher.name}</p>
            <p className="text-sm text-muted-foreground">
              {item.teacher.department} · {item.course_count} 门课程有互动
            </p>
          </div>
          <span className="text-sm font-medium">{item.score}</span>
        </Link>
      ))}
    </div>
  )
}
